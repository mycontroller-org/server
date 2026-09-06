# Access control (policies)

This document describes MyController’s **policy-based access control**: how identities, policies, resources, and service tokens work, and how to configure them with examples.

---

## 1. Overview

MyController authorizes HTTP API calls with:

1. **Authentication** – valid JWT (login or service token).
2. **User state** – user must exist and must not be **disabled**.
3. **Service token** (if used) – must exist, belong to the user, and not be expired.
4. **Authorization** – at least one attached **policy** must **Allow** the requested **action** on the requested **resource**.

There is no multi-tenant isolation in this model. Policies define _what_ a principal may do on _which_ named resources.

### Naming

This feature is **policy-based access control**: reusable **policies** are attached to **users** (and optionally narrowed on service tokens). It is not classical role-based access control (User → Role → permissions).

| Term              | Meaning                                            |
| ----------------- | -------------------------------------------------- |
| **Policy**        | Reusable permission document (`id` + `statements`) |
| **User.policies** | List of policy ids attached to the user            |
| **Statement**     | An allow rule: `actions` + `resources`             |

### Design summary

| Concept             | Role                                                                             |
| ------------------- | -------------------------------------------------------------------------------- |
| **User**            | Identity; holds password, `disabled`, and a list of **policy IDs**               |
| **Policy**          | Named document: list of **statements** (effect, actions, resources)              |
| **Service token**   | Always tied to a user; optional **extra limits** that can only **reduce** access |
| **Resource string** | `kind` or `kind:name` (name may use hierarchical wildcards)                      |
| **Action**          | Verb such as `get`, `list`, `update`, `delete`, …                                |

Effective access for a service token:

```text
effective = permissions(user policies)  ∩  token restrictions (if any)
```

If the token has no restrictions, it has the same rights as the user.

---

## 2. Concepts

### 2.1 Users

Relevant fields:

| Field      | Description                                                          |
| ---------- | -------------------------------------------------------------------- |
| `id`       | Storage identifier                                                   |
| `username` | Login name                                                           |
| `disabled` | If `true`, login and all API calls with that user’s JWT fail (`401`) |
| `policies` | List of policy IDs attached to this user                             |

Manage users:

- API: `/api/user`
- Web UI: **Settings → Users**

### 2.2 Policies

A policy is a reusable permission document.

```yaml
id: plant-room-viewer
description: Read-only access to the plant room device tree
system: false
statements:
  - effect: Allow
    actions:
      - get
      - list
    resources:
      - gateway:plant-room
      - node:plant-room.*
      - source:plant-room.*
      - field:plant-room.*
      - metric:plant-room.*
      - dashboard
      - quickid
      - status
```

| Field        | Description                                                             |
| ------------ | ----------------------------------------------------------------------- |
| `id`         | Stable name used when attaching the policy to users                     |
| `system`     | Built-in policies (`admin`, `readwrite`, `readonly`); cannot be deleted |
| `statements` | List of allow rules (see below)                                         |

Manage policies:

- API: `/api/policy`
- Web UI: **Settings → Policies**

### 2.3 Statements

Each statement has:

| Field       | Description                              |
| ----------- | ---------------------------------------- |
| `effect`    | `Allow` or `Deny`                        |
| `actions`   | List of verbs, or `*` for all            |
| `resources` | List of resource strings, or `*` for all |

Evaluation (across **all** statements on **all** attached policies):

1. If any matching statement is **`Deny`** → **denied** (Deny wins).
2. Else if any matching statement is **`Allow`** → **allowed**.
3. Else → **denied** (default deny).

Empty `effect` is treated as `Allow`.

**Device-tree cascade** (gateway → node → source → field/metric):

Both **Allow** and **Deny** cascade **down** the path. **Deny always wins.**

| Statement                          | Effect                                              |
| ---------------------------------- | --------------------------------------------------- |
| Allow `gateway:X`                  | Also allows node/source/field/metric under `X`      |
| Allow `node:X.Y` or `node:X.*`     | Also allows source/field/metric under that path     |
| Allow `source:X.Y.Z`               | Also allows field/metric under that source          |
| Deny `node:X.Y`                    | Blocks that node **and** its sources/fields/metrics |
| Deny `node:X.Y` + Allow `node:X.*` | Sibling nodes (and their fields) stay allowed       |

So a compact policy is enough:

```yaml
statements:
  - effect: Allow
    actions: [get, list]
    resources:
      - gateway:mysensor
      - node:mysensor.*
  - effect: Deny
    actions: [get, list]
    resources:
      - node:mysensor.1
```

