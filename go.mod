module github.com/rancher/kube-api-auth

go 1.26.0

toolchain go1.26.4

replace (
	github.com/docker/distribution => github.com/docker/distribution v2.8.2+incompatible
	github.com/rancher/rancher => github.com/rancher/rancher v0.0.0-20260606011257-0932e0f2e111
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20260606011257-0932e0f2e111
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20260606011257-0932e0f2e111
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.44.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.44.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.44.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.44.0
	go.opentelemetry.io/proto/otlp => go.opentelemetry.io/proto/otlp v1.8.0
	helm.sh/helm/v3 => github.com/rancher/helm/v3 v3.19.0-rancher1

	k8s.io/api => k8s.io/api v0.36.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.36.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.36.1
	k8s.io/apiserver => k8s.io/apiserver v0.36.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.36.1
	k8s.io/client-go => k8s.io/client-go v0.36.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.36.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.36.1
	k8s.io/code-generator => k8s.io/code-generator v0.36.1
	k8s.io/component-base => k8s.io/component-base v0.36.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.36.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.36.1
	k8s.io/cri-api => k8s.io/cri-api v0.36.1
	k8s.io/cri-client => k8s.io/cri-client v0.36.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.36.1
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.36.1
	k8s.io/endpointslice => k8s.io/endpointslice v0.36.1
	k8s.io/externaljwt => k8s.io/externaljwt v0.36.1
	k8s.io/kms => k8s.io/kms v0.36.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.36.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.36.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.36.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.36.1
	k8s.io/kubectl => k8s.io/kubectl v0.36.1
	k8s.io/kubelet => k8s.io/kubelet v0.36.1
	k8s.io/kubernetes => k8s.io/kubernetes v1.36.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.36.1
	k8s.io/metrics => k8s.io/metrics v0.36.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.36.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.36.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.36.1
)

require (
	github.com/gorilla/mux v1.8.1
	github.com/rancher/norman v0.9.7
	github.com/rancher/rancher v0.0.0-00010101000000-000000000000
	github.com/rancher/rancher/pkg/apis v0.0.0
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli v1.22.16
	k8s.io/api v0.36.1
	k8s.io/apimachinery v0.36.1
	k8s.io/client-go v12.0.0+incompatible
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bshuster-repo/logrus-logstash-hook v1.0.2 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/containerd/containerd v1.7.30 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/swag v0.25.4 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.4 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/fileutils v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/mangling v0.25.4 // indirect
	github.com/go-openapi/swag/netutils v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kubereboot/kured v1.13.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/matryer/moq v0.6.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/rancher/aks-operator v1.15.0-rc.1 // indirect
	github.com/rancher/ali-operator v1.15.0-rc.1 // indirect
	github.com/rancher/apiserver v0.9.7 // indirect
	github.com/rancher/eks-operator v1.15.0-rc.1 // indirect
	github.com/rancher/fleet/pkg/apis v0.16.0-alpha.9 // indirect
	github.com/rancher/gke-operator v1.15.0-rc.1 // indirect
	github.com/rancher/lasso v0.2.9 // indirect
	github.com/rancher/remotedialer v0.6.1 // indirect
	github.com/rancher/steve v0.9.14 // indirect
	github.com/rancher/system-upgrade-controller/pkg/apis v0.0.0-20260519183600-f1362a3fe1a8 // indirect
	github.com/rancher/wrangler/v3 v3.7.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rubenv/sql-migrate v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260401024825-9d38bb4040a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260511170946-3700d4141b60 // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	helm.sh/helm/v3 v3.20.2 // indirect
	k8s.io/apiextensions-apiserver v0.36.1 // indirect
	k8s.io/apiserver v0.36.1 // indirect
	k8s.io/cli-runtime v0.36.1 // indirect
	k8s.io/component-base v0.36.1 // indirect
	k8s.io/component-helpers v0.36.1 // indirect
	k8s.io/controller-manager v0.0.0 // indirect
	k8s.io/helm v2.17.0+incompatible // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-aggregator v0.36.1 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/kubectl v0.36.1 // indirect
	k8s.io/kubernetes v1.36.1 // indirect
	k8s.io/utils v0.0.0-20260319190234-28399d86e0b5 // indirect
	oras.land/oras-go/v2 v2.6.0 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.34.0 // indirect
	sigs.k8s.io/cli-utils v0.37.2 // indirect
	sigs.k8s.io/cluster-api v1.13.2 // indirect
	sigs.k8s.io/controller-runtime v0.24.1 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
