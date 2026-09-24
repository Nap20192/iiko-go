package main

import (
	"strings"
	"testing"
)

func emit(t *testing.T, spec string) string {
	t.Helper()
	out, err := Emit(mustDoc(t, spec), "gen")
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// squeeze collapses gofmt's column alignment so assertions read as source, not layout.
func squeeze(s string) string {
	f := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' })
	return strings.Join(f, " ")
}

func wantIn(t *testing.T, src string, lines ...string) {
	t.Helper()
	flat := squeeze(src)
	for _, l := range lines {
		if !strings.Contains(flat, squeeze(l)) {
			t.Errorf("missing %q in:\n%s", l, src)
		}
	}
}

func TestEmitScalarsAndOptionality(t *testing.T) {
	t.Parallel()

	src := emit(t, `{"components":{"schemas":{"A":{"type":"object",
	  "required":["id","n"],
	  "properties":{
	    "id":{"type":"string","format":"uuid"},
	    "n":{"type":"integer","format":"int32"},
	    "big":{"type":"integer","format":"int64"},
	    "price":{"type":"number","format":"double"},
	    "ok":{"type":"boolean"},
	    "when":{"type":"string","format":"date-time"}}}}}}`)

	wantIn(t, src,
		"type A struct {",
		"ID string `json:\"id\"`",
		"N int `json:\"n\"`",
		"Big int64 `json:\"big,omitempty\"`",
		"Price float64 `json:\"price,omitempty\"`",
		"Ok bool `json:\"ok,omitempty\"`",
		// iiko Cloud stamps are "yyyy-MM-dd HH:mm:ss.fff", not RFC3339.
		"When string `json:\"when,omitempty\"`",
	)
}

func TestEmitRefsArraysAndMaps(t *testing.T) {
	t.Parallel()

	src := emit(t, `{"components":{"schemas":{
	  "B":{"type":"object"},
	  "A":{"type":"object","properties":{
	    "b":{"$ref":"#/components/schemas/B"},
	    "list":{"type":"array","items":{"$ref":"#/components/schemas/B"}},
	    "tags":{"type":"array","items":{"type":"string"}},
	    "meta":{"type":"object","additionalProperties":{"type":"string"}}}}}}}`)

	wantIn(t, src,
		"B B `json:\"b,omitempty\"`",
		"List []B `json:\"list,omitempty\"`",
		"Tags []string `json:\"tags,omitempty\"`",
		"Meta map[string]string `json:\"meta,omitempty\"`",
	)
}

func TestEmitFlattensSingleAllOf(t *testing.T) {
	t.Parallel()

	// The upstream idiom: allOf:[$ref] plus a sibling description is an annotated
	// reference, while allOf:[$ref] plus own properties is inheritance.
	src := emit(t, `{"components":{"schemas":{
	  "Base":{"type":"object","required":["type"],"properties":{"type":{"type":"string"}}},
	  "A":{"type":"object","properties":{
	     "annotated":{"allOf":[{"$ref":"#/components/schemas/Base"}],"description":"x","nullable":true}}},
	  "Derived":{"allOf":[{"$ref":"#/components/schemas/Base"}],"type":"object",
	     "properties":{"extra":{"type":"string"}}}}}}`)

	wantIn(t, src,
		"Annotated Base `json:\"annotated,omitempty\"`",
		"type Derived struct {",
		"\tBase",
		"Extra string `json:\"extra,omitempty\"`",
	)
}

func TestEmitNamedStringEnum(t *testing.T) {
	t.Parallel()

	src := emit(t, `{"components":{"schemas":{
	  "OrderStatus":{"type":"string","enum":["Unconfirmed","WaitCooking"]},
	  "A":{"type":"object","properties":{"s":{"$ref":"#/components/schemas/OrderStatus"}}}}}}`)

	wantIn(t, src,
		"type OrderStatus string",
		"OrderStatusUnconfirmed OrderStatus = \"Unconfirmed\"",
		"OrderStatusWaitCooking OrderStatus = \"WaitCooking\"",
		"S OrderStatus `json:\"s,omitempty\"`",
	)
}

func TestEmitDiscriminatedFieldIsRaw(t *testing.T) {
	t.Parallel()

	// Go has no sum types; a polymorphic field stays raw so both variants
	// round-trip losslessly. Variant() builds the value on the request side.
	src := emit(t, `{"components":{"schemas":{
	  "Payment":{"type":"object","discriminator":{"propertyName":"kind","mapping":{}},
	    "properties":{"kind":{"type":"string"}}},
	  "A":{"type":"object","properties":{
	    "p":{"$ref":"#/components/schemas/Payment"},
	    "ps":{"type":"array","items":{"$ref":"#/components/schemas/Payment"}}}}}}}`)

	wantIn(t, src,
		"P json.RawMessage `json:\"p,omitempty\"`",
		"Ps []json.RawMessage `json:\"ps,omitempty\"`",
	)
}
