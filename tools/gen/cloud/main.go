// Command cloud regenerates iikocloud/gen from the live iikoCloud OpenAPI
// spec, which is hash-pinned rather than committed and repaired by guarded
// overlays, so upstream drift fails the build instead of being patched over.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const specURL = "https://api-ru.iiko.services/api-docs/docs"

func main() {
	var (
		url      = flag.String("spec", specURL, "OpenAPI spec URL")
		from     = flag.String("from", "", "read the spec from this file instead of the network")
		out      = flag.String("out", "iikocloud/gen", "output package directory")
		overlays = flag.String("overlays", "iikocloud/overlays", "overlay directory")
		pin      = flag.String("pin", "tools/gen/cloud/spec.sha256", "file holding the expected spec digest")
		repin    = flag.Bool("repin", false, "rewrite the pin file with the fetched digest")
		save     = flag.String("save", "", "also write the repaired spec here")
		digests  = flag.Bool("digests", false, "read JSON pointers on stdin, print overlay guard digests")
	)
	flag.Parse()

	if *digests {
		if err := printDigests(*url, *from); err != nil {
			fmt.Fprintln(os.Stderr, "gen/cloud:", err)
			os.Exit(1)
		}
		return
	}
	if err := run(*url, *from, *out, *overlays, *pin, *save, *repin); err != nil {
		fmt.Fprintln(os.Stderr, "gen/cloud:", err)
		os.Exit(1)
	}
}

func run(url, from, out, overlayDir, pin, save string, repin bool) error {
	raw, err := fetch(url, from)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	got := hex.EncodeToString(sum[:])
	if err := checkPin(pin, got, repin); err != nil {
		return err
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("spec is not JSON: %w", err)
	}

	applied, err := applyOverlays(overlayDir, doc)
	if err != nil {
		return err
	}

	types, err := Emit(doc, "gen")
	if err != nil {
		return err
	}
	ops, err := EmitOperations(doc, "gen")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	for name, body := range map[string][]byte{"types.go": types, "operations.go": ops} {
		if err := os.WriteFile(filepath.Join(out, name), body, 0o600); err != nil {
			return err
		}
	}
	if save != "" {
		repaired, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(save, repaired, 0o600); err != nil {
			return err
		}
	}

	schemas, _ := dig(doc, "components", "schemas")
	live, deprecated := countOperations(doc)
	fmt.Printf("spec %s (%.1f MB, sha256 %s)\n", url, float64(len(raw))/(1<<20), got[:16])
	fmt.Printf("overlays applied: %s\n", strings.Join(applied, ", "))
	fmt.Printf("schemas %d, operations %d live / %d deprecated (not generated)\n", len(schemas), live, deprecated)
	return nil
}

// printDigests computes guard values for overlay authoring: the guards must be
// produced by the same code that checks them, or they would never match.
func printDigests(url, from string) error {
	raw, err := fetch(url, from)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		ptr := strings.TrimSpace(sc.Text())
		if ptr == "" {
			continue
		}
		parent, key, err := resolveParent(doc, ptr)
		if err != nil {
			return err
		}
		if _, ok := parent[key]; !ok {
			fmt.Printf("%s\tabsent\n", ptr)
			continue
		}
		fmt.Printf("%s\t%s\n", ptr, Digest(parent[key]))
	}
	return sc.Err()
}

func fetch(url, from string) ([]byte, error) {
	if from != "" {
		return os.ReadFile(from) //nolint:gosec // a developer tool reading the path its own flag names
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// checkPin fails the build when upstream changed, so overlays are re-verified by
// a human rather than by luck.
func checkPin(path, got string, repin bool) error {
	if repin {
		return os.WriteFile(path, []byte(got+"\n"), 0o600)
	}
	want, err := os.ReadFile(path) //nolint:gosec // the pin file, named by a flag of this tool
	if err != nil {
		return fmt.Errorf("%s: %w (run with -repin after reviewing the diff)", path, err)
	}
	if w := strings.TrimSpace(string(want)); w != got {
		return fmt.Errorf("spec changed: pinned %s, fetched %s (review, then -repin)", w[:min(16, len(w))], got[:16])
	}
	return nil
}

func applyOverlays(dir string, doc map[string]any) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	applied := make([]string, 0, len(files))
	for _, f := range files {
		body, err := os.ReadFile(filepath.Join(dir, f)) //nolint:gosec // overlay files enumerated from the overlay directory
		if err != nil {
			return nil, err
		}
		var o Overlay
		if err := yaml.Unmarshal(body, &o); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		o.Name = f
		if err := o.Apply(doc); err != nil {
			return nil, err
		}
		applied = append(applied, fmt.Sprintf("%s(%d)", strings.TrimSuffix(f, ".yaml"), len(o.Actions)))
	}
	return applied, nil
}

func countOperations(doc map[string]any) (live, deprecated int) {
	paths, _ := doc["paths"].(map[string]any)
	for _, item := range paths {
		m, _ := item.(map[string]any)
		for method, v := range m {
			op, ok := v.(map[string]any)
			if !ok || !isHTTPMethod(method) {
				continue
			}
			if op["deprecated"] == true {
				deprecated++
			} else {
				live++
			}
		}
	}
	return live, deprecated
}

func isHTTPMethod(s string) bool {
	switch s {
	case "get", "post", "put", "delete", "patch", "head", "options":
		return true
	}
	return false
}
