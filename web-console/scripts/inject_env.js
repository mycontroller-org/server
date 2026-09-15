// reference https://stackoverflow.com/a/62879437
const childProcess = require("child_process")
const fs = require("fs")
const path = require("path")
const moment = require("moment")

const versionsFile = path.resolve(__dirname, "../../versions.txt")

function git(cmd) {
  try {
    return childProcess.execSync(cmd, { encoding: "utf8" }).trim()
  } catch (_err) {
    return ""
  }
}

// Same rules as scripts/version.sh so the packed UI matches the server binary.
function backendVersion(gitBranch) {
  const fromEnv = process.env.VERSION
  if (fromEnv) {
    return fromEnv
  }
  const tagged = (gitBranch || "").match(/([0-9]+\.[0-9]+\.[0-9]+)$/)
  if (tagged) {
    return tagged[1]
  }
  let staticVersion = "0.0.0"
  try {
    const line = fs
      .readFileSync(versionsFile, "utf8")
      .split("\n")
      .find((l) => l.startsWith("server="))
    if (line) {
      staticVersion = line.split("=")[1].trim()
    }
  } catch (_err) {
    // versions.txt missing; keep fallback
  }
  return `${staticVersion}-devel`
}

let gitBranch = process.env.GIT_BRANCH || git("git rev-parse --abbrev-ref HEAD")
if (gitBranch === "HEAD") {
  gitBranch = git("git describe --abbrev=0 --tags")
}

const env = {
  REACT_APP_IS_DEV_ENV: process.env.REACT_APP_IS_DEV_ENV || "",
  REACT_APP_VERSION: backendVersion(gitBranch),
  REACT_APP_GIT_BRANCH: gitBranch,
  REACT_APP_GIT_SHA: process.env.GIT_SHA || git("git rev-parse HEAD"),
  REACT_APP_GIT_SHA_SHORT: process.env.GIT_SHA_SHORT || git("git rev-parse --short HEAD"),
  REACT_APP_BUILD_DATE: process.env.BUILD_DATE || moment().format(),
}

const body = Object.keys(env)
  .map((key) => `${key}='${String(env[key]).trim()}'\n`)
  .join("")

fs.writeFileSync(path.resolve(__dirname, "../.env"), body)
