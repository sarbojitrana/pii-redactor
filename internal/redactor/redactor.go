package redactor

import (
	"sort"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
	"github.com/sarbojitrana/pii-redactor/internal/docx"
	"github.com/sarbojitrana/pii-redactor/internal/mapper"
)

type Redactor struct {
	detectors     []detect.Detector
	mapper        *mapper.Mapper
	MinConfidence float64
}

func New(detectors []detect.Detector, m *mapper.Mapper) *Redactor {
	return &Redactor{
		detectors:     detectors,
		mapper:        m,
		MinConfidence: 0.0,
	}
}

type nodeSpan struct {
	node  *docx.TextNode
	text  string
	start int
	end   int
}

type nodeEdit struct {
	start       int
	end         int
	replacement string
}

func buildSpans(para docx.Paragraph) []nodeSpan {
	offsets := para.Offsets()
	spans := make([]nodeSpan, 0, len(offsets))
	for _, o := range offsets {
		spans = append(spans, nodeSpan{node: o.Node, text: o.Node.Text(), start: o.Start, end: o.End})
	}
	return spans
}

func spansInRange(spans []nodeSpan, start, end int) []nodeSpan {
	var out []nodeSpan
	for _, sp := range spans {
		if sp.end <= start {
			continue
		}
		if sp.start >= end {
			break
		}
		out = append(out, sp)
	}
	return out
}

func (r *Redactor) ProcessDocument(doc *docx.Document) {
	for _, para := range doc.Paragraphs() {
		r.processParagraph(para)
	}
}

func (r *Redactor) processParagraph(para docx.Paragraph) {
	if len(para.TextNodes) == 0 {
		return
	}

	spans := buildSpans(para)
	text := para.Text()

	matches := detect.RunAll(r.detectors, text)
	if len(matches) == 0 {
		return
	}

	edits := make(map[*docx.TextNode][]nodeEdit)

	for _, match := range matches {
		if match.Confidence < r.MinConfidence {
			continue
		}

		affected := spansInRange(spans, match.Start, match.End)
		if len(affected) == 0 {
			continue
		}

		fakeValue := r.mapper.Map(match)

		if len(affected) == 1 {
			sp := affected[0]
			localStart := match.Start - sp.start
			if localStart < 0 {
				localStart = 0
			}
			localEnd := match.End - sp.start
			if localEnd > len(sp.text) {
				localEnd = len(sp.text)
			}
			edits[sp.node] = append(edits[sp.node], nodeEdit{localStart, localEnd, fakeValue})
			continue
		}

		first := affected[0]
		firstLocalStart := match.Start - first.start
		if firstLocalStart < 0 {
			firstLocalStart = 0
		}
		edits[first.node] = append(edits[first.node], nodeEdit{firstLocalStart, len(first.text), fakeValue})

		for _, mid := range affected[1 : len(affected)-1] {
			edits[mid.node] = append(edits[mid.node], nodeEdit{0, len(mid.text), ""})
		}

		last := affected[len(affected)-1]
		lastLocalEnd := match.End - last.start
		if lastLocalEnd > len(last.text) {
			lastLocalEnd = len(last.text)
		}
		if lastLocalEnd < 0 {
			lastLocalEnd = 0
		}
		edits[last.node] = append(edits[last.node], nodeEdit{0, lastLocalEnd, ""})
	}

	for node, nodeEdits := range edits {
		sort.Slice(nodeEdits, func(i, j int) bool {
			return nodeEdits[i].start > nodeEdits[j].start
		})
		current := node.Text()
		for _, e := range nodeEdits {
			current = current[:e.start] + e.replacement + current[e.end:]
		}
		node.SetText(current)
	}
}