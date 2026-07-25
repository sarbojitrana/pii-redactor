package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
)

type GroundTruth struct {
	Category string `json:"category"`
	Value    string `json:"value"`
}

type groundTruthFile struct {
	Entries []GroundTruth `json:"entries"`
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

func normalize(s string) string {
	return strings.TrimSpace(s)
}

func Evaluate(groundTruthPath string, actualMatches []detect.Match) (Report, error) {
	data, err := os.ReadFile(groundTruthPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ground truth: %w", err)
	}

	var gt groundTruthFile
	if err := json.Unmarshal(data, &gt); err != nil {
		return nil, fmt.Errorf("failed to parse ground truth JSON: %w", err)
	}

	expectedSets := make(map[string]map[string]bool)
	for _, e := range gt.Entries {
		if expectedSets[e.Category] == nil {
			expectedSets[e.Category] = make(map[string]bool)
		}
		expectedSets[e.Category][normalize(e.Value)] = true
	}

	actualSets := make(map[string]map[string]bool)
	for _, a := range actualMatches {
		cat := string(a.Category)
		if actualSets[cat] == nil {
			actualSets[cat] = make(map[string]bool)
		}
		actualSets[cat][normalize(a.Value)] = true
	}

	categories := make(map[string]bool)
	for cat := range expectedSets {
		categories[cat] = true
	}
	for cat := range actualSets {
		categories[cat] = true
	}

	report := make(Report)
	for cat := range categories {
		expected := expectedSets[cat]
		actual := actualSets[cat]
		m := &Metrics{}

		for val := range expected {
			if actual[val] {
				m.TruePositives++
			} else {
				m.FalseNegatives++
			}
		}
		for val := range actual {
			if !expected[val] {
				m.FalsePositives++
			}
		}

		if m.TruePositives+m.FalsePositives > 0 {
			m.Precision = float64(m.TruePositives) / float64(m.TruePositives+m.FalsePositives)
		}
		if m.TruePositives+m.FalseNegatives > 0 {
			m.Recall = float64(m.TruePositives) / float64(m.TruePositives+m.FalseNegatives)
		}
		if m.Precision+m.Recall > 0 {
			m.F1Score = 2 * (m.Precision * m.Recall) / (m.Precision + m.Recall)
		}

		report[cat] = m
	}

	return report, nil
}