package main

import (
	"strings"
	"testing"
)

func TestEmitOperationsSkipsDeprecated(t *testing.T) {
	t.Parallel()

	doc := mustDoc(t, `{"servers":[{"url":"https://api-ru.iiko.services"}],
	 "components":{"schemas":{"Req":{"type":"object"},"Resp":{"type":"object"}}},
	 "paths":{
	  "/api/1/live":{"post":{"tags":["Menu"],"summary":"Live one.",
	    "requestBody":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Req"}}}},
	    "responses":{"200":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Resp"}}}}}}},
	  "/api/1/old":{"post":{"tags":["Deprecated"],"deprecated":true,"responses":{"200":{}}}}}}`)

	out, err := EmitOperations(doc, "gen")
	if err != nil {
		t.Fatal(err)
	}
	src := string(out)
	wantIn(t, src,
		`DefaultBaseURL = "https://api-ru.iiko.services"`,
		`{Path: "/api/1/live", Tag: "Menu", Summary: "Live one.", Request: "Req", Response: "Resp"}`,
	)
	if strings.Contains(src, "/api/1/old") {
		t.Error("deprecated operation must not be generated")
	}
}

func TestEmitOperationsNamesInlineArrayResponses(t *testing.T) {
	t.Parallel()

	// Sixteen operations answer with a bare array instead of a named wrapper;
	// dropping them would silently leave those endpoints untyped.
	doc := mustDoc(t, `{"components":{"schemas":{"Item":{"type":"object"}}},
	 "paths":{"/api/1/list":{"post":{"tags":["T"],"summary":"s",
	   "responses":{"200":{"content":{"application/json":{"schema":
	     {"type":"array","items":{"$ref":"#/components/schemas/Item"}}}}}}}},
	  "/api/1/void":{"post":{"tags":["T"],"summary":"s","responses":{"200":{"description":"Success"}}}}}}`)

	out, err := EmitOperations(doc, "gen")
	if err != nil {
		t.Fatal(err)
	}
	wantIn(t, string(out),
		`{Path: "/api/1/list", Tag: "T", Summary: "s", Request: "", Response: "[]Item"}`,
		`{Path: "/api/1/void", Tag: "T", Summary: "s", Request: "", Response: ""}`,
	)
}
