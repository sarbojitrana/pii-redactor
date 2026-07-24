package detect

import "regexp"

var pinCodePattern = regexp.MustCompile(`\b[1-9]\d{5}\b`)

const addressMaxBackwardChars = 200
const addressMaxForwardChars = 40

type AddressDetector struct{}

func (d AddressDetector) Name() string { return "address_pincode_anchor" }

func (d AddressDetector) Detect(text string) []Match {
	pinIdxs := pinCodePattern.FindAllStringIndex(text, -1)
	if pinIdxs == nil {
		return nil
	}
	var matches []Match
	for _, pin := range pinIdxs {
		start := expandBackward(text, pin[0])
		end := expandForward(text, pin[1])
		matches = append(matches, Match{
			Category:   CategoryAddress,
			Value:      text[start:end],
			Start:      start,
			End:        end,
			Confidence: 0.6,
			Detector:   "address_pincode_anchor",
		})
	}
	return matches
}

func expandBackward(text string, from int) int {
	limit := from - addressMaxBackwardChars
	if limit < 0 {
		limit = 0
	}
	boundary := limit
	for i := from - 1; i >= limit; i-- {
		c := text[i]
		if c == '.' || c == '\n' {
			boundary = i + 1
			break
		}
	}
	for boundary < from && (text[boundary] == ' ' || text[boundary] == ',') {
		boundary++
	}
	return boundary
}

func expandForward(text string, from int) int {
	limit := from + addressMaxForwardChars
	if limit > len(text) {
		limit = len(text)
	}
	boundary := limit
	for i := from; i < limit; i++ {
		c := text[i]
		if c == '.' || c == '\n' {
			boundary = i
			break
		}
	}
	return boundary
}

func NewAddressDetectors() []Detector {
	return []Detector{AddressDetector{}}
}