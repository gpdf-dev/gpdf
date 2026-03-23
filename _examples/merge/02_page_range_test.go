package merge_test

import (
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
)

func TestExample_Merge_02_PageRange(t *testing.T) {
	// Generate a 5-page document.
	full := generateDoc(t, 5, "Full Document")

	// Extract only pages 2-4 (1-based).
	extracted, err := gpdf.Merge([]gpdf.Source{
		{Data: full, Pages: gpdf.PageRange{From: 2, To: 4}},
	})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	testutil.AssertValidPDF(t, extracted)
	testutil.WritePDF(t, "02_page_range.pdf", extracted)
}
