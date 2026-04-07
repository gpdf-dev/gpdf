package flatten_test

import (
	"os"
	"testing"
)

// TestGenerateTestdata regenerates the testdata PDF files.
// Run with: go test -run TestGenerateTestdata -update-testdata
//
// This is not run in normal test execution.
func TestGenerateTestdata(t *testing.T) {
	if os.Getenv("UPDATE_TESTDATA") == "" {
		t.Skip("set UPDATE_TESTDATA=1 to regenerate testdata")
	}

	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("01_basic_flatten_before", func(t *testing.T) {
		data := buildFormPDF(t)
		if err := os.WriteFile("testdata/01_basic_flatten_before.pdf", data, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote testdata/01_basic_flatten_before.pdf (%d bytes)", len(data))
	})

	t.Run("02_preserve_non_widget_before", func(t *testing.T) {
		data := buildMixedAnnotPDF(t)
		if err := os.WriteFile("testdata/02_preserve_non_widget_before.pdf", data, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote testdata/02_preserve_non_widget_before.pdf (%d bytes)", len(data))
	})

	t.Run("04_filled_form_before", func(t *testing.T) {
		data := buildFilledFormPDF(t)
		if err := os.WriteFile("testdata/04_filled_form_before.pdf", data, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote testdata/04_filled_form_before.pdf (%d bytes)", len(data))
	})
}
