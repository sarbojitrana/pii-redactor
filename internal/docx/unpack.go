// Package docx handles the .docx-as-zip mechanics: unpacking a source file
// to a working directory, and (in repack.go) rezipping a modified working
// directory back into a valid .docx. It does not know anything about PII —
// that's the detect/mapper/redactor packages' job. This package's only
// concern is getting bytes in and out of the zip container safely.
package docx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DocumentXMLPath is the fixed location of the main document body inside
// every .docx archive.
const DocumentXMLPath = "word/document.xml"

// Unpacked represents a .docx that has been extracted to a working
// directory on disk, ready for its document.xml to be parsed and mutated.
type Unpacked struct {
	// Dir is the working directory containing the extracted archive
	// contents (word/document.xml, word/_rels/, [Content_Types].xml, etc.)
	Dir string
}

// DocumentXML returns the absolute path to word/document.xml within the
// unpacked working directory.
func (u Unpacked) DocumentXML() string {
	return filepath.Join(u.Dir, filepath.FromSlash(DocumentXMLPath))
}

// Unpack extracts the .docx at srcPath into destDir, which must not already
// exist (caller decides whether that's a temp dir or a fixed working dir —
// this function won't silently overwrite one).
//
// It validates that the archive actually contains word/document.xml before
// returning success, so callers get a clear error immediately rather than a
// confusing failure three stages later in the pipeline.
func Unpack(srcPath, destDir string) (Unpacked, error) {
	if _, err := os.Stat(destDir); err == nil {
		return Unpacked{}, fmt.Errorf("unpack: destination %q already exists", destDir)
	}

	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return Unpacked{}, fmt.Errorf("unpack: opening %q as zip: %w", srcPath, err)
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return Unpacked{}, fmt.Errorf("unpack: creating %q: %w", destDir, err)
	}

	sawDocumentXML := false

	for _, f := range r.File {
		// Guard against zip-slip: a malicious or corrupt archive entry
		// that tries to escape destDir via "../" path segments.
		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return Unpacked{}, fmt.Errorf("unpack: unsafe archive entry %q", f.Name)
		}

		targetPath := filepath.Join(destDir, cleanName)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return Unpacked{}, fmt.Errorf("unpack: creating dir %q: %w", targetPath, err)
			}
			continue
		}

		if err := extractFile(f, targetPath); err != nil {
			return Unpacked{}, fmt.Errorf("unpack: extracting %q: %w", f.Name, err)
		}

		if filepath.ToSlash(cleanName) == DocumentXMLPath {
			sawDocumentXML = true
		}
	}

	if !sawDocumentXML {
		return Unpacked{}, fmt.Errorf("unpack: %q not found in archive — is %q a valid .docx?", DocumentXMLPath, srcPath)
	}

	return Unpacked{Dir: destDir}, nil
}

func extractFile(f *zip.File, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return err
	}

	return nil
}