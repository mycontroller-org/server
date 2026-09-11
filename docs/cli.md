# MyController CLI (`myc`)

This document describes the **MyController command-line client**: how to build it, log in, list and change resources, and apply gateways, nodes, sources, and fields from YAML or JSON files.

---

## 1. Overview

`myc` talks to a running MyController **server** over its HTTP API. It does not start the server.

| Command | Purpose |
| --- | --- |
| `login` / `logout` | Store or clear server credentials |
| `version` | Print client and server version |
| `get` | List resources |
| `apply` | Add, merge, or delete resources from a YAML or JSON file |
| `upload` | Upload a firmware binary to an existing firmware resource |
| `set` | Update a stored property, or set a live field value |
| `delete` | Delete resources by id |
| `enable` / `disable` | Enable or disable resources |
| `reload` | Reload a gateway or virtual assistant |

### Build

```bash
make client
# binary: builds/myc
```

Or:

```bash
go build -trimpath -o builds/myc ./cmd/client
```

---

## 2. Configuration

After a successful login, `myc` writes `$HOME/.mycontroller.yaml` (override with `--config`).

Environment variables use the prefix `MYC_` (Viper automatic env). Example: `MYC_URL`.

The stored password field is the session token, encoded as `BASE64/...`.

### Global flags

These flags apply to every command:

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `$HOME/.mycontroller.yaml` | Client config file |
| `-o`, `--output` | `console` | Output format: `console`, `wide`, `yaml`, `json` |
| `--hide-header` | `false` | Hide table headers on console output |
| `--pretty` | `false` | Pretty-print JSON |

`wide` is the same as `console` plus extra columns marked as wide (for example quick id).

---

## 3. Login and logout

```bash
# username and password
myc login http://localhost:8080 --username admin --password password

# prompt for username and password
myc login http://localhost:8080

# prompt for password only
myc login http://localhost:8080 --username admin

# service token
myc login http://localhost:8080 --token <token>

# TLS without certificate verification
myc login https://localhost:8443 --username admin --password password --insecure
```

| Flag | Default | Description |
| --- | --- | --- |
| `-u`, `--username` | | Login username |
| `-p`, `--password` | | Login password |
| `-t`, `--token` | | Service token (skips username/password) |
| `--expires-in` | `720h` | Session lifetime |
| `--insecure` | `false` | Skip TLS certificate verification |

```bash
myc logout
```

```bash
myc version
```

Prints client build information. If logged in, also queries the server version.

---

## 4. Get

List resources from the server.

```bash
myc get gateway
myc get node
myc get source --limit 50 --sort-by name --sort-order desc
myc get field --filter "gateway id=mysensor" --filter "node id==1"
myc get gateway -o yaml
myc get node -o json --pretty
myc get field -o wide
```

### Persistent flags

| Flag | Default | Description |
| --- | --- | --- |
| `--limit` | `10` | Maximum rows |
| `--sort-by` | `id` | Sort key (header title or value path) |
| `--sort-order` | `asc` | `asc` or `desc` |
| `--filter` | | Repeatable `key=value` filter |

### Filter operators

The first matching operator in the filter string is used:

| Syntax | Operator |
| --- | --- |
| `==` | equal |
| `!=` | not equal |
| `>=` | greater than or equal |
| `<=` | less than or equal |
| `>` | greater than |
| `<` | less than |
| `=` | regex (case insensitive) |

The key is matched against the table header title (spaces ignored, case insensitive). If the header has a value path, that path is used (for example `gateway id` → `gatewayId`).

### Resources

| Command | Aliases |
| --- | --- |
| `get gateway` | `gw`, `gateways` |
| `get node` | `nodes` |
| `get source` | `sources` |
| `get field` | `fields` |
| `get firmware` | `firmwares`, `fw` |
| `get data-repository` | `data-repositories`, `data-repo` |
| `get virtual-device` | `virtual-devices`, `vd` |
| `get virtual-assistant` | `virtual-assistants`, `va` |
| `get task` | `tasks` |
| `get schedule` | `schedules` |
| `get handler` | `handlers` |
| `get forward-payload` | `forward-payloads` |
| `get backup` | `backups` |

