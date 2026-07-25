package detect

import "sort"

type Category string

const (
	CategoryEmail      Category = "EMAIL"
	CategoryPhone      Category = "PHONE"
	CategorySSN        Category = "SSN"
	CategoryCreditCard Category = "CREDIT_CARD"
	CategoryIP         Category = "IP_ADDRESS"
	CategoryDOB        Category = "DOB"
	CategoryAddress    Category = "ADDRESS"
	CategoryName       Category = "NAME"
	CategoryCompany    Category = "COMPANY"
	CategoryCIN        Category = "CIN"
	CategoryPAN        Category = "PAN"
	CategoryISIN       Category = "ISIN"
)

type Match struct {
	Category   Category
	Value      string
	Start      int
	End        int
	Confidence float64
	Detector   string
}

type Detector interface {
	Name() string
	Detect(text string) []Match
}

func RunAll(detectors []Detector, text string) []Match {
	var all []Match
	for _, d := range detectors {
		all = append(all, d.Detect(text)...)
	}
	return ResolveOverlaps(all)
}

func ResolveOverlaps(matches []Match) []Match {
	byConfidence := make([]Match, len(matches))
	copy(byConfidence, matches)
	sort.Slice(byConfidence, func(i, j int) bool {
		if byConfidence[i].Confidence != byConfidence[j].Confidence {
			return byConfidence[i].Confidence > byConfidence[j].Confidence
		}
		return (byConfidence[i].End - byConfidence[i].Start) > (byConfidence[j].End - byConfidence[j].Start)
	})

	var kept []Match
	for _, m := range byConfidence {
		overlaps := false
		for _, k := range kept {
			if m.Start < k.End && k.Start < m.End {
				overlaps = true
				break
			}
		}
		if !overlaps {
			kept = append(kept, m)
		}
	}

	sort.Slice(kept, func(i, j int) bool {
		if kept[i].Start != kept[j].Start {
			return kept[i].Start < kept[j].Start
		}
		return kept[i].End < kept[j].End
	})
	return kept
}