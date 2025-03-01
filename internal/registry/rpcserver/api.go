package rpcserver

import (
	"context"
	apiv1_registry "vc/internal/gen/registry/apiv1.registry"
	apiv1_status "vc/internal/gen/status/apiv1.status"
)

// Apiv1 interface
type Apiv1 interface {
	Add(ctx context.Context, req *apiv1_registry.AddRequest) (*apiv1_registry.AddReply, error)
	Revoke(ctx context.Context, req *apiv1_registry.RevokeRequest) (*apiv1_registry.RevokeReply, error)
	Validate(ctx context.Context, req *apiv1_registry.ValidateRequest) (*apiv1_registry.ValidateReply, error)

	Status(ctx context.Context, req *apiv1_status.StatusRequest) (*apiv1_status.StatusReply, error)
}
