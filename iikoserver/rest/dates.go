package rest

import (
	"fmt"
	"time"
)

// iikoServer uses six incompatible date conventions, split by endpoint family.
// A single global formatter is guaranteed to break, so callers pick explicitly.
//
//	ReportV1  dd.MM.yyyy               all v1 /reports/*, /suppliers/{code}/pricelist
//	QueryV2   yyyy-MM-dd               v2 query params, documents, employees, cashshifts
//	OlapMs    yyyy-MM-dd'T'HH:mm:ss.SSS OLAP v2 DateRange filters, events
//	Stamp     yyyy-MM-dd'T'HH:mm:ss    /v2/reports/balance/* timestamp, byPresetId
//	DocMinute yyyy-MM-dd'T'HH:mm       v2 JSON document dateIncoming
//	InvoiceXML dd.MM.yyyy              incomingInvoice IMPORT xml only (export emits ISO)
const (
	ReportV1   = "02.01.2006"
	QueryV2    = "2006-01-02"
	OlapMs     = "2006-01-02T15:04:05.000"
	Stamp      = "2006-01-02T15:04:05"
	DocMinute  = "2006-01-02T15:04"
	InvoiceXML = "02.01.2006"
)

// ParseDay accepts either ISO or Russian dotted form, so a tool caller can
// write whichever it has and we normalize on the way out.
func ParseDay(s string) (time.Time, error) {
	for _, layout := range []string{QueryV2, ReportV1, Stamp, OlapMs} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date %q: use YYYY-MM-DD", s)
}

// ParseFlexible handles response values, which vary even more than requests:
// XML carries an offset, cashshift JSON has none and variable fractional
// digits, and legacy OLAP OpenDate renders dot-separated ("2014.01.01").
func ParseFlexible(s string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339, "2006-01-02T15:04:05.999999999", Stamp, DocMinute,
		QueryV2, ReportV1, "2006.01.02", "2006.01.02 15:04:05",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
