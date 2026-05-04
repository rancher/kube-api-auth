package v3

import (
	"github.com/rancher/lasso/pkg/controller"
	v3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/rancher/wrangler/v3/pkg/schemes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	schemes.Register(v3.AddToScheme)
}

type Interface interface {
	ClusterAuthToken() ClusterAuthTokenController
	ClusterUserAttribute() ClusterUserAttributeController
}

func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &version{
		controllerFactory: controllerFactory,
	}
}

type version struct {
	controllerFactory controller.SharedControllerFactory
}

func (v *version) ClusterAuthToken() ClusterAuthTokenController {
	return generic.NewController[*v3.ClusterAuthToken, *v3.ClusterAuthTokenList](schema.GroupVersionKind{Group: "cluster.cattle.io", Version: "v3", Kind: "ClusterAuthToken"}, "clusterauthtokens", true, v.controllerFactory)
}

func (v *version) ClusterUserAttribute() ClusterUserAttributeController {
	return generic.NewController[*v3.ClusterUserAttribute, *v3.ClusterUserAttributeList](schema.GroupVersionKind{Group: "cluster.cattle.io", Version: "v3", Kind: "ClusterUserAttribute"}, "clusteruserattributes", true, v.controllerFactory)
}