---

## 5. Apply

`myc apply` creates, merges, or deletes **gateways**, **nodes**, **sources**, **fields**, **firmware**, and **data repositories** from a YAML or JSON file.

Firmware **binaries** are not part of apply. Create the firmware resource with apply, then upload the file with `myc upload firmware`.

```bash
myc apply -f resources.yaml
myc apply -f resources.yaml --dry-run
myc apply -f resources.yaml --replace
myc apply -f nodes.yaml -f sources.yaml
myc apply -f - --dry-run < resources.json
```

| Flag | Description |
| --- | --- |
| `-f`, `--filename` | YAML or JSON file. Repeat for multiple files. `-` reads stdin. Required. |
| `--dry-run` | Check the server and print the table without writing changes |
| `--replace` | If an `add` target already exists, delete it and recreate it with the **same id** |

The command prints one table and exits `1` if any row failed. There is no extra summary line after the table.

```text
RESOURCE                           ACTION  STATUS
gateway: mysensor                  add     ok
node: mysensor.1                   add     ok
source: mysensor.1.dht             add     replaced
field: mysensor.1.dht.temperature  add     failed: already exists
```

### 5.1 Action and status

**ACTION** is the operation from the file (`add`, `merge`, `delete`), not the internal recreate step.

| STATUS | Meaning |
| --- | --- |
| `ok` | add, merge, or delete succeeded |
| `replaced` | the resource already existed and was deleted then recreated with the same id |
| `failed: …` | the operation did not run, with a short reason |
| `not available` | delete target was not found; remaining resources still run |
| `dry-run` | `--dry-run` and the operation would have succeeded |

`--replace` or `replace: true` in the file only affects **add**. It does not change ACTION to `replace`.

### 5.2 File shapes

The file may be:

- a single object
- a YAML stream of documents separated by `---`
- a YAML or JSON array of objects
- an object with a shared header and an `items` list

JSON is detected when the file starts with `{` or `[`.

### 5.3 Single resource

```yaml
kind: gateway
operation: add
id: mysensor
description: MySensors USB gateway
enabled: true
```

```yaml
kind: node
operation: add
gatewayId: mysensor
nodeId: "1"
name: Living Room
labels:
  location: living-room
```

```yaml
kind: source
operation: merge
gatewayId: mysensor
nodeId: "1"
sourceId: dht
name: DHT Sensor
```

```yaml
kind: field
operation: delete
gatewayId: mysensor
nodeId: "1"
sourceId: dht
fieldId: temperature
```

`kind` aliases: `gateway` / `gw` / `gateways`, `node` / `nodes`, `source` / `sources`, `field` / `fields`.

`operation` aliases: `add` / `create`, `update`, `delete` / `remove`.

### 5.4 Items list

`kind`, `operation`, and `replace` on the document apply to every item. Other keys on the document are defaults merged into each item (item keys win).

```yaml
kind: field
operation: add
replace: true
gatewayId: mysensor
nodeId: "1"
sourceId: dht
items:
  - fieldId: temperature
    name: Temperature
    metricType: gauge
    unit: °C
  - fieldId: humidity
    name: Humidity
    metricType: gauge
    unit: "%"
```

If an item includes `fieldId`, it is applied as a **field** even when the document `kind` is `source` or `node`.

```yaml
kind: source
operation: add
replace: true
items:
  - gatewayId: mysensor
    nodeId: "1"
    sourceId: dht
    fieldId: temperature
    name: Temperature
    metricType: gauge
    unit: °C
```

That item is a field, not a source.

### 5.5 Identity and required fields

