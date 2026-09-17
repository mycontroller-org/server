#!/bin/bash
set -euo pipefail

# Open a release PR. Does not push to main and does not create a tag.
#
#   1. Fetch and fast-forward local main to upstream/main
#   2. Create release-<version> from that main
#   3. Bring in all current commits and uncommitted work
#   4. Set versions.txt to NEXT_VERSION (the following devel release)
#   5. Commit and open a pull request
#
# Usage: make setup-release VERSION=x.y.z NEXT_VERSION=x.y.z

VERSION="${1:-}"
NEXT_VERSION="${2:-}"
MAIN_BRANCH="${SETUP_RELEASE_BASE:-main}"
UPSTREAM_REMOTE="${SETUP_RELEASE_UPSTREAM:-upstream}"
ORIGIN_REMOTE="${SETUP_RELEASE_ORIGIN:-origin}"

if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || [[ ! "${NEXT_VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "usage: make setup-release VERSION=x.y.z NEXT_VERSION=x.y.z" >&2
  exit 1
fi

if [ "${VERSION}" = "${NEXT_VERSION}" ]; then
  echo "NEXT_VERSION must differ from VERSION" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "not a git repository" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh is required to open the pull request" >&2
  exit 1
fi

if ! git remote get-url "${ORIGIN_REMOTE}" >/dev/null 2>&1; then
  echo "git remote '${ORIGIN_REMOTE}' is not configured" >&2
  exit 1
fi

if ! git remote get-url "${UPSTREAM_REMOTE}" >/dev/null 2>&1; then
  echo "git remote '${UPSTREAM_REMOTE}' is not configured (needed to sync ${MAIN_BRANCH})" >&2
  exit 1
fi

github_repo() {
  local url="$1"
  url="${url%.git}"
  url="${url#git@github.com:}"
  url="${url#https://github.com/}"
  url="${url#ssh://git@github.com/}"
  url="${url#git://github.com/}"
  echo "${url}"
}

SOURCE_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [ "${SOURCE_BRANCH}" = "HEAD" ]; then
  echo "detached HEAD; checkout a branch first" >&2
  exit 1
fi

FINISHED=0
STASHED=0
cleanup() {
  if [ "${FINISHED}" = "1" ]; then
    return
  fi
  current="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
  if [ -n "${current}" ] && [ "${current}" != "${SOURCE_BRANCH}" ]; then
    git checkout "${SOURCE_BRANCH}" >/dev/null 2>&1 || true
  fi
  if [ "${STASHED}" = "1" ]; then
    git stash pop >/dev/null 2>&1 || true
    STASHED=0
  fi
}
trap cleanup EXIT

RELEASE_BRANCH="release-${VERSION}"
if [ "${SOURCE_BRANCH}" = "${RELEASE_BRANCH}" ]; then
  echo "already on ${RELEASE_BRANCH}; checkout the working branch and retry" >&2
  exit 1
fi

if git show-ref --verify --quiet "refs/heads/${RELEASE_BRANCH}"; then
  echo "local branch ${RELEASE_BRANCH} already exists" >&2
  exit 1
fi

ORIGIN_URL="$(git remote get-url "${ORIGIN_REMOTE}")"
UPSTREAM_URL="$(git remote get-url "${UPSTREAM_REMOTE}")"
ORIGIN_REPO="$(github_repo "${ORIGIN_URL}")"
UPSTREAM_REPO="$(github_repo "${UPSTREAM_URL}")"
ORIGIN_OWNER="${ORIGIN_REPO%%/*}"

echo "fetching ${UPSTREAM_REMOTE} and ${ORIGIN_REMOTE}"
git fetch --prune "${UPSTREAM_REMOTE}"
git fetch --prune "${ORIGIN_REMOTE}"

if git show-ref --verify --quiet "refs/remotes/${ORIGIN_REMOTE}/${RELEASE_BRANCH}"; then
  echo "remote branch ${ORIGIN_REMOTE}/${RELEASE_BRANCH} already exists" >&2
  exit 1
fi

if ! git diff --quiet || ! git diff --cached --quiet || [ -n "$(git ls-files --others --exclude-standard)" ]; then
  echo "stashing local changes"
  git stash push --include-untracked -m "setup-release ${VERSION}"
  STASHED=1
fi

echo "syncing local ${MAIN_BRANCH} with ${UPSTREAM_REMOTE}/${MAIN_BRANCH}"
git checkout "${MAIN_BRANCH}"
if ! git merge --ff-only "${UPSTREAM_REMOTE}/${MAIN_BRANCH}"; then
  echo "local ${MAIN_BRANCH} could not fast-forward to ${UPSTREAM_REMOTE}/${MAIN_BRANCH}" >&2
  exit 1
fi

echo "creating ${RELEASE_BRANCH} from ${MAIN_BRANCH}"
git checkout -b "${RELEASE_BRANCH}"

if [ "${SOURCE_BRANCH}" != "${MAIN_BRANCH}" ]; then
  echo "bringing commits from ${SOURCE_BRANCH}"
  if ! git merge --squash "${SOURCE_BRANCH}"; then
    echo "could not apply ${SOURCE_BRANCH} onto ${MAIN_BRANCH}" >&2
    exit 1
  fi
fi

if [ "${STASHED}" = "1" ]; then
  echo "restoring stashed changes"
  if ! git stash pop; then
    STASHED=0
    FINISHED=1
    echo "stash pop had conflicts; resolve them on ${RELEASE_BRANCH}, then commit and open the PR" >&2
    exit 1
  fi
  STASHED=0
fi

cat > versions.txt <<EOF
# keep the next release version
server=${NEXT_VERSION}
EOF

git add -A
if git ls-files --error-unmatch web-console/tx >/dev/null 2>&1; then
  git rm -f --cached web-console/tx >/dev/null 2>&1 || true
fi

if git diff --cached --quiet; then
  echo "nothing to commit" >&2
  exit 1
fi

COMMIT_MSG="Release ${VERSION}

Bump versions.txt to ${NEXT_VERSION} (next devel release) and include the pending changes.

After this pull request is merged, tag v${VERSION} so CI publishes with that version."

git commit -m "${COMMIT_MSG}"

echo "pushing ${RELEASE_BRANCH} to ${ORIGIN_REMOTE}"
git push -u "${ORIGIN_REMOTE}" "${RELEASE_BRANCH}"

PR_BODY="$(cat <<EOF
## Release ${VERSION}

This pull request is for **release ${VERSION}**.

\`versions.txt\` is bumped to \`${NEXT_VERSION}\` so untagged builds report \`${NEXT_VERSION}-devel\`. The release version comes from the git tag \`v${VERSION}\` (\`scripts/version.sh\`).

### Changes
\`\`\`
$(git diff --stat "${UPSTREAM_REMOTE}/${MAIN_BRANCH}...HEAD")
\`\`\`

### After merge
On the upstream repository, tag the merge commit:

\`\`\`
git checkout ${MAIN_BRANCH}
git pull
git tag -a v${VERSION} -m v${VERSION}
git push ${UPSTREAM_REMOTE} v${VERSION}
\`\`\`

CI publishes on \`v*\` tags.
EOF
)"

PR_HEAD="${RELEASE_BRANCH}"
if [ "${ORIGIN_REPO}" != "${UPSTREAM_REPO}" ]; then
  PR_HEAD="${ORIGIN_OWNER}:${RELEASE_BRANCH}"
fi

echo "opening pull request against ${UPSTREAM_REPO}"
gh pr create \
  --repo "${UPSTREAM_REPO}" \
  --base "${MAIN_BRANCH}" \
  --head "${PR_HEAD}" \
  --title "Release ${VERSION}" \
  --body "${PR_BODY}"

echo "done: release ${VERSION} pull request is open from ${RELEASE_BRANCH}"
FINISHED=1
