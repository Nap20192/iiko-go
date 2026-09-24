// Package iikoserver is a client for the iikoServer back-office API (/resto/api).
//
// One Client holds one session, and a session is a licence seat rather than a
// rate-limit slot, so Close must always run.
package iikoserver

import (
	"context"

	"github.com/Nap20192/iiko-go/iikoserver/cashshifts"
	"github.com/Nap20192/iiko-go/iikoserver/corporation"
	"github.com/Nap20192/iiko-go/iikoserver/documents"
	"github.com/Nap20192/iiko-go/iikoserver/edi"
	"github.com/Nap20192/iiko-go/iikoserver/events"
	"github.com/Nap20192/iiko-go/iikoserver/nomenclature"
	"github.com/Nap20192/iiko-go/iikoserver/pricing"
	"github.com/Nap20192/iiko-go/iikoserver/recipes"
	"github.com/Nap20192/iiko-go/iikoserver/reports"
	"github.com/Nap20192/iiko-go/iikoserver/rest"
	"github.com/Nap20192/iiko-go/iikoserver/staff"
	"github.com/Nap20192/iiko-go/iikoserver/suppliers"
)

// Config is the transport's configuration, re-exported so callers need not
// import rest to construct a Client.
type Config = rest.Config

// Client is the facade over the domains. Every domain shares one transport, so
// one token, one mutex and one logout cover the whole surface.
type Client struct {
	rest *rest.Client

	Corporation  *corporation.Service
	Nomenclature *nomenclature.Service
	Recipes      *recipes.Service
	Reports      *reports.Service
	Cashshifts   *cashshifts.Service
	Documents    *documents.Service
	EDI          *edi.Service
	Events       *events.Service
	Pricing      *pricing.Service
	Staff        *staff.Service
	Suppliers    *suppliers.Service
}

// New builds the facade. It performs no I/O: the session is opened lazily on the
// first call, so constructing a Client does not yet occupy a licence seat.
func New(cfg Config) *Client {
	c := rest.New(cfg)
	return &Client{
		rest:         c,
		Corporation:  corporation.New(c),
		Nomenclature: nomenclature.New(c),
		Recipes:      recipes.New(c),
		Reports:      reports.New(c),
		Cashshifts:   cashshifts.New(c),
		Documents:    documents.New(c),
		EDI:          edi.New(c),
		Events:       events.New(c),
		Pricing:      pricing.New(c),
		Staff:        staff.New(c),
		Suppliers:    suppliers.New(c),
	}
}

// Close releases the licence seat; a leaked token holds it until the idle timeout.
func (c *Client) Close(ctx context.Context) error { return c.rest.Close(ctx) }

// AllowWrite reports whether this process may touch production data.
func (c *Client) AllowWrite() bool { return c.rest.AllowWrite() }

// LicenceInfo reports free slots for a licence module.
func (c *Client) LicenceInfo(ctx context.Context, moduleID string) (string, error) {
	return c.rest.LicenceInfo(ctx, moduleID)
}
