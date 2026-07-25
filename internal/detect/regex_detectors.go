package detect

import "regexp"

var emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

var phonePattern = regexp.MustCompile(`\b(?:\+91[\-\s]?)?[6-9]\d{4}[\-\s]?\d{5}\b`)

var ssnPattern = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)

var creditCardPattern = regexp.MustCompile(`\b\d{4}[\- ]\d{4}[\- ]\d{4}[\- ]\d{4}\b|\b\d{16}\b`)

var ipPattern = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|1?\d?\d)\b`)

var dobPattern = regexp.MustCompile(`\b\d{1,2}[/\-]\d{1,2}[/\-]\d{2,4}\b`)

type EmailDetector struct{}

func (d EmailDetector) Name() string { return "regex_email" }

func (d EmailDetector) Detect(text string) []Match {
	return matchesFromPattern(emailPattern, text, CategoryEmail, d.Name(), 0.95)
}

type PhoneDetector struct{}

func (d PhoneDetector) Name() string { return "regex_phone" }

func (d PhoneDetector) Detect(text string) []Match {
	return matchesFromPattern(phonePattern, text, CategoryPhone, d.Name(), 0.85)
}

type SSNDetector struct{}

func (d SSNDetector) Name() string { return "regex_ssn" }

func (d SSNDetector) Detect(text string) []Match {
	return matchesFromPattern(ssnPattern, text, CategorySSN, d.Name(), 0.9)
}

type CreditCardDetector struct{}

func (d CreditCardDetector) Name() string { return "regex_credit_card" }

func (d CreditCardDetector) Detect(text string) []Match {
	return matchesFromPattern(creditCardPattern, text, CategoryCreditCard, d.Name(), 0.8)
}

type IPDetector struct{}

func (d IPDetector) Name() string { return "regex_ip" }

func (d IPDetector) Detect(text string) []Match {
	return matchesFromPattern(ipPattern, text, CategoryIP, d.Name(), 0.9)
}

type DOBDetector struct{}

func (d DOBDetector) Name() string { return "regex_dob" }

func (d DOBDetector) Detect(text string) []Match {
	return matchesFromPattern(dobPattern, text, CategoryDOB, d.Name(), 0.5)
}

func matchesFromPattern(pattern *regexp.Regexp, text string, category Category, detectorName string, confidence float64) []Match {
	idxs := pattern.FindAllStringIndex(text, -1)
	if idxs == nil {
		return nil
	}
	matches := make([]Match, 0, len(idxs))
	for _, idx := range idxs {
		matches = append(matches, Match{
			Category:   category,
			Value:      text[idx[0]:idx[1]],
			Start:      idx[0],
			End:        idx[1],
			Confidence: confidence,
			Detector:   detectorName,
		})
	}
	return matches
}

func NewRegexDetectors() []Detector {
	return []Detector{
		EmailDetector{},
		PhoneDetector{},
		SSNDetector{},
		CreditCardDetector{},
		IPDetector{},
		DOBDetector{},
	}
}