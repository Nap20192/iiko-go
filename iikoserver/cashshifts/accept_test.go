package cashshifts

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// This GET creates the acceptance document when none exists, so calling it to
// "just look" changes the shift. The name and godoc have to say so.
func TestClosedSessionDocumentPutsTheSessionInThePath(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"doc-1","session":{"groupId":"g-1"},"items":[{"status":"ACCEPTED"}]}`))
	})
	got, err := s.ClosedSessionDocument(context.Background(), "s-1")
	if err != nil {
		t.Fatalf("document: %v", err)
	}
	if got.ID != "doc-1" {
		t.Errorf("decoded %+v", got)
	}
	if !strings.HasSuffix(gotPath, "/api/v2/cashshifts/closedSessionDocument/s-1") {
		t.Errorf("session id belongs in the path, got %q", gotPath)
	}
	if _, err := s.ClosedSessionDocument(context.Background(), " "); err == nil {
		t.Error("a blank session id must be refused: this call creates a document")
	}
}

// sumReal may be edited only on pay-outs; iiko ignores it elsewhere, which reads
// as "the manager's correction was accepted" when it was silently dropped.
func TestAcceptShiftRejectsEditedSumsOutsidePayOuts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		doc     ClosedSessionDocument
		wantErr string
	}{
		{
			name: "edited pay-out is allowed",
			doc: ClosedSessionDocument{ID: "d-1", Items: []ClosedSessionItem{
				{Type: "PAYOUT", SumReal: 90, SumExpected: 100},
			}},
		},
		{
			name: "edited card total is refused",
			doc: ClosedSessionDocument{ID: "d-1", Items: []ClosedSessionItem{
				{Type: "CARD", SumReal: 90, SumExpected: 100},
			}},
			wantErr: "sumReal is editable only for pay-outs",
		},
		{
			name:    "no id",
			doc:     ClosedSessionDocument{Items: []ClosedSessionItem{{Type: "PAYOUT"}}},
			wantErr: "document id is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(`{"result":"SUCCESS","response":{"id":"d-1"}}`))
			})
			err := s.AcceptShift(context.Background(), tt.doc)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("accept: %v", err)
			}
			if !strings.HasSuffix(gotPath, EndpointCashShiftSave) {
				t.Errorf("path = %q", gotPath)
			}
		})
	}
}

// Cyrillic in a pay-out comment needs the charset spelled out in Content-Type,
// or iiko mangles it.
func TestAddPayOutSendsCharsetForCyrillicComments(t *testing.T) {
	t.Parallel()
	var gotCT, gotBody, gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		gotCT, gotBody, gotPath = r.Header.Get("Content-Type"), string(b), r.URL.Path
		_, _ = w.Write([]byte(`{"result":"SUCCESS","response":{"payOutSettings":{}}}`))
	})
	_, err := s.AddPayOut(context.Background(), PayOutRequest{
		SessionID: "s-1", Sum: 500, Comment: "Такси для курьера",
	})
	if err != nil {
		t.Fatalf("addPayOut: %v", err)
	}
	if !strings.Contains(strings.ToLower(gotCT), "charset=utf-8") {
		t.Errorf("Content-Type must name the charset, got %q", gotCT)
	}
	if !strings.Contains(gotBody, "Такси") {
		t.Errorf("the Cyrillic comment must survive, body was %q", gotBody)
	}
	if !strings.HasSuffix(gotPath, EndpointAddPayOut) {
		t.Errorf("path = %q", gotPath)
	}
	if _, err := s.AddPayOut(context.Background(), PayOutRequest{Sum: 1}); err == nil {
		t.Error("a pay-out without a session must be refused")
	}
}

// Payroll bounds are inclusive at both ends.
func TestPayrollsSendAnInclusiveWindow(t *testing.T) {
	t.Parallel()
	one, _ := rest.ParseDay("2026-03-05")
	var gotQ url.Values
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ, gotPath = r.URL.Query(), r.URL.Path
		_, _ = w.Write([]byte(`[{"id":"p-1","sum":1000}]`))
	})
	got, err := s.Payrolls(context.Background(), one, one, "d-1", false)
	if err != nil {
		t.Fatalf("payrolls: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("decoded %d", len(got))
	}
	if !strings.HasSuffix(gotPath, EndpointPayrolls) {
		t.Errorf("path = %q", gotPath)
	}
	if gotQ.Get("dateFrom") != "2026-03-05" || gotQ.Get("dateTo") != "2026-03-05" {
		t.Errorf("one inclusive day must send both bounds as itself, got %v", gotQ)
	}
	if gotQ.Get("includeDeleted") != "false" {
		t.Errorf("includeDeleted must be explicit, got %q", gotQ.Get("includeDeleted"))
	}
}

func TestPayInOutTypesHitsItsPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`[{"id":"t-1","name":"Инкассация"}]`))
	})
	got, err := s.PayInOutTypes(context.Background(), false)
	if err != nil {
		t.Fatalf("types: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("decoded %d", len(got))
	}
	if !strings.HasSuffix(gotPath, EndpointPayInOutTypes) {
		t.Errorf("path = %q", gotPath)
	}
}
