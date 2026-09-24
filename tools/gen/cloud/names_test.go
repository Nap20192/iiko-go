package main

import "testing"

func TestEmitDisambiguatesCollidingNames(t *testing.T) {
	t.Parallel()

	src := emit(t, `{"components":{"schemas":{
	  "ns.Request.CreateOrder.Address":{"type":"object","properties":{"a":{"type":"string"}}},
	  "ns.Response.Order.Address":{"type":"object","properties":{"b":{"type":"string"}}}}}}`)

	wantIn(t, src, "type CreateOrderAddress struct {", "type OrderAddress struct {")
}
