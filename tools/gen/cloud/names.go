package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// schemaNames assigns each component schema a short Go name, lengthening a name
// with its enclosing namespace segments only as far as collisions require.
func schemaNames(schemas map[string]any) map[string]string {
	keys := make([]string, 0, len(schemas))
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	segs := make(map[string][]string, len(keys))
	depth := make(map[string]int, len(keys))
	for _, k := range keys {
		segs[k] = splitSegments(k)
		depth[k] = 1
	}

	name := func(k string) string {
		s := segs[k]
		d := min(depth[k], len(s))
		return strings.Join(mapSlice(s[len(s)-d:], exported), "")
	}

	for range 8 {
		byName := map[string][]string{}
		for _, k := range keys {
			n := name(k)
			byName[n] = append(byName[n], k)
		}
		grew := false
		for _, group := range byName {
			if len(group) < 2 {
				continue
			}
			for _, k := range group {
				if depth[k] < len(segs[k]) {
					depth[k]++
					grew = true
				}
			}
		}
		if !grew {
			break
		}
	}

	out := make(map[string]string, len(keys))
	used := map[string]bool{}
	for _, k := range keys {
		n := name(k)
		for i := 2; used[n]; i++ {
			n = fmt.Sprintf("%s%d", name(k), i)
		}
		used[n] = true
		out[k] = n
	}
	return out
}

var nonWord = regexp.MustCompile(`[^A-Za-z0-9]+`)

// splitSegments turns a CLR-ish or snake_case component key into namespace parts.
func splitSegments(key string) []string {
	var out []string
	for _, part := range strings.Split(key, ".") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		out = []string{"Unnamed"}
	}
	return out
}

var initialisms = map[string]string{
	"id": "ID", "ids": "IDs", "url": "URL", "uri": "URI", "api": "API",
	"http": "HTTP", "https": "HTTPS", "sms": "SMS", "vat": "VAT",
	"json": "JSON", "xml": "XML", "guid": "GUID", "uuid": "UUID",
	"inn": "INN", "kpp": "KPP", "sku": "SKU", "qr": "QR", "gtin": "GTIN",
}

// exported turns any upstream spelling into an exported Go identifier.
func exported(s string) string {
	words := splitWords(s)
	var b strings.Builder
	for _, w := range words {
		if up, ok := initialisms[strings.ToLower(w)]; ok {
			b.WriteString(up)
			continue
		}
		b.WriteString(strings.ToUpper(w[:1]) + w[1:])
	}
	out := b.String()
	if out == "" {
		return "Field"
	}
	if out[0] >= '0' && out[0] <= '9' {
		return "N" + out
	}
	return out
}

func splitWords(s string) []string {
	var words []string
	for _, chunk := range nonWord.Split(s, -1) {
		if chunk == "" {
			continue
		}
		start := 0
		for i := 1; i < len(chunk); i++ {
			prev, cur := chunk[i-1], chunk[i]
			lowerToUpper := prev >= 'a' && prev <= 'z' && cur >= 'A' && cur <= 'Z'
			digitBoundary := (prev >= '0' && prev <= '9') != (cur >= '0' && cur <= '9')
			if lowerToUpper || digitBoundary {
				words = append(words, chunk[start:i])
				start = i
			}
		}
		words = append(words, chunk[start:])
	}
	return words
}

func mapSlice[T, U any](in []T, f func(T) U) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}
