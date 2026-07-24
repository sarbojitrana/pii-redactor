package eval

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"text/template"
)

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
    <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
    <Default Extension="xml" ContentType="application/xml"/>
    <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const relsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const docXMLTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>
        <w:p>
            <w:r>
                <w:rPr><w:b/><w:sz w:val="32"/></w:rPr>
                <w:t>PII Redaction Evaluation Report</w:t>
            </w:r>
        </w:p>
        <w:p><w:r><w:t></w:t></w:r></w:p>
        {{range $category, $metrics := .}}
        <w:p>
            <w:r>
                <w:rPr><w:b/></w:rPr>
                <w:t xml:space="preserve">Category: {{$category}}</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t xml:space="preserve">  Precision: {{$metrics.Precision | printf "%.2f"}} | Recall: {{$metrics.Recall | printf "%.2f"}} | F1: {{$metrics.F1Score | printf "%.2f"}}</w:t>
            </w:r>
        </w:p>
        <w:p>
            <w:r>
                <w:t xml:space="preserve">  (True Positives: {{$metrics.TruePositives}}, False Positives: {{$metrics.FalsePositives}}, False Negatives: {{$metrics.FalseNegatives}})</w:t>
            </w:r>
        </w:p>
        <w:p><w:r><w:t></w:t></w:r></w:p>
        {{end}}
    </w:body>
</w:document>`

// GenerateDocx takes the computed evaluation report and packages it into a valid .docx file.
func GenerateDocx(report Report, outPath string) error {
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create report docx: %w", err)
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	// Helper function to write a string directly to a file inside the ZIP archive
	addFile := func(name, content string) error {
		f, err := w.Create(name)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(content))
		return err
	}

	// 1. Write the mandatory OOXML structural files
	if err := addFile("[Content_Types].xml", contentTypesXML); err != nil {
		return err
	}
	if err := addFile("_rels/.rels", relsXML); err != nil {
		return err
	}

	// 2. Execute the Go template to generate the WordprocessingML body
	tmpl, err := template.New("document").Parse(docXMLTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse document template: %w", err)
	}

	var docXMLBuffer bytes.Buffer
	if err := tmpl.Execute(&docXMLBuffer, report); err != nil {
		return fmt.Errorf("failed to execute document template: %w", err)
	}

	// 3. Write the compiled XML payload into the ZIP archive
	if err := addFile("word/document.xml", docXMLBuffer.String()); err != nil {
		return err
	}

	return nil
}