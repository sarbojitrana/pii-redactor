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
	sort.Slice(all, func(i, j int) bool {
		if all[i].Start != all[j].Start {
			return all[i].Start < all[j].Start
		}
		return all[i].End < all[j].End
	})
	return all
}