You do **not** need separate `source:` / `field:` lines unless you want a narrower
scope (e.g. only one source). Kind-wide Deny (`node`, `node:*`, `*`) still blocks
the whole kind (and descendants).

### 2.4 Service tokens

Service tokens always have a `userId`. They act **as that user**, with optional tightening:

| Field                       | Description                                                   |
| --------------------------- | ------------------------------------------------------------- |
| `userId`                    | Owning user (immutable after create)                          |
| `neverExpire` / `expiresOn` | Lifetime of the token                                         |
| `actions`                   | Optional: only these actions (subset of the user’s)           |
| `resources`                 | Optional: only these resource patterns (subset of the user’s) |

Empty `actions` and `resources` mean “no extra limit” (same as the user).

Evaluation:

```text
1. User policies must allow (action, resource)
2. If token has restrictions, they must also allow (action, resource)
```

A token **cannot** grant more than the user has.

---

## 3. Actions

| Action    | Typical HTTP use                                               |
| --------- | -------------------------------------------------------------- |
| `get`     | `GET /api/{kind}/{id}`, metrics query, quickid, status details |
| `list`    | `GET /api/{kind}` (collection)                                 |
| `create`  | Create-style endpoints (when mapped)                           |
| `update`  | `POST` create-or-update body                                   |
| `delete`  | `DELETE`                                                       |
| `enable`  | `.../enable`                                                   |
| `disable` | `.../disable`                                                  |
| `reload`  | `.../reload`                                                   |
| `action`  | `/api/action`, node/gateway actions                            |
| `*`       | All actions                                                    |

---

## 4. Resource kinds (complete list)

Resource strings look like:

```text
kind
kind:name
kind:name-with.dots.and.*
*
```

All kinds recognized by the authorization engine are listed below.  
**RO/RW** = included in built-in `readonly` / `readwrite`.  
**Admin** = only via `admin` (`*`) or an explicit custom policy.

### 4.0 Full inventory

| Kind               | Primary APIs                  | Name for fine-grained rules                                         | In `readonly` / `readwrite`                |
| ------------------ | ----------------------------- | ------------------------------------------------------------------- | ------------------------------------------ |
| `gateway`          | `/api/gateway`                | `gatewayId`                                                         | Yes                                        |
| `node`             | `/api/node`                   | `gatewayId.nodeId`                                                  | Yes                                        |
| `source`           | `/api/source`                 | `gatewayId.nodeId.sourceId`                                         | Yes                                        |
| `field`            | `/api/field`                  | `gatewayId.nodeId.sourceId.fieldId`                                 | Yes                                        |
| `task`             | `/api/task`                   | task `id`                                                           | Yes                                        |
| `schedule`         | `/api/schedule`               | schedule `id`                                                       | Yes                                        |
| `handler`          | `/api/handler`                | handler `id`                                                        | Yes                                        |
| `dashboard`        | `/api/dashboard`              | dashboard `id` (often UUID)                                         | Yes                                        |
| `firmware`         | `/api/firmware`               | firmware `id`                                                       | Yes                                        |
| `forwardpayload`   | `/api/forwardpayload`         | id                                                                  | Yes                                        |
| `datarepository`   | `/api/datarepository`         | id                                                                  | Yes                                        |
| `virtualdevice`    | `/api/virtualdevice`          | id                                                                  | Yes                                        |
| `virtualassistant` | `/api/virtualassistant`       | id                                                                  | Yes                                        |
| `servicetoken`     | `/api/servicetoken`           | entity id                                                           | Yes                                        |
| `metric`           | `/api/metric`                 | same hierarchy as **field** path                                    | Yes (kind-level in built-ins; see metrics) |
| `action`           | `/api/action`                 | optional target name                                                | Yes                                        |
| `status`           | `/api/server/status`          | (kind only)                                                         | Yes                                        |
| `quickid`          | `/api/quickid`                | API entry; each `?id=` is checked as the target kind (field/node/…) | Yes                                        |
| `user`             | `/api/user`                   | user id when applicable                                             | **No** (admin / custom)                    |
| `policy`           | `/api/policy`                 | policy id when applicable                                           | **No** (admin / custom)                    |
| `settings`         | `/api/settings`               | (kind only)                                                         | **No** (admin / custom)                    |
| `backup`           | `/api/backup`, `/api/restore` | (kind only)                                                         | **No** (admin / custom)                    |
| `*`                | all of the above              | everything                                                          | `admin` only (as `*`)                      |

