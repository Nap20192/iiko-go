package main

import (
	"fmt"
	"sort"
	"strings"
)

type genReport struct {
	written    map[string]int
	pending    map[string]int
	collisions []string
	unresolved []string
	skippedFor map[string]string
}

// assign decides which domain owns each DTO: an explicit override first, then
// the endpoints it is used by, then the domain of the DTO that embeds it.
func assign(dtos []dto) (map[string]string, *genReport) {
	rep := &genReport{
		written: map[string]int{}, pending: map[string]int{},
		skippedFor: map[string]string{},
	}
	byName := map[string]dto{}
	for _, d := range dtos {
		byName[d.Name] = d
	}

	out := map[string]string{}
	for _, d := range dtos {
		if why, ok := skipped[d.Name]; ok {
			rep.skippedFor[d.Name] = why
			continue
		}
		if dom, ok := overrides[d.Name]; ok {
			out[d.Name] = dom
			continue
		}
		if dom := fromPaths(d); dom != "" {
			out[d.Name] = dom
		}
	}

	// A DTO reachable only as another DTO's field inherits that DTO's domain.
	for range dtos {
		progress := false
		for _, d := range dtos {
			if out[d.Name] != "" {
				continue
			}
			if _, isSkipped := skipped[d.Name]; isSkipped {
				continue
			}
			for _, u := range d.UsedBy {
				owner, _, isField := strings.Cut(u, ".")
				if !isField {
					continue
				}
				if dom := out[owner]; dom != "" {
					out[d.Name] = dom
					progress = true
					break
				}
			}
		}
		if !progress {
			break
		}
	}

	for _, d := range dtos {
		if _, isSkipped := skipped[d.Name]; isSkipped {
			continue
		}
		if out[d.Name] == "" {
			rep.unresolved = append(rep.unresolved,
				fmt.Sprintf("%s (used_by %v)", d.Name, d.UsedBy))
		}
	}
	sort.Strings(rep.unresolved)
	return out, rep
}

// fromPaths returns the single domain owning every /resto path this DTO serves.
// A DTO spanning domains is left to an override rather than guessed.
func fromPaths(d dto) string {
	seen := map[string]bool{}
	for _, u := range d.UsedBy {
		if !strings.HasPrefix(u, "/resto") {
			continue
		}
		if dom := domainOf(u); dom != "" {
			seen[dom] = true
		}
	}
	if len(seen) != 1 {
		return ""
	}
	for dom := range seen {
		return dom
	}
	return ""
}

func (r *genReport) print() {
	doms := make([]string, 0, len(r.written))
	for d := range r.written {
		doms = append(doms, d)
	}
	sort.Strings(doms)
	total := 0
	for _, d := range doms {
		fmt.Printf("  %-14s %3d types -> iikoserver/%s/types.go\n", d, r.written[d], d)
		total += r.written[d]
	}
	fmt.Printf("  generated %d types across %d domains\n", total, len(doms))

	if len(r.pending) > 0 {
		keys := make([]string, 0, len(r.pending))
		for d := range r.pending {
			keys = append(keys, d)
		}
		sort.Strings(keys)
		fmt.Println("  waiting for a package to exist:")
		for _, d := range keys {
			fmt.Printf("    %-12s %3d types\n", d, r.pending[d])
		}
	}
	if len(r.collisions) > 0 {
		fmt.Printf("  not generated, a hand-written type already owns the name (%d):\n", len(r.collisions))
		for _, c := range r.collisions {
			fmt.Println("    ", c)
		}
	}
	if len(r.unresolved) > 0 {
		fmt.Printf("  UNRESOLVED, no domain (%d):\n", len(r.unresolved))
		for _, u := range r.unresolved {
			fmt.Println("    ", u)
		}
	}
}
