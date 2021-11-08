package controllers

import (
	"context"

	"github.com/rancher/rancher/pkg/types/config"
)

func Start(ctx context.Context, apiContext *config.UserOnlyContext) error {
	return apiContext.Start(ctx)
}
