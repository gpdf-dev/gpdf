package examples_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"github.com/gpdf-dev/gpdf/template"
)

const (
	outputDir = "_output"
	goldenDir = "testdata/golden"
)

var updateGolden = os.Getenv("UPDATE_GOLDEN") != ""

func writePDF(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}
	path := outputDir + "/" + name
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", path, err)
	}
	t.Logf("Written %s (%d bytes)", path, len(data))
}

func assertValidPDF(t *testing.T, data []byte) {
	t.Helper()
	if len(data) == 0 {
		t.Fatal("Generated PDF is empty")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

// assertMatchesGolden compares data against the golden file in testdata/golden/.
// When UPDATE_GOLDEN=1, it updates the golden file instead of comparing.
func assertMatchesGolden(t *testing.T, filename string, data []byte) {
	t.Helper()
	goldenPath := goldenDir + "/" + filename

	if updateGolden {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatalf("Failed to create golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, data, 0644); err != nil {
			t.Fatalf("Failed to update golden file %s: %v", goldenPath, err)
		}
		t.Logf("Updated golden file %s (%d bytes)", goldenPath, len(data))
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file %s (run with UPDATE_GOLDEN=1 to create): %v", goldenPath, err)
	}
	if !bytes.Equal(data, golden) {
		t.Errorf("Output does not match golden file %s (got %d bytes, want %d bytes; run with UPDATE_GOLDEN=1 to update)", goldenPath, len(data), len(golden))
	}
}

// generatePDF is a helper that calls Generate, validates the output, writes the file,
// and compares against the golden file.
func generatePDF(t *testing.T, filename string, doc *template.Document) {
	t.Helper()
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertValidPDF(t, data)
	writePDF(t, filename, data)
	assertMatchesGolden(t, filename, data)
}

// testImagePNG creates a small test PNG image (colored rectangle).
func testImagePNG(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create test PNG: %v", err)
	}
	return buf.Bytes()
}

// testImageJPEG creates a small test JPEG image (colored rectangle).
func testImageJPEG(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("Failed to create test JPEG: %v", err)
	}
	return buf.Bytes()
}
