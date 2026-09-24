// Package iikocloud is a client for the iikoCloud (iikoTransport) API: generated
// types over hand-written services, sharing nothing with pkg/iikoserver (ADR-0003).
package iikocloud

import (
	"github.com/Nap20192/iiko-go/iikocloud/customers"
	"github.com/Nap20192/iiko-go/iikocloud/employees"
	"github.com/Nap20192/iiko-go/iikocloud/finance"
	"github.com/Nap20192/iiko-go/iikocloud/inventory"
	"github.com/Nap20192/iiko-go/iikocloud/menu"
	"github.com/Nap20192/iiko-go/iikocloud/orders"
	"github.com/Nap20192/iiko-go/iikocloud/organizations"
	"github.com/Nap20192/iiko-go/iikocloud/webhooks"

	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Config is the credential set from the developer portal plus the restaurant’s API key.
type Config = rest.Config

// Option overrides a transport default.
type Option = rest.Option

// Client is one authenticated iikoCloud session, split by domain.
type Client struct {
	Customers     *customers.Service
	Employees     *employees.Service
	Finance       *finance.Service
	Inventory     *inventory.Service
	Menu          *menu.Service
	Orders        *orders.Service
	Organizations *organizations.Service
	Webhooks      *webhooks.Service
}

// New builds a client. It performs no network call; the token is fetched lazily.
func New(cfg Config, opts ...Option) (*Client, error) {
	r, err := rest.New(cfg, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		Customers:     customers.New(r),
		Employees:     employees.New(r),
		Finance:       finance.New(r),
		Inventory:     inventory.New(r),
		Menu:          menu.New(r),
		Orders:        orders.New(r),
		Organizations: organizations.New(r),
		Webhooks:      webhooks.New(r),
	}, nil
}
