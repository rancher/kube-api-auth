# Contributing

We are not accepting code contributions at the moment. A clear issue is the
most useful thing you can send.

Open issues in [rancher/rancher](https://github.com/rancher/rancher/issues)
with a title that starts with `[kube-api-auth] `.

## Report a bug

Use the [bug report template](https://github.com/rancher/rancher/issues/new?template=bug_report.md&title=%5Bkube-api-auth%5D%20)
and include:

- the kube-api-auth image:
  `kubectl -n cattle-system get daemonset kube-api-auth -o jsonpath='{.spec.template.spec.containers[0].image}'`
- the Rancher version, and whether it is community or Prime
- the downstream cluster's distribution (RKE2, K3s or another) and
  Kubernetes version
- the exact command you ran against the cluster's authorized cluster
  endpoint, and its full output
- the kube-api-auth logs from the time of the failure:
  `kubectl -n cattle-system logs daemonset/kube-api-auth`
- what you expected to happen
- steps that reproduce it

Remove tokens, passwords and server URLs that must stay private from
everything you paste.

## Request a feature

Use the [feature request template](https://github.com/rancher/rancher/issues/new?template=feature_request.md&title=%5Bkube-api-auth%5D%20).
Describe the workflow you are trying to do and how you do it today, not
only the change you want.

## Report a security issue

Never in a public issue: see the
[Rancher security policy](https://github.com/rancher/rancher/security/policy).

## Ask a question

Use the [Rancher forums](https://forums.rancher.com/) or
[Rancher Slack](https://slack.rancher.io/).
