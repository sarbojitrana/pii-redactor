package eval

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
)

// GroundTruth represents a manually labeled PII span from testdata/ground_truth.json.
type GroundTruth struct {
	Category string `json:"category"`
	Value    string `json:"value"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
}

type Metrics struct {
	TruePositives  int
	FalsePositives int
	FalseNegatives int
	Precision      float64
	Recall         float64
	F1Score        float64
}

type Report map[string]*Metrics

// Evaluate compares detected matches against a JSON file of ground-truth spans.
func Evaluate(groundTruthPath string, actualMatches []detect.Match) (Report, error) {
	data, err := os.ReadFile(groundTruthPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ground truth: %w", err)
	}

	var expected []GroundTruth
	if err := json.Unmarshal(data, &expected); err != nil {
		return nil, fmt.Errorf("failed to parse ground truth JSON: %w", err)
	}

	report := make(Report)

	// Pre-populate report keys
	for _, e := range expected {
		if _, ok := report[e.Category]; !ok {
			report[e.Category] = &Metrics{}
		}
	}
	for _, a := range actualMatches {
		cat := string(a.Category)
		if _, ok := report[cat]; !ok {
			report[cat] = &Metrics{}
		}
	}

	// Calculate True Positives and False Negatives
	for _, exp := range expected {
		found := false
		for _, act := range actualMatches {
			if string(act.Category) == exp.Category && overlaps(exp.Start, exp.End, act.Start, act.End) {
				found = true
				break
			}
		}
		if found {
			report[exp.Category].TruePositives++
		} else {
			report[exp.Category].FalseNegatives++
		}
	}

	// Calculate False Positives
	for _, act := range actualMatches {
		valid := false
		for _, exp := range expected {
			if string(act.Category) == exp.Category && overlaps(exp.Start, exp.End, act.Start, act.End) {
				valid = true
				break
			}
		}
		if !valid {
			report[string(act.Category)].FalsePositives++
		}
	}

	// Compute Precision, Recall, and F1
	for _, m := range report {
		if m.TruePositives+m.FalsePositives > 0 {
			m.Precision = float64(m.TruePositives) / float64(m.TruePositives+m.FalsePositives)
		}
		if m.TruePositives+m.FalseNegatives > 0 {
			m.Recall = float64(m.TruePositives) / float64(m.TruePositives+m.FalseNegatives)
		}
		if m.Precision+m.Recall > 0 {
			m.F1Score = 2 * (m.Precision * m.Recall) / (m.Precision + m.Recall)
		}
	}

	return report, nil
}

// overlaps checks if two spans intersect. This is crucial because exact index matching
// is too brittle (e.g., if a regex captures a trailing space that the human missed).
func overlaps(start1, end1, start2, end2 int) bool {
	return start1 < end2 && start2 < end1
}