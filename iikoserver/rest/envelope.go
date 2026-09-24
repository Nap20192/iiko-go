package rest

import (
	"encoding/json"
	"fmt"
	"strings"
)

// iiko v2 JSON is inconsistent about envelopes, and this is a real parser trap:
//
//	list / create        -> {"result":"SUCCESS","errors":[],"response":…,"revision":N}
//	byId                 -> the bare object
//	byNumber             -> a bare array
//
// On top of that, "result":"ERROR" arrives with HTTP 200 — status alone is not
// success. DecodeV2 normalizes all of it: it unwraps when an envelope is
// present, passes bare payloads through, and turns result=ERROR into an error
// carrying the server's own messages.
type v2Envelope struct {
	Result   string          `json:"result"`
	Errors   json.RawMessage `json:"errors"`
	Response json.RawMessage `json:"response"`
	Revision *int64          `json:"revision"`
}

// v2Error covers both shapes seen in the wild: {value, code} objects and plain
// strings, which different endpoints and doc samples disagree about.
type v2Error struct {
	Value string `json:"value"`
	Code  string `json:"code"`
}

func (e v2Error) String() string {
	switch {
	case e.Value != "" && e.Code != "":
		return e.Code + ": " + e.Value
	case e.Value != "":
		return e.Value
	default:
		return e.Code
	}
}

// DecodeV2 unmarshals a v2 response body into out, tolerating both the
// enveloped and bare forms. The returned revision is non-nil only when the
// server supplied one (the 7.8-era endpoints do; older ones make you track it).
func DecodeV2(data []byte, out any) (revision *int64, err error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("empty response body")
	}

	// A bare array can never be an envelope, so don't even probe.
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(data, out); err != nil {
			return nil, fmt.Errorf("decode bare array: %w", err)
		}
		return nil, nil
	}

	var env v2Envelope
	if err := json.Unmarshal(data, &env); err == nil && env.Result != "" {
		if !strings.EqualFold(env.Result, "SUCCESS") {
			return nil, &Error{
				Status: 200, // iiko reports business failures with HTTP 200
				Body:   "result=" + env.Result + "; " + formatV2Errors(env.Errors),
				Hint:   "The request was accepted but rejected by business logic. The message above comes from iiko and is written for a human — read it literally.",
			}
		}
		if len(env.Response) > 0 {
			if err := json.Unmarshal(env.Response, out); err != nil {
				return nil, fmt.Errorf("decode enveloped response: %w", err)
			}
			return env.Revision, nil
		}
		// SUCCESS with no response payload: nothing to decode.
		return env.Revision, nil
	}

	// No envelope — a bare object from byId.
	if err := json.Unmarshal(data, out); err != nil {
		return nil, fmt.Errorf("decode bare object: %w", err)
	}
	return nil, nil
}

func formatV2Errors(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "no error detail"
	}
	var objs []v2Error
	if err := json.Unmarshal(raw, &objs); err == nil && len(objs) > 0 {
		parts := make([]string, 0, len(objs))
		for _, o := range objs {
			parts = append(parts, o.String())
		}
		return strings.Join(parts, "; ")
	}
	var strs []string
	if err := json.Unmarshal(raw, &strs); err == nil && len(strs) > 0 {
		return strings.Join(strs, "; ")
	}
	return string(raw)
}
