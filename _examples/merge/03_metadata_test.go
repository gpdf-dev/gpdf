package merge_test

import (
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
)

func TestExample_Merge_03_Metadata(t *testing.T) {
	cover := generateDoc(t, 1, "Cover")
	body := generateDoc(t, 2, "Body")
	appendix := generateDoc(t, 1, "Appendix")

	// Merge with custom metadata.
	merged, err := gpdf.Merge(
		[]gpdf.Source{
			{Data: cover},
			{Data: body},
			{Data: appendix},
		},
		gpdf.WithMergeMetadata("Policy Bundle", "Example Ltd", "gpdf"),
	)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	testutil.AssertValidPDF(t, merged)
	testutil.WritePDF(t, "03_metadata.pdf", merged)
}
