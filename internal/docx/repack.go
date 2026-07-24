package docx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// Repack takes the original DOCX path, creates a new file at outPath, and streams
// all original contents across. When it encounters word/document.xml, it writes
// the modifiedXML instead of the original text.
func Repack(inPath, outPath string, modifiedXML []byte) error {
	// 1. Open the original docx as a zip archive
	r, err := zip.OpenReader(inPath)
	if err != nil {
		return fmt.Errorf("failed to open original docx: %w", err)
	}
	defer r.Close()

	// 2. Create the output file
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// 3. Create a new zip writer for the output docx
	w := zip.NewWriter(outFile)
	
	// Ensure the zip writer flushes and closes properly
	defer func() {
		if err := w.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close zip writer: %v\n", err)
		}
	}()

	// 4. Iterate through every file in the original zip
	for _, f := range r.File {
		// Construct a fresh header to avoid carrying over old CRC32s or file sizes
		// from the original struct, which would corrupt the zip if our new XML is a different length.
		header := &zip.FileHeader{
			Name:     f.Name,
			Method:   f.Method,   // Crucial: Preserves compression (e.g., Store vs Deflate)
			Modified: f.Modified, // Preserves original timestamps
		}

		writer, err := w.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", f.Name, err)
		}

		// 5. Intercept the core document XML
		if f.Name == "word/document.xml" {
			if _, err := writer.Write(modifiedXML); err != nil {
				return fmt.Errorf("failed to write modified XML: %w", err)
			}
			continue
		}

		// 6. Stream all other files (styles, rels, images) directly
		srcFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open original file %s: %w", f.Name, err)
		}

		if _, err := io.Copy(writer, srcFile); err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to copy content of %s: %w", f.Name, err)
		}
		
		srcFile.Close()
	}

	return nil
}