| Kind | Identity | Required for add/update | Required for delete |
| --- | --- | --- | --- |
| gateway | `id` | `id` | `id` |
| firmware | `id` | `id` | `id` |
| data-repository | `id` | `id` | `id` |
| node | `gatewayId` + `nodeId` | `gatewayId`, `nodeId` | `id` or `gatewayId`+`nodeId` |
| source | `gatewayId` + `nodeId` + `sourceId` | those three | `id` or those three |
| field | `gatewayId` + `nodeId` + `sourceId` + `fieldId` | those four | `id` or those four |

Lookup uses `id` when it is set, otherwise the natural keys.

Gateway, firmware, and data-repository HTTP APIs require an `id` on save; supply it in the file. For a new node or source without `id`, the client generates a UUID. A new field may omit `id`; the server assigns one.

Apply of firmware writes **metadata only** (`id`, `description`, `labels`). The binary stays empty until `myc upload firmware`. Updating firmware metadata keeps the existing file. Replacing a firmware deletes the old file; upload again after replace.

### 5.6 Operations

**add**

- Resource missing: create it. Status `ok`.
- Resource present: fail with `failed: already exists`, unless `--replace` or `replace: true`.
- With replace: delete the existing resource and create the new one using the **deleted resource’s id**, so references keep working. Status `replaced`.

**merge** (`update` is accepted as an alias)

- Resource present: deep-merge the file onto the live resource and save. The existing storage id is always kept. Status `ok`.
- Resource missing: `failed: not found`.

Merge rules:

- Keys not in the file stay as they are.
- Nested maps (`labels`, `others`, `provider`, `data`, and any other object) are merged key by key.
- Arrays of objects are merged by a key on each item, tried in this order: `id`, `key`, `fieldId`, `field`, `name`, `type`, `sourceId`. Matching items are deep-merged; new items are appended; items only on the server stay.
- Arrays of scalars (strings, numbers) are replaced by the file value.

**delete**

- Resource present: delete it. Status `ok`.
- Resource missing: `not available`. This is not a failure; later resources still run.

### 5.7 Parent checks

Add, merge, and replace require the parent to exist:

| Resource | Parent |
| --- | --- |
| gateway | none |
| firmware | none |
| data-repository | none |
| node | gateway (`gatewayId`) |
| source | node (`gatewayId` + `nodeId`) |
| field | source (`gatewayId` + `nodeId` + `sourceId`) |

A parent added (or replaced/merged) **earlier in the same apply** counts as present. A parent deleted earlier in the same apply counts as missing. If a parent write fails during apply, later children in the same file are not saved.

```yaml
kind: gateway
operation: add
id: mysensor
enabled: true
---
kind: node
operation: add
gatewayId: mysensor
nodeId: "1"
name: Living Room
```

If the gateway is missing and is not added earlier in the file:

```text
RESOURCE          ACTION  STATUS
node: mysensor.1  add     failed: parent gateway mysensor is not present
```

Delete does not require a parent.

### 5.8 Resource fields

Any field on the server type can be set in the file. Common ones:

**Gateway** (`kind: gateway`)

| Field | Description |
| --- | --- |
| `id` | Gateway id (required) |
| `description` | Text |
| `enabled` | `true` / `false` |
| `reconnectDelay` | Duration string, for example `15s` |
| `queueFailedMessage` | `true` / `false` |
| `provider` | Provider map (`type`, `protocol`, …) |
| `messageLogger` | Logger map |
| `labels` | String map |
| `others` | Free-form map |

**Node**

| Field | Description |
| --- | --- |
| `id` | Storage id (optional on add) |
| `gatewayId`, `nodeId` | Natural keys |
| `name` | Display name |
| `labels`, `others` | Maps |
| `state` | Status object |

**Source**

| Field | Description |
| --- | --- |
| `id` | Storage id (optional on add) |
| `gatewayId`, `nodeId`, `sourceId` | Natural keys |
| `name` | Display name |
| `labels`, `others` | Maps |

**Firmware** (`kind: firmware`)

