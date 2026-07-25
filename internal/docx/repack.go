package docx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

func Repack(inPath, outPath string, modifiedXML []byte) (err error) {
	r, err := zip.OpenReader(inPath)
	if err != nil {
		return fmt.Errorf("failed to open original docx: %w", err)
	}
	defer r.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer func() {
		closeErr := w.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close zip writer: %w", closeErr)
		}
	}()

	sawDocumentXML := false

	for _, f := range r.File {
		header := &zip.FileHeader{
			Name:     f.Name,
			Method:   f.Method,
			Modified: f.Modified,
		}

		writer, headerErr := w.CreateHeader(header)
		if headerErr != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", f.Name, headerErr)
		}

		if f.Name == "word/document.xml" {
			if _, writeErr := writer.Write(modifiedXML); writeErr != nil {
				return fmt.Errorf("failed to write modified XML: %w", writeErr)
			}
			sawDocumentXML = true
			continue
		}

		srcFile, openErr := f.Open()
		if openErr != nil {
			return fmt.Errorf("failed to open original file %s: %w", f.Name, openErr)
		}

		_, copyErr := io.Copy(writer, srcFile)
		srcFile.Close()
		if copyErr != nil {
			return fmt.Errorf("failed to copy content of %s: %w", f.Name, copyErr)
		}
	}

	if !sawDocumentXML {
		return fmt.Errorf("repack: %q not found in original archive %q — refusing to write a copy with no redactions applied", "word/document.xml", inPath)
	}

	return nil
}