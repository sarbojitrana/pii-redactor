package redactor

import (
	"sort"
	"strings"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
	"github.com/sarbojitrana/pii-redactor/internal/docx"
	"github.com/sarbojitrana/pii-redactor/internal/mapper"
)

// Redactor orchestrates the detection and mapping process across a document.
type Redactor struct {
	detectors []detect.Detector
	mapper    *mapper.Mapper 
}

// New creates a new Redactor instance.
func New(detectors []detect.Detector, m *mapper.Mapper) *Redactor {
	return &Redactor{
		detectors: detectors,
		mapper:    m,
	}
}

// ProcessDocument modifies the docx.Document in-place by redacting detected PII.
func (r *Redactor) ProcessDocument(doc *docx.Document) {
	for _, para := range doc.Paragraphs() {
		for _, node := range para.TextNodes {
			r.processNode(node)
		}
	}
}

func (r *Redactor) processNode(node *docx.TextNode) {
	originalText := node.Text()
	if strings.TrimSpace(originalText) == "" {
		return
	}

	// 1. Detect all potential PII spans
	rawMatches := detect.RunAll(r.detectors, originalText)
	if len(rawMatches) == 0 {
		return
	}

	// 2. Resolve overlapping spans (longest match wins)
	resolvedMatches := resolveOverlaps(rawMatches)

	// 3. Sort matches by Start index in ascending order for safe reverse-iteration
	sort.Slice(resolvedMatches, func(i, j int) bool {
		return resolvedMatches[i].Start < resolvedMatches[j].Start
	})

	// 4. Apply replacements from right to left (end to start)
	// This prevents earlier replacements from invalidating the indices of later matches.
	redactedText := originalText
	for i := len(resolvedMatches) - 1; i >= 0; i-- {
		match := resolvedMatches[i]
		
		// Generate deterministic fake value keyed on category and original value
		fakeValue := r.mapper.Map(match)
		
		// Splice the string using the original indices
		redactedText = redactedText[:match.Start] + fakeValue + redactedText[match.End:]
	}

	// 5. Update the XML node safely
	node.SetText(redactedText)
}

// resolveOverlaps implements a "longest match wins" strategy for overlapping spans.
func resolveOverlaps(matches []detect.Match) []detect.Match {
	if len(matches) <= 1 {
		return matches
	}

	// Sort by length descending, then by confidence descending
	sort.Slice(matches, func(i, j int) bool {
		lenI := matches[i].End - matches[i].Start
		lenJ := matches[j].End - matches[j].Start
		if lenI != lenJ {
			return lenI > lenJ
		}
		return matches[i].Confidence > matches[j].Confidence
	})

	var resolved []detect.Match
	
	for _, current := range matches {
		overlap := false
		for _, accepted := range resolved {
			// Check if current match boundaries overlap with any accepted match boundaries
			if current.Start < accepted.End && current.End > accepted.Start {
				overlap = true
				break
			}
		}
		
		// If it doesn't overlap with a longer, already-accepted match, keep it.
		if !overlap {
			resolved = append(resolved, current)
		}
	}

	return resolved
}