// Package testutil provides shared test helpers for gpdf example tests.
package testutil

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
	OutputDir = "_output"
	GoldenDir = "testdata/golden"
)

var UpdateGolden = os.Getenv("UPDATE_GOLDEN") != ""

func WritePDF(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}
	path := OutputDir + "/" + name
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", path, err)
	}
	t.Logf("Written %s (%d bytes)", path, len(data))
}

func AssertValidPDF(t *testing.T, data []byte) {
	t.Helper()
	if len(data) == 0 {
		t.Fatal("Generated PDF is empty")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

// AssertMatchesGolden compares data against the golden file in testdata/golden/.
// When UPDATE_GOLDEN=1, it updates the golden file instead of comparing.
func AssertMatchesGolden(t *testing.T, filename string, data []byte) {
	t.Helper()
	goldenPath := GoldenDir + "/" + filename

	if UpdateGolden {
		if err := os.MkdirAll(GoldenDir, 0755); err != nil {
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

// GeneratePDF calls Generate, validates the output, writes the file,
// and compares against the golden file.
func GeneratePDF(t *testing.T, filename string, doc *template.Document) {
	t.Helper()
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	AssertValidPDF(t, data)
	WritePDF(t, filename, data)
	AssertMatchesGolden(t, filename, data)
}

// TestImagePNG creates a small test PNG image (colored rectangle).
func TestImagePNG(t *testing.T, w, h int, c color.Color) []byte {
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

// TestImagePNGAlpha creates a test PNG image with a checkerboard alpha pattern.
// Even-column pixels use the given color with full opacity; odd-column pixels
// use the color with 50% transparency.
func TestImagePNGAlpha(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	r0, g0, b0, _ := c.RGBA()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.RGBA{R: uint8(r0 >> 8), G: uint8(g0 >> 8), B: uint8(b0 >> 8), A: 255})
			} else {
				img.Set(x, y, color.RGBA{R: uint8(r0 >> 8), G: uint8(g0 >> 8), B: uint8(b0 >> 8), A: 128})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create test alpha PNG: %v", err)
	}
	return buf.Bytes()
}

// TestImagePNGGradientAlpha creates a test PNG image with a horizontal alpha gradient.
// The left edge is fully transparent (A=0) and the right edge is fully opaque (A=255).
func TestImagePNGGradientAlpha(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	r0, g0, b0, _ := c.RGBA()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			a := uint8(255 * x / (w - 1))
			img.Set(x, y, color.RGBA{R: uint8(r0 >> 8), G: uint8(g0 >> 8), B: uint8(b0 >> 8), A: a})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create gradient alpha PNG: %v", err)
	}
	return buf.Bytes()
}

// WriteTestImageFile writes image data to a temporary file and returns the path.
func WriteTestImageFile(t *testing.T, data []byte, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/" + name
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write test image file: %v", err)
	}
	return path
}

// TestImageJPEG creates a small test JPEG image (colored rectangle).
func TestImageJPEG(t *testing.T, w, h int, c color.Color) []byte {
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
