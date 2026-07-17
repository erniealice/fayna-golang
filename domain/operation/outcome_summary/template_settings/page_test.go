package template_settings

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"
)

// makeZip builds an in-memory ZIP from name→content entries. Entry order is not
// guaranteed (map iteration), which is fine for these structural assertions.
func makeZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %q: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("zip write %q: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// minimalDocx is the smallest archive that passes validateDocxArchive: both
// mandatory OOXML parts, safe paths, well under every cap.
func minimalDocx(t *testing.T) []byte {
	t.Helper()
	return makeZip(t, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0"?><Types/>`,
		"_rels/.rels":         `<?xml version="1.0"?><Relationships/>`,
		"word/document.xml":   `<?xml version="1.0"?><w:document/>`,
	})
}

func TestValidateDocxArchive_Valid(t *testing.T) {
	if err := validateDocxArchive(minimalDocx(t)); err != nil {
		t.Fatalf("expected a well-formed docx to pass, got: %v", err)
	}
}

func TestValidateDocxArchive_NotAZip(t *testing.T) {
	if err := validateDocxArchive([]byte("this is definitely not a zip archive")); err == nil {
		t.Fatal("expected non-zip bytes to be rejected")
	}
}

func TestValidateDocxArchive_Empty(t *testing.T) {
	if err := validateDocxArchive(nil); err == nil {
		t.Fatal("expected empty bytes to be rejected")
	}
}

func TestValidateDocxArchive_MissingDocument(t *testing.T) {
	z := makeZip(t, map[string]string{
		"[Content_Types].xml": `<Types/>`,
		"word/styles.xml":     `<styles/>`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected an archive missing word/document.xml to be rejected")
	}
}

func TestValidateDocxArchive_MissingContentTypes(t *testing.T) {
	z := makeZip(t, map[string]string{
		"word/document.xml": `<w:document/>`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected an archive missing [Content_Types].xml to be rejected")
	}
}

func TestValidateDocxArchive_PathTraversal(t *testing.T) {
	z := makeZip(t, map[string]string{
		"[Content_Types].xml":  `<Types/>`,
		"word/document.xml":    `<w:document/>`,
		"../../etc/passwd.xml": `nope`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected a '..' traversal entry to be rejected")
	}
}

func TestValidateDocxArchive_AbsolutePath(t *testing.T) {
	z := makeZip(t, map[string]string{
		"[Content_Types].xml": `<Types/>`,
		"word/document.xml":   `<w:document/>`,
		"/abs/evil.xml":       `nope`,
	})
	if err := validateDocxArchive(z); err == nil {
		t.Fatal("expected an absolute-path entry to be rejected")
	}
}

func TestValidateDocxArchive_TooManyEntries(t *testing.T) {
	entries := map[string]string{
		"[Content_Types].xml": `<Types/>`,
		"word/document.xml":   `<w:document/>`,
	}
	for i := 0; i < maxArchiveEntries+1; i++ {
		entries[fmt.Sprintf("word/media/e%d.bin", i)] = "x"
	}
	if err := validateDocxArchive(makeZip(t, entries)); err == nil {
		t.Fatalf("expected an archive with > %d entries to be rejected", maxArchiveEntries)
	}
}
