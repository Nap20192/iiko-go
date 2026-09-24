package rest

import (
	"strings"
	"testing"
)

func TestDecodeV2HandlesAllThreeShapes(t *testing.T) {
	type doc struct {
		ID     string `json:"id"`
		Number string `json:"documentNumber"`
	}

	t.Run("enveloped list", func(t *testing.T) {
		var out []doc
		rev, err := DecodeV2([]byte(`{"result":"SUCCESS","errors":[],"response":[{"id":"a"},{"id":"b"}],"revision":42}`), &out)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 2 || out[0].ID != "a" {
			t.Fatalf("bad decode: %+v", out)
		}
		if rev == nil || *rev != 42 {
			t.Fatal("revision cursor must survive; only the 7.8-era endpoints supply one")
		}
	})

	t.Run("bare object from byId", func(t *testing.T) {
		var out doc
		rev, err := DecodeV2([]byte(`{"id":"x","documentNumber":"7"}`), &out)
		if err != nil {
			t.Fatal(err)
		}
		if out.ID != "x" || out.Number != "7" {
			t.Fatalf("bad decode: %+v", out)
		}
		if rev != nil {
			t.Fatal("byId supplies no revision")
		}
	})

	t.Run("bare array from byNumber", func(t *testing.T) {
		var out []doc
		if _, err := DecodeV2([]byte(`[{"id":"z"}]`), &out); err != nil {
			t.Fatal(err)
		}
		if len(out) != 1 || out[0].ID != "z" {
			t.Fatalf("bad decode: %+v", out)
		}
	})
}

// iiko reports business failures with HTTP 200 and result=ERROR. Treating
// status alone as success silently swallows them.
func TestDecodeV2TreatsResultErrorAsFailure(t *testing.T) {
	var out []struct{}
	_, err := DecodeV2([]byte(`{"result":"ERROR","errors":[{"code":"E42","value":"Склад не указан"}],"response":null}`), &out)
	if err == nil {
		t.Fatal(`result=ERROR with HTTP 200 must be an error`)
	}
	if !strings.Contains(err.Error(), "Склад не указан") {
		t.Fatalf("server's own message must reach the caller: %v", err)
	}
	if !strings.Contains(err.Error(), "E42") {
		t.Fatalf("error code should be preserved: %v", err)
	}
}

// Doc samples and endpoints disagree on whether errors are objects or strings.
func TestDecodeV2ToleratesStringErrors(t *testing.T) {
	var out []struct{}
	_, err := DecodeV2([]byte(`{"result":"ERROR","errors":["plain message"]}`), &out)
	if err == nil || !strings.Contains(err.Error(), "plain message") {
		t.Fatalf("string-shaped errors must decode too: %v", err)
	}
}
func TestDecodeV2RejectsEmptyBody(t *testing.T) {
	var out []struct{}
	if _, err := DecodeV2(nil, &out); err == nil {
		t.Fatal("empty body must not silently succeed")
	}
}
