# kube-api-auth

Webhook token authenticator that lets a downstream Kubernetes cluster
verify Rancher-issued user tokens without a live connection back to
Rancher.

## Overview

Rancher issues bearer tokens to users and mirrors the material a cluster
needs to validate them into two custom resources in that cluster:

- `ClusterAuthToken` (`cluster.cattle.io/v3`) — per-token record with
  the user id, expiry, and hash reference.
- `ClusterUserAttribute` (`cluster.cattle.io/v3`) — per-user record
  with group membership and refresh state.

The token's hash lives in a `Secret` (`hash` field) alongside the
`ClusterAuthToken`. `kube-api-auth` runs as an HTTP service in the
downstream cluster and answers `TokenReview` requests from the local
`kube-apiserver`, verifying the presented bearer token against those
resources and returning the user/groups the apiserver should trust.

## How it works

1. A client sends a request to `kube-apiserver` with a bearer token
   `<accessKey>:<secretKey>`.
2. `kube-apiserver` is configured with a webhook token authenticator
   pointing at `kube-api-auth`. It forwards the token as a `TokenReview`
   to `/v1/authenticate`.
3. `kube-api-auth` looks up the `ClusterAuthToken` by `accessKey`,
   fetches the associated `Secret`, and verifies `secretKey` against the
   stored hash (SHA3-512 by default; scrypt and SHA-256 formats accepted
   for existing tokens).
4. On success it resolves the `ClusterUserAttribute` for the token's
   user, updates `LastUsedAt` (at most once per second), and — when the
   `auth-provider-refresh-debounce-seconds` ConfigMap value is set —
   marks the user for a refresh if `LastRefresh` is overdue.
5. It returns a `TokenReview` response with `authenticated: true`, the
   `username`, and the user's `groups`.

The service also migrates legacy `ClusterAuthToken.SecretKeyHash` values
into `Secret`s on first use.

## Endpoints

| Method     | Path              | Purpose                        |
|------------|-------------------|--------------------------------|
| GET, HEAD  | `/healthcheck`    | Liveness / readiness probe.    |
| POST       | `/v1/authenticate`| `TokenReview` verification.    |

Paths match exactly. `/healthcheck/` with a trailing slash returns 404. Any
other method on a known path returns 405 with an `Allow` header.

## Configuration

Everything is configured through the environment. There are no flags and no
commands; `kube-api-auth` serves. It refuses to start if given arguments it
does not understand, so a stale flag fails loudly instead of being ignored. A
bare `serve`, the command earlier versions needed, is accepted and does nothing
beyond logging a deprecation warning.

| Variable           | Default          | Description                                      |
|--------------------|------------------|--------------------------------------------------|
| `CATTLE_DEBUG`     | off              | Verbose logging. `RANCHER_DEBUG` works too.      |
| `KUBECONFIG`       | in-cluster       | Path to kubeconfig for the downstream cluster.   |
| `CATTLE_NAMESPACE` | `cattle-system`  | Namespace holding the `cluster.cattle.io/v3` CRs and hash `Secret`s. |
| `CATTLE_LISTEN`    | `127.0.0.1:6440` | Host:port to serve HTTP on.                      |

The version is logged at startup.

Runtime configuration knob:

- `ConfigMap` named `auth-provider-refresh-debounce-seconds` in
  `CATTLE_NAMESPACE`, key `value`: integer seconds. Negative or missing
  disables periodic user refresh.

## Building and running

```
make                    # runs ./scripts/ci: build, test, validate, package
make build              # binary only, into ./bin/kube-api-auth
CATTLE_LISTEN=127.0.0.1:6440 ./bin/kube-api-auth
```

Container image is `rancher/kube-api-auth`; see `package/Dockerfile`.

## Development

The `Makefile` exposes each file under `scripts/` as a target
(`$(shell ls scripts)`), so any script name is a valid `make <name>`:

```
make build                # static binary in ./bin
make test                 # go test ./...
make validate             # go mod tidy/verify, generate, fmt, then fail if the tree is dirty
make lint                 # golangci-lint run
make check-rancher-sync   # verify forked files match rancher/rancher pin
make generate             # go generate ./... (regenerates pkg/generated)
make package              # container image build
```

## Relationship to rancher/rancher

`kube-api-auth` used to import the whole `github.com/rancher/rancher`
Go module. It no longer does. The current direct dependency footprint is:

- `github.com/rancher/rancher/pkg/apis` — the CRD types
  (`ClusterAuthToken`, `ClusterUserAttribute`). This is a *separate* Go
  module (submodule of `rancher/rancher`) and is the intended public
  surface for downstream consumers.
- `github.com/rancher/wrangler/v3` — controller/factory scaffolding.
- `github.com/rancher/lasso` — controller runtime primitives.

Two source trees under `pkg/` are **forks** of Rancher-internal code:

- `pkg/auth/hashers/` — mirror of
  `rancher/rancher/pkg/auth/tokens/hashers` (SHA3/SHA-256/scrypt hash
  format shared with Rancher's token controllers).
- `pkg/clusterauth/` — mirror of the token-verification subset of
  `rancher/rancher/pkg/controllers/managementuser/clusterauthtoken/common`.

Forking these lets us drop the full `rancher/rancher` module dep, but
means we have to pull in upstream fixes (algorithm hardening, security patches).
The pinned upstream revision and per-file SHA-256 hashes live in 
[`scripts/rancher-sync.json`](scripts/rancher-sync.json).
CI runs [`scripts/check-rancher-sync`](scripts/check-rancher-sync) on
every push and fails when upstream files change since the pin.
To resolve:

```
scripts/check-rancher-sync           # shows the diff
# ...either sync local files with upstream, or bump the revision in
# scripts/rancher-sync.json, then:
scripts/check-rancher-sync update    # rewrites the pinned hashes
```

## Versioning

`kube-api-auth` releases from git tags as an OCI image consumed by
`rancher/rancher`. `main` carries the leading edge; its minor is bumped
to match the lowest Rancher minor being introduced when a breaking
change lands. Maintenance lines live on `release/v0.<MIN>` branches.

See [VERSION.md](VERSION.md) for the branch ↔ minor ↔ Rancher mapping
and the full scheme.

## Contact

For bugs, questions, corrections, suggestions: open an issue in
`rancher/rancher` with a title starting with `[kube-api-auth]`.

Or [click here](//github.com/rancher/rancher/issues/new?title=%5Bkube-api-auth%5D%20)
to create a new issue.

## License

Copyright (c) 2019-2026 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