Related endpoints that are **not** separate policy kinds:

| Endpoint                          | Behavior                                                          |
| --------------------------------- | ----------------------------------------------------------------- |
| `GET /api/status`                 | Public minimal status; no policy required                         |
| `GET /api/version`                | Not gated as a policy kind in the same way as server status       |
| `GET/POST /api/user/login`, OAuth | Public auth entry                                                 |
| `GET/POST /api/user/profile`      | Own profile; allowed for the logged-in user without `user` rights |
| `/api/gateway-sleeping-queue`     | Treated as **gateway** access                                     |
| `/api/firmware/upload/...`        | Treated as **firmware** update                                    |

If a path maps to a segment that is not in the table, the engine still uses that segment as `kind` (normalized for multi-word APIs such as `forwardpayload`). Prefer the kinds above in policies.

### 4.1 Device tree (hierarchical names)

Device entities use **protocol / configuration ids**, not only storage UUIDs, for policy names:

| Kind      | Name format                                 | Example                                          |
| --------- | ------------------------------------------- | ------------------------------------------------ |
| `gateway` | `{gatewayId}`                               | `gateway:plant-room`                             |
| `node`    | `{gatewayId}.{nodeId}`                      | `node:plant-room.sensor-01`                      |
| `source`  | `{gatewayId}.{nodeId}.{sourceId}`           | `source:plant-room.sensor-01.climate`            |
| `field`   | `{gatewayId}.{nodeId}.{sourceId}.{fieldId}` | `field:plant-room.sensor-01.climate.temperature` |

Storage may still use a UUID as primary key for node/source/field. For **get by UUID**, the server loads the entity and checks the **business name** above.

### 4.2 Other entities (usually by `id`)

| Kind               | Name              | Notes                              |
| ------------------ | ----------------- | ---------------------------------- |
| `task`             | task `id`         | User-chosen or generated id        |
| `schedule`         | schedule `id`     |                                    |
| `handler`          | handler `id`      |                                    |
| `dashboard`        | dashboard `id`    | Often a **UUID** created by the UI |
| `firmware`         | firmware `id`     |                                    |
| `forwardpayload`   | id                | API path `/api/forwardpayload`     |
| `datarepository`   | id                | API path `/api/datarepository`     |
| `virtualdevice`    | id                |                                    |
| `virtualassistant` | id                |                                    |
| `servicetoken`     | entity id         |                                    |
| `user`             | user management   | Create/list/update/delete users    |
| `policy`           | policy management | Create/list/update/delete policies |
| `settings`         | system settings   |                                    |
| `backup`           | backup / restore  |                                    |

### 4.3 API capabilities (not devices)

| Kind      | API                  | Purpose                                                                                                                                                   |
| --------- | -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `metric`  | `/api/metric`        | Time-series; hierarchical name like **field** path                                                                                                        |
| `quickid` | `/api/quickid`       | Resolve quick IDs for widgets; **each `id` is authorized as that resource** (e.g. `field:gw.n.s.f`), so Deny on a node also blocks that field via quickid |
| `status`  | `/api/server/status` | Detailed server status (authenticated)                                                                                                                    |
| `action`  | `/api/action`        | Generic resource actions; **each target is authorized as its own resource**, so `action` alone controls nothing you cannot already write                   |
| `*`       | everything           | Full access                                                                                                                                               |

Note: `GET /api/status` (minimal public status) does not require these policies.

`settings` is split per document, so a read-only console grant does not expose credentials:

| Resource                            | API                                     | Notes                                     |
| ----------------------------------- | --------------------------------------- | ----------------------------------------- |
| `settings:system_settings`          | `GET /api/settings/system`              | What the console needs; in `readonly`/`readwrite` |
| `settings:system_backup_locations`  | `GET /api/settings/backuplocations`     | May hold storage credentials              |
| `settings:system_dynamic_secrets`   | `GET /api/settings/system/jwtsecret/reset` | Resets the JWT secret: signs out everyone |
| `settings`                          | `POST /api/settings`                    | Write; the body names the document        |

### 4.4 Wildcards

| Pattern                                          | Matches                                   |
| ------------------------------------------------ | ----------------------------------------- |
| `*`                                              | All kinds and names                       |
| `field` or `field:*`                             | All fields                                |
| `field:plant-room.*`                             | All fields under gateway `plant-room`     |
| `field:plant-room.sensor-01.*`                   | All fields under that node                |
| `field:plant-room.sensor-01.climate.*`           | All fields under that source              |
| `field:plant-room.sensor-01.climate.temperature` | Exactly one field                         |
| `node:plant-room.*`                              | All nodes under gateway `plant-room`      |
| `metric:plant-room.*`                            | Metrics for all fields under that gateway |

