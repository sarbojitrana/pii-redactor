package detect

import (
	"sort"
	"strings"
)

type gazetteerDetector struct {
	name     string
	category Category
	entries  []string
}

func newGazetteerDetector(name string, category Category, entries []string) *gazetteerDetector {
	sorted := make([]string, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) > len(sorted[j])
	})
	return &gazetteerDetector{name: name, category: category, entries: sorted}
}

func (g *gazetteerDetector) Name() string { return g.name }

func (g *gazetteerDetector) Detect(text string) []Match {
	lower := strings.ToLower(text)
	claimed := make([]bool, len(text))
	var matches []Match

	for _, entry := range g.entries {
		entryLower := strings.ToLower(entry)
		if entryLower == "" {
			continue
		}
		searchFrom := 0
		for {
			idx := strings.Index(lower[searchFrom:], entryLower)
			if idx == -1 {
				break
			}
			start := searchFrom + idx
			end := start + len(entry)
			searchFrom = end

			if !hasWordBoundaries(text, start, end) {
				continue
			}
			if rangeClaimed(claimed, start, end) {
				continue
			}
			for i := start; i < end; i++ {
				claimed[i] = true
			}
			matches = append(matches, Match{
				Category:   g.category,
				Value:      text[start:end],
				Start:      start,
				End:        end,
				Confidence: 0.85,
				Detector:   g.name,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].Start < matches[j].Start })
	return matches
}

func hasWordBoundaries(text string, start, end int) bool {
	if start > 0 && isWordChar(text[start-1]) {
		return false
	}
	if end < len(text) && isWordChar(text[end]) {
		return false
	}
	return true
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func rangeClaimed(claimed []bool, start, end int) bool {
	for i := start; i < end; i++ {
		if claimed[i] {
			return true
		}
	}
	return false
}

func NewNameGazetteer(names []string) Detector {
	return newGazetteerDetector("gazetteer_name", CategoryName, names)
}

func NewCompanyGazetteer(companies []string) Detector {
	return newGazetteerDetector("gazetteer_company", CategoryCompany, companies)
}