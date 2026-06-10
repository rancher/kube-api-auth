# kube-api-auth

A microservice for user authentication in kubernetes clusters.

## Building

`make`

## Running

`./bin/kube-api-auth`

## Versioning

`kube-api-auth` is released directly from git tags as an OCI image consumed by `rancher/rancher`. The `master` branch carries the leading edge; its minor version is bumped to match the lowest Rancher minor being introduced when a breaking change lands. Maintenance lines for older Rancher versions live on `release/v0.<MIN>` branches (where `<MIN>` is the lowest Rancher minor served by that line).

See [VERSION.md](VERSION.md) for the current branch ↔ minor ↔ Rancher mapping and the full scheme.

## Contact

For bugs, questions, comments, corrections, suggestions, etc., open an issue in rancher/rancher with a title starting with `[kube-api-auth]`.

Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Bkube-api-auth%5D%20) to create a new issue.

## License

Copyright (c) 2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
