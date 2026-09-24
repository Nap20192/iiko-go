package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Overlay is one YAML file of repairs for a known defect class in the upstream spec.
type Overlay struct {
	Name    string   `yaml:"-"`
	Because string   `yaml:"because"`
	Actions []Action `yaml:"actions"`
}

// Action is a single guarded edit. Expect pins the current content so a silent
// upstream fix breaks the build instead of being patched over.
type Action struct {
	Because string         `yaml:"because"`
	Op      string         `yaml:"op"`
	Target  string         `yaml:"target"`
	Expect  string         `yaml:"expect"`
	Value   any            `yaml:"value"`
	Match   map[string]any `yaml:"match"`
	Count   int            `yaml:"count"`
}

// Digest is the guard value: sha256 over the canonical JSON of a spec fragment.
func Digest(v any) string {
	b, err := json.Marshal(normalize(v))
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// normalize rounds YAML-decoded values through JSON so a guard computed from a
// YAML literal matches one computed from the JSON spec.
func normalize(v any) any {
	switch t := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[fmt.Sprint(k)] = normalize(val)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = normalize(val)
		}
		return m
	case []any:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = normalize(val)
		}
		return s
	case int:
		return float64(t)
	default:
		return v
	}
}

// Apply runs every action in order, failing on the first stale guard.
func (o Overlay) Apply(doc map[string]any) error {
	for i, a := range o.Actions {
		if err := a.apply(doc); err != nil {
			return fmt.Errorf("%s: action %d (%s %s): %w", o.Name, i, a.Op, a.Target, err)
		}
	}
	return nil
}

func (a Action) apply(doc map[string]any) error {
	switch a.Op {
	case "set":
		return a.set(doc)
	case "merge":
		return a.merge(doc)
	case "removeParameter":
		return a.removeParameter(doc)
	case "renameSchema":
		return a.renameSchema(doc)
	default:
		return fmt.Errorf("unknown op %q", a.Op)
	}
}

func (a Action) set(doc map[string]any) error {
	parent, key, err := resolveParent(doc, a.Target)
	if err != nil {
		return err
	}
	if err := a.check(parent[key], hasKey(parent, key)); err != nil {
		return err
	}
	parent[key] = normalize(a.Value)
	return nil
}

// removeParameter is a whole-document rule: the defect it repairs is repeated on
// every operation, so it is guarded by a count rather than by a fragment hash.
func (a Action) removeParameter(doc map[string]any) error {
	n := 0
	paths, _ := doc["paths"].(map[string]any)
	for _, item := range paths {
		op, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for _, v := range op {
			o, ok := v.(map[string]any)
			if !ok {
				continue
			}
			ps, ok := o["parameters"].([]any)
			if !ok {
				continue
			}
			kept := ps[:0]
			for _, p := range ps {
				if pm, ok := p.(map[string]any); ok && matches(pm, a.Match) {
					n++
					continue
				}
				kept = append(kept, p)
			}
			o["parameters"] = kept
		}
	}
	if n != a.Count {
		return fmt.Errorf("guard count %d: matched %d", a.Count, n)
	}
	return nil
}

func matches(m, want map[string]any) bool {
	for k, v := range want {
		if fmt.Sprint(m[k]) != fmt.Sprint(v) {
			return false
		}
	}
	return true
}

// renameSchema retargets a component key and every $ref pointing at it; the
// upstream CLR-qualified names are not usable as identifiers by any generator.
func (a Action) renameSchema(doc map[string]any) error {
	comps, _ := doc["components"].(map[string]any)
	schemas, ok := comps["schemas"].(map[string]any)
	if !ok {
		return fmt.Errorf("no components.schemas")
	}
	cur, present := schemas[a.Target]
	if err := a.check(cur, present); err != nil {
		return err
	}
	to, ok := a.Value.(string)
	if !ok {
		return fmt.Errorf("renameSchema value must be a string")
	}
	if _, clash := schemas[to]; clash {
		return fmt.Errorf("rename target %q already exists", to)
	}
	delete(schemas, a.Target)
	schemas[to] = cur
	rewriteRefs(doc, "#/components/schemas/"+a.Target, "#/components/schemas/"+to)
	return nil
}

func rewriteRefs(node any, from, to string) {
	switch t := node.(type) {
	case map[string]any:
		if r, ok := t["$ref"].(string); ok && r == from {
			t["$ref"] = to
		}
		for _, v := range t {
			rewriteRefs(v, from, to)
		}
	case []any:
		for _, v := range t {
			rewriteRefs(v, from, to)
		}
	}
}

func (a Action) merge(doc map[string]any) error {
	parent, key, err := resolveParent(doc, a.Target)
	if err != nil {
		return err
	}
	cur, ok := parent[key].(map[string]any)
	if err := a.check(parent[key], hasKey(parent, key)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("merge target is not an object")
	}
	patch, ok := normalize(a.Value).(map[string]any)
	if !ok {
		return fmt.Errorf("merge value is not an object")
	}
	for k, v := range patch {
		cur[k] = v
	}
	return nil
}

// check compares the current fragment against the pinned guard. "absent" pins
// the absence of the key, which is what most structural repairs rely on.
func (a Action) check(cur any, present bool) error {
	switch {
	case a.Expect == "":
		return fmt.Errorf("guard missing: expect is required")
	case a.Expect == "absent":
		if present {
			return fmt.Errorf("guard expected absent, found %s", Digest(cur))
		}
	case !present:
		return fmt.Errorf("guard %s: target absent", a.Expect)
	default:
		if got := Digest(cur); got != a.Expect {
			return fmt.Errorf("guard %s: found %s", a.Expect, got)
		}
	}
	return nil
}

func hasKey(m map[string]any, k string) bool {
	_, ok := m[k]
	return ok
}

// resolveParent walks a JSON pointer to the container of its last token.
func resolveParent(doc map[string]any, ptr string) (parent map[string]any, key string, err error) {
	toks, err := pointerTokens(ptr)
	if err != nil {
		return nil, "", err
	}
	cur := doc
	for _, t := range toks[:len(toks)-1] {
		next, ok := cur[t].(map[string]any)
		if !ok {
			return nil, "", fmt.Errorf("path %q: %q is not an object", ptr, t)
		}
		cur = next
	}
	return cur, toks[len(toks)-1], nil
}

func pointerTokens(ptr string) ([]string, error) {
	if !strings.HasPrefix(ptr, "/") {
		return nil, fmt.Errorf("pointer %q must start with /", ptr)
	}
	toks := strings.Split(ptr[1:], "/")
	for i, t := range toks {
		toks[i] = strings.NewReplacer("~1", "/", "~0", "~").Replace(t)
	}
	return toks, nil
}