Trailing `.*` means “this segment and all deeper segments” for hierarchical names.

---

## 5. Built-in policies

Created automatically on startup / upgrade:

| Policy ID   | Access                                                                                                                 |
| ----------- | ---------------------------------------------------------------------------------------------------------------------- |
| `admin`     | `actions: [*]`, `resources: [*]`                                                                                       |
| `readwrite` | Read/write on operational kinds (devices, tasks, dashboards, metric, …). **No** `user`, `policy`, `settings`, `backup` |
| `readonly`  | `get` / `list` on the same operational kinds                                                                           |

Existing users with an empty `policies` list receive **`admin`** during upgrade so installs are not locked out.

Default user on fresh install: `admin` / `admin` with policy `admin`.

---

## 6. How enforcement works

### 6.1 Request path

```text
HTTP request
  → JWT valid?
  → user active (not disabled)?
  → service token valid (if present)?
  → map path + method → action + resource
  → (optional) resolve UUID → business name
  → Allowed(user policies ∩ token limits)?          ← layer 1: may you reach this endpoint?
  → authorize each target named in the body?        ← layer 2: may you write *this* object?
  → handler / storage
```

Most checks run in HTTP middleware. List queries also **inject storage filters** so only allowed rows are returned from the database (not a full load then filter in memory).

### 6.1a Two layers, and why

A resource string without a name (`gateway`, not `gateway:gw1`) means **the collection**, and a
collection check is deliberately permissive: a grant on `gateway:gw1` lets you *reach*
`GET /api/gateway` and `POST /api/gateway`, because the row filter or the object check decides
what you may actually touch.

That matters because several endpoints name their targets in the **body**, not the path:

| Request                                   | Body                | Object level check         |
| ----------------------------------------- | ------------------- | -------------------------- |
| `POST /api/gateway`                       | `{"id":"gw1",...}`  | `update` on `gateway:gw1`  |
| `POST /api/gateway/enable`                | `["gw1","gw2"]`     | `enable` on each gateway   |
| `DELETE /api/gateway`                     | `["gw1"]`           | `delete` on `gateway:gw1`  |
| `POST /api/field`                         | `{"gatewayId":...}` | `update` on `field:<path>` |
| `POST /api/action`                        | `[{"resource":…}]`  | `action` on each quick id  |
| `GET /api/action/node?id=a&id=b`          | –                   | `action` on **every** id   |
| `GET /api/quickid?id=…`                   | –                   | `get` on each quick id     |
| `GET`/`POST /api/metric`                  | quick id / tags     | `get` on the target field  |

A payload that names **no** target — creating an object whose id the server generates — requires a
**kind-wide** grant (`*`, `gateway`, or `gateway:*`). A grant on one named object is never enough to
create new ones, which is what keeps `user:<own-id>` from being a path to `admin`.

### 6.1b Service tokens are personal

`/api/servicetoken` is always scoped to the caller, whatever the policies say. Tokens act as their
owner, so no principal can read, widen (drop the `actions`/`resources` limits, set `neverExpire`) or
delete another principal's tokens. To revoke someone else's access, disable the user.

### 6.2 List queries

Example client request:

```http
GET /api/gateway?filter=[{"k":"id","o":"in","v":["plant-room","workshop"]}]
```

Combined with policy `gateway:plant-room` only:

```text
client filter  AND  policy scope
→ only plant-room (intersection)
```

Multiple policy resource names become an **OR** of groups at query level, then **AND**ed with the client filter.

### 6.3 Get by UUID

For `GET /api/node/{uuid}`:

1. Load node by UUID.
2. Build name `gatewayId.nodeId`.
3. Check `get` on `node:gatewayId.nodeId`.

Policies should use business names (or wildcards), not node UUIDs, for device tree resources.

### 6.4 Metrics

Metrics are **not** covered by `field:…` alone.

| Resource   | Controls                                    |
| ---------- | ------------------------------------------- |
| `field:…`  | List/get field configuration and values API |
| `metric:…` | `/api/metric` for that hierarchical path    |

The UI often POSTs metrics with `tags.id = <field UUID>`. The server resolves that UUID to the field path and checks `metric:<path>`.

Examples:

