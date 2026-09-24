// Package webhooks reads and updates an organization’s webhook subscription settings.
package webhooks

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the webhooks half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// GetSettings get webhooks settings for specified organization and authorized API login.
func (s *Service) GetSettings(ctx context.Context, req gen.GetWebHookSettingsRequest) (*gen.GetWebHookSettingsResponse, error) {
	return rest.Call[gen.GetWebHookSettingsResponse](ctx, s.c, getSettings, req)
}

// UpdateSettings update webhooks settings for specified organization and authorized API login.
func (s *Service) UpdateSettings(ctx context.Context, req gen.UpdateWebHookSettingsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, updateSettings, req)
}
