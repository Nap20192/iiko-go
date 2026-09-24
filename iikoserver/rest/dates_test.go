package rest

import "testing"

func TestDateFamiliesDiffer(t *testing.T) {
	d, err := ParseDay("2026-03-05")
	if err != nil {
		t.Fatal(err)
	}
	// These four are the formats that a single global formatter would collapse.
	for _, tc := range []struct{ layout, want, why string }{
		{ReportV1, "05.03.2026", "v1 /reports/* take DD.MM.YYYY"},
		{QueryV2, "2026-03-05", "v2 query params take yyyy-MM-dd"},
		{OlapMs, "2026-03-05T00:00:00.000", "OLAP DateRange needs milliseconds"},
		{DocMinute, "2026-03-05T00:00", "v2 document dateIncoming stops at minutes"},
	} {
		if got := d.Format(tc.layout); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.why, got, tc.want)
		}
	}

	if _, err := ParseDay("05/03/2026"); err == nil {
		t.Error("unparseable date must be rejected with a usable message")
	}
}

func TestParseFlexible(t *testing.T) {
	t.Parallel()
	// Responses are far less consistent than requests: XML carries an offset,
	// cashshift JSON has none and variable fractional digits, and legacy OLAP
	// OpenDate renders dot-separated.
	tests := []struct {
		name, in string
		wantYMD  string
		ok       bool
	}{
		{name: "XML with offset", in: "2012-07-04T23:00:00+04:00", wantYMD: "2012-07-04", ok: true},
		{name: "cashshift, two fractional digits", in: "2017-02-21T22:28:18.63", wantYMD: "2017-02-21", ok: true},
		{name: "stamp, no fraction", in: "2026-03-05T12:00:00", wantYMD: "2026-03-05", ok: true},
		{name: "v2 document, minutes only", in: "2026-03-05T12:00", wantYMD: "2026-03-05", ok: true},
		{name: "plain ISO date", in: "2026-03-05", wantYMD: "2026-03-05", ok: true},
		{name: "v1 report, dotted DMY", in: "05.03.2026", wantYMD: "2026-03-05", ok: true},
		{name: "legacy OLAP OpenDate, dotted YMD", in: "2014.01.01", wantYMD: "2014-01-01", ok: true},
		{name: "garbage is reported, not guessed", in: "yesterday", ok: false},
		{name: "empty", in: "", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ParseFlexible(tt.in)
			if ok != tt.ok {
				t.Fatalf("ParseFlexible(%q) ok = %v, want %v", tt.in, ok, tt.ok)
			}
			if ok && got.Format(QueryV2) != tt.wantYMD {
				t.Errorf("parsed %q as %s, want %s", tt.in, got.Format(QueryV2), tt.wantYMD)
			}
		})
	}
}