```yaml
# All metrics under one gateway
- metric:plant-room.*

# One node
- metric:plant-room.sensor-01.*

# One field
- metric:plant-room.sensor-01.climate.temperature

# All metrics (any device)
- metric
# or
- metric:*
```

### 6.5 Dashboards

Dashboard **id** is often a **UUID** generated by the web UI. Policies match that id:

```yaml
# All dashboards
- dashboard
# or
- dashboard:*

# One dashboard (use the real id from GET /api/dashboard)
- dashboard:3f2a9c1e-8b4d-4e2f-9a1b-0c7d6e5f4a3b
```

The dashboard **title** is not used for authorization.

Opening a dashboard only loads layout. Widgets still need `field`, `metric`, `quickid`, etc., as appropriate.

### 6.6 Own profile

`/api/user/profile` is allowed for the logged-in user without requiring the `user` resource (so people can change their own password/profile).

Managing other users requires the `user` resource (typically `admin` or a custom identity policy).

---

## 7. Management APIs and UI

| Resource       | API prefix          | UI                        |
| -------------- | ------------------- | ------------------------- |
| Users          | `/api/user`         | Settings → Users          |
| Policies       | `/api/policy`       | Settings → Policies       |
| Service tokens | `/api/servicetoken` | Settings → Service Tokens |

Only principals with policy rights on `user` / `policy` can manage them (e.g. built-in `admin`).

---

### 8.0 Allow all operational, but deny settings and backup

```yaml
id: operator-no-settings
statements:
  - effect: Allow
    actions: ["*"]
    resources: ["*"]
  - effect: Deny
    actions: ["*"]
    resources:
      - settings
      - backup
      - user
      - policy
```

### 8.0b Allow all gateways except one

```yaml
statements:
  - effect: Allow
    actions: [get, list]
    resources: [gateway:*]
  - effect: Deny
    actions: ["*"]
    resources: [gateway:secret-gw]
```

List queries apply Deny as well:

- **Id-keyed kinds** (`gateway`, `task`, …): exact Deny ids use `NotIn` on `id`.
- **Hierarchical kinds** (`node`, `source`, `field`): e.g. Deny `node:mysensor.1` excludes that node from list via `NOT (gatewayId=mysensor AND nodeId=1)`.

Deny statements must include the **`list`** action (or `*`) to affect list results; a Deny with only `get` still blocks get-by-id but not list.

## 8. Complete examples

Example device layout used below:

```text
Gateway id:   plant-room
  Node id:    sensor-01
    Source:   climate
      Fields: temperature, humidity
  Node id:    pump-01
    Source:   motor
      Fields: running, runtime
Gateway id:   workshop
  ...
```

### 8.1 Full administrator

Use built-in policy:

```yaml
# user.policies
policies: [admin]
```

### 8.2 Operator (all devices, no user/policy admin)

```yaml
policies: [readwrite]
```

### 8.3 Global read-only

```yaml
policies: [readonly]
```

### 8.4 Single gateway, read-only, with charts and one dashboard

```yaml
id: plant-room-viewer
description: View plant-room devices, metrics, and a shared dashboard
system: false
statements:
  - effect: Allow
    actions: [get, list]
    resources:
      - gateway:plant-room
      - node:plant-room.*
      - source:plant-room.*
      - field:plant-room.*
      - metric:plant-room.*
      - dashboard:3f2a9c1e-8b4d-4e2f-9a1b-0c7d6e5f4a3b
      - quickid
      - status
```

Attach to a user:

```yaml
username: alice
policies: [plant-room-viewer]
disabled: false
```

### 8.5 Control only the pump under plant-room

```yaml
id: pump-operator
description: Control pump node only; read climate sensors
system: false
statements:
  - effect: Allow
    actions: [get, list]
    resources:
      - gateway:plant-room
      - node:plant-room.sensor-01
      - source:plant-room.sensor-01.*
      - field:plant-room.sensor-01.*
      - metric:plant-room.sensor-01.*
      - quickid
  - effect: Allow
    actions: [get, list, update, action, enable, disable]
    resources:
      - node:plant-room.pump-01
      - source:plant-room.pump-01.*
      - field:plant-room.pump-01.*
      - metric:plant-room.pump-01.*
      - action
```

### 8.6 Metrics for one field only

```yaml
id: temp-chart-only
statements:
  - effect: Allow
    actions: [get, list]
    resources:
      - field:plant-room.sensor-01.climate.temperature
      - metric:plant-room.sensor-01.climate.temperature
      - quickid
```

Without `metric:…`, field access alone does **not** open `/api/metric`.

