package controllers

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/types/config"
)

func Start(ctx context.Context, namespace string, apiContext *config.UserOnlyContext) error {
	return controller.SyncThenStart(ctx, 1,
		apiContext.Cluster.ClusterAuthTokens(namespace).Controller(),
		apiContext.Cluster.ClusterUserAttributes(namespace).Controller(),
		apiContext.Core.ConfigMaps(namespace).Controller(),
	)
}
