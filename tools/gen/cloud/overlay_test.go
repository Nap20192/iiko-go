package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustDoc(t *testing.T, s string) map[string]any {
	t.Helper()
	var d map[string]any
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestApplySetRequiresMatchingGuard(t *testing.T) {
	t.Parallel()

	doc := mustDoc(t, `{"components":{"schemas":{"A":{"type":"bool"}}}}`)
	// sha256 of the canonical JSON `{"type":"bool"}`.
	const guard = "3d9e4e6e5e8b1e2c2e0c2ca7d7f4a2dbf5e6b6a7c8d9e0f1a2b3c4d5e6f70819"

	o := Overlay{Name: "t", Actions: []Action{{
		Op: "set", Target: "/components/schemas/A/type", Expect: guard, Value: "boolean",
	}}}
	if err := o.Apply(doc); err == nil || !strings.Contains(err.Error(), "guard") {
		t.Fatalf("stale guard must fail, got %v", err)
	}
}

func TestApplySetWithCorrectGuard(t *testing.T) {
	t.Parallel()

	doc := mustDoc(t, `{"components":{"schemas":{"A":{"type":"bool"}}}}`)
	o := Overlay{Name: "t", Actions: []Action{{
		Op: "set", Target: "/components/schemas/A/type", Expect: Digest("bool"), Value: "boolean",
	}}}
	if err := o.Apply(doc); err != nil {
		t.Fatal(err)
	}
	got := doc["components"].(map[string]any)["schemas"].(map[string]any)["A"].(map[string]any)["type"]
	if got != "boolean" {
		t.Fatalf("got %v, want boolean", got)
	}
}

func TestApplyMergeKeepsSiblings(t *testing.T) {
	t.Parallel()

	doc := mustDoc(t, `{"p":{"type":"float","description":"keep me"}}`)
	cur := map[string]any{"type": "float", "description": "keep me"}
	o := Overlay{Name: "t", Actions: []Action{{
		Op: "merge", Target: "/p", Expect: Digest(cur),
		Value: map[string]any{"type": "number", "format": "double"},
	}}}
	if err := o.Apply(doc); err != nil {
		t.Fatal(err)
	}
	p := doc["p"].(map[string]any)
	if p["type"] != "number" || p["format"] != "double" || p["description"] != "keep me" {
		t.Fatalf("got %v", p)
	}
}

func TestRemoveParameterCountGuard(t *testing.T) {
	t.Parallel()

	const doc = `{"paths":{"/a":{"post":{"parameters":[
		{"name":"Authorization","in":"header"},{"name":"Timeout","in":"header"}]}}}}`

	if err := (Overlay{Name: "t", Actions: []Action{{
		Op: "removeParameter", Match: map[string]any{"name": "Authorization", "in": "header"}, Count: 2,
	}}}).Apply(mustDoc(t, doc)); err == nil || !strings.Contains(err.Error(), "guard") {
		t.Fatalf("wrong count must fail, got %v", err)
	}

	d := mustDoc(t, doc)
	if err := (Overlay{Name: "t", Actions: []Action{{
		Op: "removeParameter", Match: map[string]any{"name": "Authorization", "in": "header"}, Count: 1,
	}}}).Apply(d); err != nil {
		t.Fatal(err)
	}
	ps := d["paths"].(map[string]any)["/a"].(map[string]any)["post"].(map[string]any)["parameters"].([]any)
	if len(ps) != 1 || ps[0].(map[string]any)["name"] != "Timeout" {
		t.Fatalf("got %v", ps)
	}
}

func TestRenameSchemaRewritesRefs(t *testing.T) {
	t.Parallel()

	d := mustDoc(t, `{"components":{"schemas":{
		"Wrap`+"`"+`1[[X]]":{"type":"object"},
		"B":{"properties":{"w":{"$ref":"#/components/schemas/Wrap`+"`"+`1[[X]]"}}}}}}`)
	o := Overlay{Name: "t", Actions: []Action{{
		Op: "renameSchema", Target: "Wrap`1[[X]]", Value: "WrapX", Expect: Digest(map[string]any{"type": "object"}),
	}}}
	if err := o.Apply(d); err != nil {
		t.Fatal(err)
	}
	sc := d["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := sc["WrapX"]; !ok {
		t.Fatalf("not renamed: %v", sc)
	}
	if _, ok := sc["Wrap`1[[X]]"]; ok {
		t.Fatal("old key still present")
	}
	ref := sc["B"].(map[string]any)["properties"].(map[string]any)["w"].(map[string]any)["$ref"]
	if ref != "#/components/schemas/WrapX" {
		t.Fatalf("ref not rewritten: %v", ref)
	}
}