### 8.7 Identity administrator (users and policies only)

```yaml
id: identity-admin
statements:
  - effect: Allow
    actions: ["*"]
    resources:
      - user
      - policy
```

Combine with another policy if that person also needs device access.

### 8.8 Service token narrower than the user

User has `readwrite`. Token for automation:

```yaml
name: plant-room-metrics-bot
userId: <alice-user-id>
neverExpire: false
expiresOn: "2027-12-31"
actions:
  - get
  - list
resources:
  - field:plant-room.*
  - metric:plant-room.*
  - quickid
```

The bot cannot update devices or touch other gateways, even though Alice could.

### 8.9 Disable a user

```yaml
username: bob
disabled: true
policies: [readonly]
```

Existing JWTs for Bob are rejected on the next request (user cache updated on save).

---

## 9. Client filters and policies together

Clients may pass their own filters (example):

```http
GET /api/field?limit=20&offset=0&filter=[{"k":"gatewayId","o":"eq","v":"plant-room"}]
```

Policy scope always applies as well. The result is the **intersection**:

```text
rows matching client filters  ∩  rows allowed by policy/token
```

Asking for a gateway outside the policy yields an **empty list**, not a way to expand access.  
Denied kinds or verbs yield **403 Forbidden**.

---

## 10. Migration and bootstrap

| Situation                          | Behavior                                                                  |
| ---------------------------------- | ------------------------------------------------------------------------- |
| Fresh install                      | Built-in policies created; default `admin` user gets policy `admin`       |
| Upgrade from before access control | Patch `2.2.0-1` creates policies; users with empty `policies` get `admin` |
| HTTP startup                       | Built-in policies ensured (statements reset); users with no policies are **logged**, not granted anything |

An empty `policies` list means **no access**, and it stays that way. Only the one-time
`2.2.0-1` upgrade grants `admin` to policy-less users, because on a pre-RBAC install every
existing user really was an administrator. Stripping a user's policies is therefore a valid way
to lock them out, and a restart will not undo it.

Built-in policies (`admin`, `readwrite`, `readonly`) are code: their statements are rewritten on
every start, and the API rejects edits to them. Copy one into a new policy to customise it.

### Deleting policies

A policy that is still attached to a user cannot be deleted — the API returns the affected
usernames. Detach it first. This prevents a dangling policy id, and prevents a user's list from
silently becoming empty (which would remove all of their access).

---

## 11. Performance notes

- Users, policies, and service tokens used for auth are kept in an **in-memory cache**, refreshed on write.
- List scoping is applied as **storage query filters** (including OR of name patterns).
- Get-by-UUID for device entities does one lookup to resolve the business name before the policy check.

---

## 12. Quick reference

### Attach policy to user (conceptually)

```yaml
# Policy
id: plant-room-viewer
statements: [...]

# User
username: alice
policies: [plant-room-viewer]
```

### Minimal console-friendly viewer for one gateway

```yaml
actions: [get, list]
resources:
  - gateway:plant-room
  - node:plant-room.*
  - source:plant-room.*
  - field:plant-room.*
  - metric:plant-room.*
  - dashboard # all dashboards; or dashboard:<uuid>
  - quickid
  - status
```

### What not to confuse

| Do not assume               | Reality                            |
| --------------------------- | ---------------------------------- |
| `field:…` implies metrics   | Need `metric:…` as well            |
| Dashboard title in policy   | Use dashboard **id** (often UUID)  |
| Node UUID in policy         | Use `gatewayId.nodeId`             |
| Token can elevate rights    | Token can only **narrow** the user |
| Disabled user keeps working | JWT rejected while disabled        |

---

## 13. Related code (for developers)

| Area                                            | Location                                      |
| ----------------------------------------------- | --------------------------------------------- |
| Policy types                                    | `pkg/types/policy`                            |
| Engine, match, list/query filters, metrics auth | `pkg/api/policy`                              |
| Middleware authz                                | `pkg/http_router/middleware/auth.go`          |
| User / policy HTTP routes                       | `pkg/http_router/routes/user.go`, `policy.go` |
| Upgrade seed                                    | `pkg/upgrade/v2_2_0__1.go`                    |
| Web UI                                          | Settings → Users, Settings → Policies         |

---

## 14. Changelog (feature introduction)

Policy-based access control was introduced for server release line **2.2.0** (upgrade id `2.2.0-1`): built-in policies, user `policies` / `disabled`, service token restrictions, and enforcement on the HTTP API.
