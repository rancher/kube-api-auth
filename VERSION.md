# Versioning

`kube-api-auth` ships as an OCI image consumed by `rancher/rancher` and is
released directly from git tags.

The `main` branch carries the leading edge. The main minor is bumped to
match the lowest Rancher minor the next release will serve — for example,
main moves to `v0.13` when it starts tracking Rancher v2.13+. Tags on
`main` are sequential patches within the minor: `v0.13.0`, `v0.13.1`, …

Maintenance lines for Rancher versions no longer on main live on
`release/v0.<MIN>` branches (where `<MIN>` is the lowest Rancher minor
served by that line), cut from the initial `v0.<MIN>.0` tag. Tags on
maintenance branches are sequential: `v0.<MIN>.0`, `v0.<MIN>.1`, …

## Supported release lines

Keep the table well-formed: one row per branch, columns in the order shown.

| Branch              | Minor   | Rancher Versions     | Status |
|---------------------|---------|----------------------|--------|
| main                | v0.13   | v2.13, v2.14, v2.15  | active |
| release/v0.10       | v0.10   | v2.10, v2.11, v2.12  | active |

Status values: `active` (receives fixes), `eol` (no longer maintained). EOL
rows stay in the table — they keep the historical mapping discoverable and
let image-scanning tooling correlate old tags to the Rancher lines that
shipped them.

## Currently deployed tags

Which kube-api-auth tag is consumed by the latest patch of each supported
Rancher minor. Update this when a Rancher line bumps its
`pkg/apis/management.cattle.io/v3/tools_system_images.go` pin. Superseded
tags drop out (rancher/rancher git history of `tools_system_images.go` is
authoritative for per-patch precision).

| kube-api-auth Tag | Rancher Minors      |
|-------------------|---------------------|
| v0.2.6            | v2.13, v2.14, v2.15 |
| v0.2.4            | v2.10, v2.11, v2.12 |
