package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
	"github.com/sarbojitrana/pii-redactor/internal/docx"
	"github.com/sarbojitrana/pii-redactor/internal/mapper"
	"github.com/sarbojitrana/pii-redactor/internal/redactor"
)

func main() {
	// 1. Parse CLI Flags
	inPath := flag.String("in", "", "Path to the input .docx file (e.g., input/Red_Herring_Prospectus.docx)")
	outPath := flag.String("out", "", "Path to save the redacted .docx file")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: redact --in <input.docx> --out <output.docx>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	start := time.Now()
	log.Printf("Starting redaction pipeline on %s...", *inPath)

	// Inside cmd/redact/main.go (Step 2)
	
	// Create a temporary directory for unpacking
	tempDir, err := os.MkdirTemp("", "redactor-*")
	if err != nil {
		log.Fatalf("Fatal: failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up when done

	// Unpack the DOCX to disk
	unpacked, err := docx.Unpack(*inPath, tempDir)
	if err != nil {
		log.Fatalf("Fatal: unpack failed: %v", err)
	}

	// Read the extracted document.xml into memory
	xmlBytes, err := os.ReadFile(unpacked.DocumentXML())
	if err != nil {
		log.Fatalf("Fatal: failed to read document.xml: %v", err)
	}

	// 3. Parse XML into text nodes
	doc, err := docx.ParseDocumentXML(xmlBytes)
	if err != nil {
		log.Fatalf("Fatal: failed to parse XML structure: %v", err)
	}

	// 4. Initialize Detectors using the slice-based constructors
	var detectors []detect.Detector
	detectors = append(detectors, detect.NewRegexDetectors()...)
	detectors = append(detectors, detect.NewGazetteerDetectors()...)
	detectors = append(detectors, detect.NewAddressDetectors()...)

	// 5. Initialize Mapper (Deterministic Fake Generator)
	m := mapper.NewMapper()

	// 6. Initialize Orchestrator and Process
	r := redactor.New(detectors, m)
	r.ProcessDocument(doc)

	// 7. Serialize Mutated XML
	modifiedXML, err := doc.Serialize()
	if err != nil {
		log.Fatalf("Fatal: failed to serialize mutated XML: %v", err)
	}

	// 8. Repack into new valid DOCX
	err = docx.Repack(*inPath, *outPath, modifiedXML)
	if err != nil {
		log.Fatalf("Fatal: failed to repack docx to %s: %v", *outPath, err)
	}

	log.Printf("Successfully redacted document in %v. Saved to %s", time.Since(start), *outPath)
}