| Field | Description |
| --- | --- |
| `id` | Firmware id (required) |
| `description` | Text |
| `labels` | String map (for example `platform`, `ms_flash_slot`) |

Do not put a binary path in apply. Use `myc upload firmware`.

**Data repository** (`kind: data-repository`)

Aliases: `datarepository`, `data-repo`, `data-repositories`.

| Field | Description |
| --- | --- |
| `id` | Repository id (required) |
| `description` | Text |
| `readOnly` | `true` / `false` |
| `labels` | String map |
| `data` | Free-form map (scripts and policy values) |

**Field**

| Field | Description |
| --- | --- |
| `id` | Storage id (optional on add) |
| `gatewayId`, `nodeId`, `sourceId`, `fieldId` | Natural keys |
| `name` | Display name |
| `metricType` | For example `gauge` |
| `unit` | For example `°C` |
| `formatter` | Payload formatter (`onReceive`) |
| `labels`, `others` | Maps |

Numeric ids in YAML (for example `nodeId: 1`) are accepted and stored as strings.

### 5.9 Full example

```yaml
kind: gateway
operation: add
id: mysensor
description: MySensors USB gateway
enabled: true
---
kind: node
operation: add
gatewayId: mysensor
nodeId: "1"
name: Living Room
labels:
  location: living-room
---
kind: source
operation: add
gatewayId: mysensor
nodeId: "1"
sourceId: dht
name: DHT Sensor
---
kind: field
operation: add
gatewayId: mysensor
nodeId: "1"
sourceId: dht
fieldId: temperature
name: Temperature
metricType: gauge
unit: °C
---
kind: field
operation: add
replace: true
gatewayId: mysensor
nodeId: "1"
sourceId: dht
items:
  - fieldId: humidity
    name: Humidity
    metricType: gauge
    unit: "%"
```

```yaml
kind: firmware
operation: add
id: stm32-app-slot-a
description: STM32 slot A image
labels:
  ms_flash_slot: A
---
kind: data-repository
operation: add
id: ota_stm32_ab
description: STM32 A/B OTA policy
data:
  disabled: false
```

Verify first:

```bash
myc apply -f resources.yaml --dry-run
```

Then apply. Use `--replace` when you want existing `add` targets recreated instead of failing.

---

## 6. Upload firmware

Upload a binary to an **existing** firmware resource. Apply the firmware metadata first.

```bash
myc apply -f firmware.yaml
myc upload firmware stm32-app-slot-a ./app-slot-a.signed.bin
myc upload fw stm32-app-slot-a ./app-slot-a.signed.bin
```

| Argument | Description |
| --- | --- |
| `<id>` | Firmware resource id |
| `<file>` | Path to the binary on disk |

On success the client prints the stored file details:

```text
uploaded
  firmware        stm32-app-slot-a
  name            firmware.signed.bin
  internal name   stm32-app-slot-a.bin
  size            67.83 KiB
  checksum        sha256:b1f3c09e4acbbf8ca40e6197bad871757c950cac4329b63aff8b171fdbdb3fe6
  modified        a minute ago
```

The server updates `file.name`, `file.size`, `file.checksum` (sha256), and `file.modifiedOn`.

If the firmware id is not present:

```text
firmware stm32-app-slot-a is not present
```

---

## 7. Set

`set <kind>` updates a **stored property** on a resource (scripts, description, labels, and other fields). Live sensor/actuator values use a separate command: `set value field`.

### Nested property (scripts and other text)

```bash
myc set <kind> <id> <key-path> <value>
myc set <kind> <id> <key-path> --file script.js
myc set <kind> <id> --path <key-path> --file script.js
```

| Kind | Aliases | Id |
| --- | --- | --- |
| `gateway` | `gw`, `gateways` | gateway id |
| `node` | `nodes` | storage id or `gatewayId.nodeId` |
| `source` | `sources` | storage id or `gatewayId.nodeId.sourceId` |
| `field` | `fields` | storage id or `gatewayId.nodeId.sourceId.fieldId` |
| `firmware` | `firmwares`, `fw` | firmware id |
| `data-repository` | `data-repo`, `data-repositories` | repository id |

