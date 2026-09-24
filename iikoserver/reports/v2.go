package reports

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// OlapPresets lists every saved, non-deleted report configuration.
//
// Prefer fetching a preset here and POSTing it to /olap yourself: the docs warn
// that byPresetId's output shape shifts between upgrades.
func (s *Service) OlapPresets(ctx context.Context) ([]OlapPreset, error) {
	var out []OlapPreset
	if _, err := s.rest.GetV2(ctx, EndpointOlapPresets, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// presetTypes are the documented values, lowercase on the wire.
var presetTypes = map[string]bool{"stock": true, "sales": true, "transactions": true, "deliveries": true}

// OlapPresetsByType lists presets of one type. The path segment is lowercase.
func (s *Service) OlapPresetsByType(ctx context.Context, presetType string) ([]OlapPreset, error) {
	t := strings.ToLower(strings.TrimSpace(presetType))
	if !presetTypes[t] {
		return nil, fmt.Errorf("unknown preset type %q; valid: stock, sales, transactions, deliveries", presetType)
	}
	var out []OlapPreset
	if _, err := s.rest.GetV2(ctx, trimTemplate(EndpointOlapPresetsByType)+t, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// OlapByPresetID runs a saved report.
//
// dateFrom is inclusive and dateTo EXCLUSIVE, both stamped to the second, and
// the totals toggle is named `summary` here where the POST body calls it
// `buildSummary`. The docs recommend not using this at all — its output format
// changes across upgrades, unlike a preset you fetch and post yourself.
func (s *Service) OlapByPresetID(ctx context.Context, presetID string, from, to time.Time, summary bool) ([]map[string]any, error) {
	if strings.TrimSpace(presetID) == "" {
		return nil, fmt.Errorf("presetId is required")
	}
	q := url.Values{
		"dateFrom": {from.Format(rest.Stamp)},
		"dateTo":   {to.Format(rest.Stamp)},
		"summary":  {rest.BoolStr(summary)},
	}
	var out []map[string]any
	if _, err := s.rest.GetV2(ctx, trimTemplate(EndpointOlapByPresetID)+url.PathEscape(presetID), q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CounteragentBalance is one counteragent's balance at a moment.
type CounteragentBalance struct {
	Counteragent string  `json:"counteragent"`
	Account      string  `json:"account"`
	Sum          float64 `json:"sum"`
}

// CounteragentBalances reports balances at a timestamp.
//
// Always prefer this over OLAP's StartBalance/FinalBalance, which scan the whole
// transaction history and can hang the server.
func (s *Service) CounteragentBalances(ctx context.Context, at time.Time, counteragents, accounts []string) ([]CounteragentBalance, error) {
	q := url.Values{"timestamp": {at.Format(rest.Stamp)}}
	for _, c := range counteragents {
		q.Add("counteragent", c)
	}
	for _, a := range accounts {
		q.Add("account", a)
	}
	var out []CounteragentBalance
	if _, err := s.rest.GetV2(ctx, EndpointBalanceCounteragents, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// egaisNotWrittenOff is the sentinel iiko uses for "still in stock". Reading it
// as a date reports the stock as disposed of in the year 2500.
const egaisNotWrittenOff = "2500-01-01T00:00:00"

// EgaisMark is one excise mark on the third register.
type EgaisMark struct {
	Mark         string `json:"mark"`
	WriteoffDate string `json:"writeoffDate"`
}

// NotWrittenOff reports whether this mark is still in stock.
func (m EgaisMark) NotWrittenOff() bool { return m.WriteoffDate == egaisNotWrittenOff }

// EgaisMarks is the marks response; FullUpdate means the caller's cache is stale
// and must be discarded rather than merged.
type EgaisMarks struct {
	FullUpdate bool        `json:"fullUpdate"`
	Marks      []EgaisMark `json:"marks"`
}

// EgaisMarks lists excise marks, optionally only those changed since a revision.
func (s *Service) EgaisMarks(ctx context.Context, fsRarIDs []string, revisionFrom int) (*EgaisMarks, error) {
	q := url.Values{"revisionFrom": {fmt.Sprint(revisionFrom)}}
	for _, id := range fsRarIDs {
		q.Add("fsRarId", id)
	}
	var out EgaisMarks
	if _, err := s.rest.GetV2(ctx, EndpointEgaisMarks, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// trimTemplate turns a path constant's {placeholder} tail into a prefix.
func trimTemplate(p string) string {
	if i := strings.Index(p, "{"); i >= 0 {
		return p[:i]
	}
	return p
}
