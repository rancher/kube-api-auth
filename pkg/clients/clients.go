package clients

import (
	"context"
	"fmt"

	cluster "github.com/rancher/kube-api-auth/pkg/generated/controllers/cluster.cattle.io"
	clusterv3 "github.com/rancher/kube-api-auth/pkg/generated/controllers/cluster.cattle.io/v3"
	clusterv3api "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/rancher/wrangler/v3/pkg/schemes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
)

func init() {
	_ = schemes.Register(corev1.AddToScheme)
	_ = schemes.Register(clusterv3api.AddToScheme)
}

type Clients struct {
	ClusterAuthTokens     clusterv3.ClusterAuthTokenClient
	ClusterAuthTokenCache clusterv3.ClusterAuthTokenCache

	ClusterUserAttributes     clusterv3.ClusterUserAttributeClient
	ClusterUserAttributeCache clusterv3.ClusterUserAttributeCache

	Secrets      corev1client.SecretInterface
	SecretLister corev1listers.SecretNamespaceLister

	ConfigMapLister corev1listers.ConfigMapNamespaceLister

	clusterFactory *cluster.Factory
	coreFactory    informers.SharedInformerFactory
}

func New(ctx context.Context, cfg *rest.Config, namespace string) (*Clients, error) {
	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes clientset: %w", err)
	}

	clusterFactory, err := cluster.NewFactoryFromConfigWithOptions(cfg, &generic.FactoryOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("creating cluster controller factory: %w", err)
	}

	coreFactory := informers.NewSharedInformerFactoryWithOptions(k8s, 0,
		informers.WithNamespace(namespace),
	)

	clusterAPI := clusterFactory.Cluster().V3()

	return &Clients{
		ClusterAuthTokens:     clusterAPI.ClusterAuthToken(),
		ClusterAuthTokenCache: clusterAPI.ClusterAuthToken().Cache(),

		ClusterUserAttributes:     clusterAPI.ClusterUserAttribute(),
		ClusterUserAttributeCache: clusterAPI.ClusterUserAttribute().Cache(),

		Secrets:      k8s.CoreV1().Secrets(namespace),
		SecretLister: coreFactory.Core().V1().Secrets().Lister().Secrets(namespace),

		ConfigMapLister: coreFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(namespace),

		clusterFactory: clusterFactory,
		coreFactory:    coreFactory,
	}, nil
}

func (c *Clients) Start(ctx context.Context) error {
	c.coreFactory.Start(ctx.Done())
	c.coreFactory.WaitForCacheSync(ctx.Done())

	if err := c.clusterFactory.Start(ctx, 5); err != nil {
		return fmt.Errorf("starting cluster controllers: %w", err)
	}

	return nil
}
