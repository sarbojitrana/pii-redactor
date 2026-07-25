package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
	"github.com/sarbojitrana/pii-redactor/internal/docx"
	"github.com/sarbojitrana/pii-redactor/internal/eval"
)

func main() {
	inPath := flag.String("in", "", "Path to the .docx file to evaluate detectors against")
	groundTruthPath := flag.String("ground-truth", "testdata/ground_truth.json", "Path to ground_truth.json")
	reportOutPath := flag.String("out", "output/eval_report.docx", "Path to write the eval_report.docx")
	flag.Parse()

	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: eval --in <input.docx> [--ground-truth testdata/ground_truth.json] [--out output/eval_report.docx]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	tempDir, err := os.MkdirTemp("", "eval-*")
	if err != nil {
		log.Fatalf("Fatal: failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	unpackDir := filepath.Join(tempDir, "unpacked")
	unpacked, err := docx.Unpack(*inPath, unpackDir)
	if err != nil {
		log.Fatalf("Fatal: unpack failed: %v", err)
	}

	xmlBytes, err := os.ReadFile(unpacked.DocumentXML())
	if err != nil {
		log.Fatalf("Fatal: failed to read document.xml: %v", err)
	}

	doc, err := docx.ParseDocumentXML(xmlBytes)
	if err != nil {
		log.Fatalf("Fatal: failed to parse XML structure: %v", err)
	}

	var detectors []detect.Detector
	detectors = append(detectors, detect.NewRegexDetectors()...)
	detectors = append(detectors, detect.NewGazetteerDetectors()...)
	detectors = append(detectors, detect.NewAddressDetectors()...)

	var allMatches []detect.Match
	for _, para := range doc.Paragraphs() {
		text := para.Text()
		println(para.Text())
		if text == "" {
			continue
		}
		allMatches = append(allMatches, detect.RunAll(detectors, text)...)
	}

	report, err := eval.Evaluate(*groundTruthPath, allMatches)
	if err != nil {
		log.Fatalf("Fatal: evaluation failed: %v", err)
	}

	fmt.Println("Category            Precision  Recall  F1      TP  FP  FN")
	for cat, m := range report {
		fmt.Printf("%-20s %.2f       %.2f    %.2f    %d   %d   %d\n",
			cat, m.Precision, m.Recall, m.F1Score, m.TruePositives, m.FalsePositives, m.FalseNegatives)
	}

	if err := eval.GenerateDocx(report, *reportOutPath); err != nil {
		log.Fatalf("Fatal: failed to generate eval_report.docx: %v", err)
	}

	fmt.Printf("\nEval report written to %s\n", *reportOutPath)
}