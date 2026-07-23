package main

import (
	"os"

	v3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	controllergen "github.com/rancher/wrangler/v3/pkg/controller-gen"
	"github.com/rancher/wrangler/v3/pkg/controller-gen/args"
)

func main() {
	_ = os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/kube-api-auth/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"cluster.cattle.io": {
				Types: []any{
					v3.ClusterAuthToken{},
					v3.ClusterUserAttribute{},
				},
			},
		},
	})
}
