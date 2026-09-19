# Changelog

Notable changes to MyController are listed here. Use this file as the source for GitHub release notes.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Version numbers follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

Draft for **2.3.0** (`versions.txt`). Compared to [v2.2.0](https://github.com/mycontroller-org/server/releases/tag/v2.2.0).

### Added

- Release archives for **macOS** (`darwin/amd64` Intel and `darwin/arm64` Apple Silicon).
- **Created on** timestamp on all stored resources (gateway, node, source, field, firmware, dashboard, data repository, forward payload, schedule, task, handler, virtual device, virtual assistant, user, settings). New records get the current time. Existing records stay at the zero time (shown as empty / never). Policy and service account already had this field.
- **Service account enable**: edit form, list, details, and bulk list actions (Enable / Disable), using the same `enabled` field as other resources. Accounts that are not enabled cannot log in or keep a session.
- CLI: `myc disable service-account` / `myc enable sa` (optional `--user`).
- Topology: hierarchical **quick id** with copy (`gw1`, `gw1.node1`, `gw1.node1.source1`, `gw1.node1.source1.field1`) on the sidebar and resource details. CLI `myc get` already printed this column.
- Topology: **Download** (SVG of the current graph, as shown on screen).
- Topology: multi-select **gateway** and **node** filters (empty = all). Child-nodes applies to every selected node.
- Topology: **Alt-drag** moves a resource and its children together (left or right Alt, press or release at any time during the drag).
- Locale key `topology` in every language file.

### Changed

- Indic console locales (Hindi, Kannada, Malayalam, Tamil, Telugu): IoT resource names (gateway, node, source, field, firmware, data repository, resource, virtual device, virtual assistant, payload, forward payload, quick ID) are native-script loanwords (Tamil கேட்வே, நோடு, ஃபீல்டு), not Latin English and not calques such as நுழைவாயில். Product names and protocol abbreviations stay Latin.
- Console colour scheme is PatternFly **light**, **dark**, or **system** (follows the OS). Custom themes from the data repository (`labels.gui_theme`) are removed. The choice is stored in the browser (`localStorage` `mc-theme`). The header has a compact joined theme control and a locale dropdown (`🇬🇧 EN-UK`). Dark mode uses PatternFly tokens for labels, forms, topology, and widgets. Dashboard tiles no longer use a grey fill.
- Release workflow builds host binaries (Go 1.27.1 in Actions) and copies only those binaries into Alpine 3.24 images on the same runner (no artifact upload/download). Images are pushed to GHCR, Quay, and Docker Hub as a single multi-arch `:${VERSION}` tag (no per-arch tags). `main` republishes the `development` pre-release by deleting and recreating it.
- Policy **Import** uses `storage.Upsert` so `createdOn` / `modifiedOn` come from the backup file (same as other resources). Built-in policies are still skipped.
- Built-in policies (admin / readwrite / readonly) are no longer rewritten on every start. `ModifiedOn` stays unless the code definition changed. Existing rows without `CreatedOn` are backfilled from `ModifiedOn` once.
- `make setup-release` now requires both versions: `make setup-release VERSION=x.y.z NEXT_VERSION=x.y.z`. The PR writes `NEXT_VERSION` into `versions.txt`.
- Topology Force layout: dragging one node no longer bounces the others. Layout waits for a real surface size and fits after the simulation settles. Force no longer crashes when a node is mid-attach (`getParent` on a source with no parent).
- Topology node/gateway filters use a checkbox select with ticks; the closed field shows selected names, not “All … + count”.
- Topology sidebar: long field values truncate with a tooltip; timestamps stay visible. Divider above labels and the chart matches the other rules in the panel. Esc closes the sidebar; click outside does not.
- Sleeping queue tab skips `/api/gateway-sleeping-queue` when the gateway is disabled or the node has `sleep_queue_disabled` (avoids the 2s timeout).
- Successful **login** clears notification drawer and toasters. Failed `/user/login` is not added as an interceptor alert. Page refresh does not wipe other alerts.

### Fixed

- Source `Save` looks up `createdOn` by storage id, so renaming gateway/node/source id does not reset the timestamp.
- Settings live `update()` path preserves `createdOn`.
- Select multi-select used by Task / Schedule / User / Policy forms still uses typeahead chips; checkbox style is only for Topology filters.

### Upgrade notes

- No data migration is required for `createdOn`. Existing resources keep a zero timestamp.
- User and service account `disabled` becomes `enabled` (upgrade `2.3.0-1`). Existing accounts stay usable.

---

## [2.2.0] - 2026-09-17

See the [v2.2.0 GitHub release](https://github.com/mycontroller-org/server/releases/tag/v2.2.0) for the full notes (access control, CLI aliases, in-tree web console, topology, locales, MySensors FOTA, HTTPS persist).
