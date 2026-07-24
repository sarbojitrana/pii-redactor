package docx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnpack_RealProspectus(t *testing.T) {
	src := filepath.Join("..", "..", "input", "Red Herring Prospectus.merged.docx")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture not present at %s (run `make prep` first): %v", src, err)
	}

	dest := t.TempDir() // auto-cleaned; also lets Unpack's "must not exist" check pass cleanly
	dest = filepath.Join(dest, "unpacked")

	u, err := Unpack(src, dest)
	if err != nil {
		t.Fatalf("Unpack returned error: %v", err)
	}

	docXML := u.DocumentXML()
	info, err := os.Stat(docXML)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", docXML, err)
	}
	if info.Size() == 0 {
		t.Fatalf("%s is empty", docXML)
	}
}

func TestUnpack_RejectsExistingDestDir(t *testing.T) {
	destDir := t.TempDir() // already exists
	_, err := Unpack("irrelevant.docx", destDir)
	if err == nil {
		t.Fatal("expected error when destDir already exists, got nil")
	}
}

func TestUnpack_RejectsNonDocx(t *testing.T) {
	badFile := filepath.Join(t.TempDir(), "not-a-zip.docx")
	if err := os.WriteFile(badFile, []byte("plainly not a zip archive"), 0o644); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "unpacked")
	_, err := Unpack(badFile, dest)
	if err == nil {
		t.Fatal("expected error unpacking a non-zip file, got nil")
	}
}