// Command gen-server turns research/dtos.yaml into one types.go per domain of
// iikoserver. The catalog is the source of truth; these files are not.
package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type field struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	DateFormat  string   `yaml:"date_format"`
	Required    *bool    `yaml:"required"`
	ReadOnly    bool     `yaml:"read_only"`
	Unimplement bool     `yaml:"declared_but_unimplemented"`
	Enum        any      `yaml:"enum"`    // a list of values, or the name of one
	Of          string   `yaml:"of"`      // one nested value of this type
	MapOf       string   `yaml:"map_of"`  // a map keyed by a name the docs do not enumerate
	XML         string   `yaml:"xml"`     // attr | chardata; empty means an element
	Wrapper     string   `yaml:"wrapper"` // list child element name: parent>child
	Aliases     []string `yaml:"aliases"`
	Note        any      `yaml:"note"`
}

// enumValues returns the inline values; a bare string is a reference to an enum
// declared elsewhere and carries none.
func (f field) enumValues() []string { return asStrings(f.Enum) }

type enumDef struct {
	Name   string   `yaml:"name"`
	Values []string `yaml:"values"`
	Note   any      `yaml:"note"`
}

type dto struct {
	Name   string    `yaml:"name"`
	Format string    `yaml:"format"`
	Slug   any       `yaml:"source_slug"` // a slug, or several when domains merged
	UsedBy []string  `yaml:"used_by"`
	Fields []field   `yaml:"fields"`
	Enums  []enumDef `yaml:"enums"`
	Notes  any       `yaml:"notes"`
}

// asStrings normalises the catalog's "one value or a list" fields.
func asStrings(v any) []string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return strings.Join(asStrings(v), " ")
}

func main() {
	root, err := moduleRoot()
	if err != nil {
		fatal(err)
	}
	files, report, err := generate(root)
	if err != nil {
		fatal(err)
	}
	for path, src := range files {
		//nolint:gosec // G306: generated source is committed and must be world-readable
		if err := os.WriteFile(path, src, 0o644); err != nil {
			fatal(err)
		}
	}
	report.print()
}

// generate renders every domain's types.go without writing anything, so the
// same pipeline can be asserted against what is on disk.
func generate(root string) (map[string][]byte, *genReport, error) {
	var dtos []dto
	//nolint:gosec // G304: the path is this repository's own catalog, not input
	raw, err := os.ReadFile(filepath.Join(root, "research", "dtos.yaml"))
	if err != nil {
		return nil, nil, err
	}
	if err := yaml.Unmarshal(raw, &dtos); err != nil {
		return nil, nil, err
	}

	assigned, report := assign(dtos)
	byName := map[string]dto{}
	for _, d := range dtos {
		byName[d.Name] = d
	}
	byDomain := map[string][]dto{}
	for _, d := range dtos {
		if dom := assigned[d.Name]; dom != "" {
			byDomain[dom] = append(byDomain[dom], d)
		}
	}
	// a shared leaf goes wherever it is referenced, since domains cannot import
	// each other and the layout has no shared package
	for name := range duplicated {
		leaf, ok := byName[name]
		if !ok {
			continue
		}
		for _, d := range dtos {
			dom := assigned[d.Name]
			if dom == "" {
				continue
			}
			for _, f := range d.Fields {
				if f.Of != name {
					continue
				}
				if !slices.ContainsFunc(byDomain[dom], func(x dto) bool { return x.Name == name }) {
					byDomain[dom] = append(byDomain[dom], leaf)
				}
				break
			}
		}
	}

	domains := make([]string, 0, len(byDomain))
	for dom := range byDomain {
		domains = append(domains, dom)
	}
	sort.Strings(domains)

	out := map[string][]byte{}
	for _, dom := range domains {
		dir := filepath.Join(root, "iikoserver", dom)
		if _, err := os.Stat(dir); err != nil {
			report.pending[dom] = len(byDomain[dom])
			continue
		}
		existing, err := handWritten(dir)
		if err != nil {
			return nil, nil, err
		}
		var emit []dto
		for _, d := range byDomain[dom] {
			if where, clash := existing[d.Name]; clash {
				report.collisions = append(report.collisions,
					fmt.Sprintf("%s.%s already hand-written in %s", dom, d.Name, where))
				continue
			}
			emit = append(emit, d)
		}
		if len(emit) == 0 {
			continue
		}
		src, err := render(dom, emit)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", dom, err)
		}
		out[filepath.Join(dir, "types.go")] = src
		report.written[dom] = len(emit)
	}
	sort.Strings(report.collisions)
	return out, report, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "gen-server:", err)
	os.Exit(1)
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod above %s", dir)
		}
		dir = parent
	}
}

// handWritten indexes type names already declared by hand, so generation never
// silently replaces a type someone wrote and tested.
func handWritten(dir string) (map[string]string, error) {
	out := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		n := e.Name()
		// _test.go is skipped too: a helper type in a test must not silently
		// suppress a DTO the catalog says this domain has.
		if !strings.HasSuffix(n, ".go") || n == "types.go" || strings.HasSuffix(n, "_test.go") {
			continue
		}
		//nolint:gosec // G304: walking this repository's own package directory
		body, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(body), "\n") {
			if !strings.HasPrefix(line, "type ") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				out[parts[1]] = n
			}
		}
	}
	return out, nil
}

func formatSrc(b []byte) ([]byte, error) {
	src, err := format.Source(b)
	if err != nil {
		return b, fmt.Errorf("generated code does not parse: %w", err)
	}
	return src, nil
}
