package documents

import (
	"encoding/json"
	"strings"
	"testing"
)

// iiko reads the presence of id as "edit this document"; its absence as
// "create a new one". Sending an empty id on create would be an edit of nothing,
// and dropping a real id on edit would silently create a duplicate document.
func TestDocumentIDIsSentOnlyWhenSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		doc    any
		id     string
		wantID bool
	}{
		{name: "new writeoff carries no id", doc: WriteoffDocumentDto{}},
		{name: "edited writeoff carries its id", doc: WriteoffDocumentDto{ID: "w-1"}, id: "w-1", wantID: true},
		{name: "new transfer carries no id", doc: InternalTransferDto{}},
		{name: "edited transfer carries its id", doc: InternalTransferDto{ID: "t-1"}, id: "t-1", wantID: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw, err := json.Marshal(tt.doc)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("decode own output: %v", err)
			}
			id, present := got["id"]
			if present != tt.wantID {
				t.Fatalf("id present = %v, want %v; body was %s", present, tt.wantID, raw)
			}
			if tt.wantID && id != tt.id {
				t.Errorf("id = %v, want %q", id, tt.id)
			}
		})
	}
}

// The other half of the same rule: fields the docs call "strip before re-import"
// must never leave, however they were populated by an export.
func TestInvoiceReadOnlyFieldsAreNotSent(t *testing.T) {
	t.Parallel()
	doc := IncomingInvoiceDto{ID: "inv-1", DocumentNumber: "N-1"}
	doc.StripReadOnly()
	if doc.ID != "" {
		t.Errorf("StripReadOnly must clear the exported id, got %q", doc.ID)
	}
	if doc.DocumentNumber != "N-1" {
		t.Errorf("StripReadOnly must leave writable fields alone, got %q", doc.DocumentNumber)
	}
}

func TestWriteoffMarshalOmitsServerComputedLineFields(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(WriteoffDocumentItemDto{ProductID: "p-1", Amount: 2, Num: 7, Cost: 99})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)
	// num, measureUnitId and cost are computed by iiko; sending them is rejected.
	for _, banned := range []string{`"num"`, `"cost"`, `"measureUnitId"`} {
		if strings.Contains(body, banned) {
			t.Errorf("%s is server-computed and must not be sent; body was %s", banned, body)
		}
	}
	if !strings.Contains(body, `"productId"`) || !strings.Contains(body, `"amount"`) {
		t.Errorf("writable line fields must survive; body was %s", body)
	}
}