The key path uses dots and matches JSON field names:

| Example | What it updates |
| --- | --- |
| `formatter.onReceive` | Field receive script |
| `data.onConfig` | Data-repository script |
| `data.onBlock` | Data-repository script |
| `description` | Description text |
| `labels.ms_flash_slot` | A single label |

`--file` always stores the file contents as raw text (useful for JavaScript). Without `--file`, the last argument is the value. Inline values that are valid JSON (`true`, `false`, numbers, objects, arrays) are stored as that type; other inline text is stored as a string.

```bash
myc set field mysensor.1.dht.temperature formatter.onReceive --file on_receive.js
myc set field mysensor.1.dht.temperature formatter.onReceive "return value;"
myc set data-repository ota_stm32_ab data.onConfig --file onConfig.js
myc set data-repo ota_stm32_ab --path data.onBlock --file onBlock.js
myc set gateway mysensor description "USB gateway"
myc set node mysensor.1 others.note --file note.txt
myc set firmware stm32-app-slot-a labels.ms_flash_slot A
```

Several ids can be given; they all receive the same path and value:

```bash
myc set field id-1 id-2 formatter.onReceive --file on_receive.js
```

### Live field value

Use `set value field`. This sends an action; it does not change stored metadata.

```bash
myc set value field gw1.1.1.V_CUSTOM 23.5
myc set value field mysensor.1.dht.temperature 21.0
```

Do not use `myc set field` for this. `set field` always updates a stored key path (`formatter.onReceive`, `name`, `unit`, …).

---

## 8. Delete, enable, disable, reload

These commands take one or more **storage ids** (the `id` column from `get`, not the quick id), except as noted.

### Delete

```bash
myc delete gateway <id> [<id>...]
myc delete node <id>
myc delete source <id>
myc delete field <id>
```

| Resource | Aliases |
| --- | --- |
| `gateway` | `gw`, `gateways` |
| `node` | `nodes` |
| `source` | `sources` |
| `field` | `fields` |
| `firmware` | `firmwares`, `fw` |
| `data-repository` | `data-repositories`, `data-repo` |
| `virtual-device` | `virtual-devices`, `vd` |
| `virtual-assistant` | `virtual-assistants`, `va` |
| `task` | `tasks` |
| `schedule` | `schedules` |
| `handler` | `handlers` |
| `forward-payload` | `forward-payloads` |
| `backup` | `backups` |

### Enable / disable

```bash
myc enable gateway <id>
myc disable task <id>
```

Supported: `gateway`, `virtual-device`, `virtual-assistant`, `task`, `schedule`, `handler` (same aliases as `get`).

### Reload

```bash
myc reload gateway <id>
myc reload virtual-assistant <id>
```

Supported: `gateway`, `virtual-assistant`.

---

## 9. Quick ids

Several commands and the UI refer to resources by **quick id**:

| Kind | Format | Example |
| --- | --- | --- |
| gateway | `{gatewayId}` | `mysensor` |
| node | `{gatewayId}.{nodeId}` | `mysensor.1` |
| source | `{gatewayId}.{nodeId}.{sourceId}` | `mysensor.1.dht` |
| field | `{gatewayId}.{nodeId}.{sourceId}.{fieldId}` | `mysensor.1.dht.temperature` |

`get` shows a `quick id` column in `-o wide`. `apply` prints the same shape after the kind in the RESOURCE column (`node: mysensor.1`).

`delete`, `enable`, `disable`, and `reload` use the storage **id** (UUID or configured gateway id), not the dotted quick id.

---

## 10. Exit status

| Situation | Exit code |
| --- | --- |
| All apply rows succeeded, or delete-not-available only | `0` |
| Any apply row `failed` | `1` |
| Missing file, parse error, or other command error | `1` |
