module github.com/rancher/kube-api-auth

go 1.16

replace (
	github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible // oras dep requires a replace is set

	github.com/docker/docker => github.com/docker/docker v20.10.6+incompatible // oras dep requires a replace is set
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e

	github.com/rancher/lasso => github.com/rancher/lasso v0.0.0-20220110205840-98715bdd6b5b
	github.com/rancher/norman => github.com/rancher/norman v0.0.0-20220107203912-4feb41eafabd
	github.com/rancher/rancher => github.com/rancher/rancher v0.0.0-20220225023242-635286172d41
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20220225023242-635286172d41
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20220225023242-635286172d41

	helm.sh/helm/v3 => github.com/rancher/helm/v3 v3.5.4-rancher.1

	k8s.io/api => k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.3
	k8s.io/apiserver => k8s.io/apiserver v0.22.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.3
	k8s.io/client-go => github.com/rancher/client-go v1.22.3-rancher.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.3
	k8s.io/code-generator => k8s.io/code-generator v0.22.3
	k8s.io/component-base => k8s.io/component-base v0.22.3
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.3
	k8s.io/controller-manager => k8s.io/controller-manager v0.22.3
	k8s.io/cri-api => k8s.io/cri-api v0.22.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.3
	k8s.io/kubectl => k8s.io/kubectl v0.22.3
	k8s.io/kubelet => k8s.io/kubelet v0.22.3
	k8s.io/kubernetes => k8s.io/kubernetes v1.22.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.3
	k8s.io/metrics => k8s.io/metrics v0.22.3
	k8s.io/mount-utils => k8s.io/mount-utils v0.22.3
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.3
)

require (
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/bshuster-repo/logrus-logstash-hook v1.0.2 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/rancher/lasso v0.0.0-20220222230204-fc50f3dd670d // indirect
	github.com/rancher/norman v0.0.0-20220107203912-4feb41eafabd
	github.com/rancher/rancher v0.0.0-20220225023242-635286172d41
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.2
	rsc.io/letsencrypt v0.0.3 // indirect
)
