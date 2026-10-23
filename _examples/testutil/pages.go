package testutil

import (
	"regexp"
	"strings"
	"testing"
)

var (
	pageContentsRef = regexp.MustCompile(`<< /Contents (\d+) 0 R [^>]*>> >> /Type /Page >>`)
	streamObject    = regexp.MustCompile(`(?s)(?:^|\n)(\d+) 0 obj\n<< /Length \d+ >>\nstream\n(.*?)endstream`)
)

// PageContents returns the uncompressed content stream of each page in
// the generated PDF, in page order. It only understands the plain
// (unfiltered, unencrypted) output produced by gpdf's default writer.
func PageContents(t *testing.T, data []byte) []string {
	t.Helper()
	streams := map[string]string{}
	for _, m := range streamObject.FindAllSubmatch(data, -1) {
		streams[string(m[1])] = string(m[2])
	}
	var pages []string
	for _, m := range pageContentsRef.FindAllSubmatch(data, -1) {
		content, ok := streams[string(m[1])]
		if !ok {
			t.Fatalf("content stream %s 0 obj not found", m[1])
		}
		pages = append(pages, content)
	}
	if len(pages) == 0 {
		t.Fatal("no pages found in PDF")
	}
	return pages
}

// AssertTableSplitsFromFirstPage checks that a table preceded by other
// content starts on the first page (right after intro), that its header
// is repeated on every page, and that every body row is rendered exactly
// once across the whole document.
func AssertTableSplitsFromFirstPage(t *testing.T, data []byte, intro, header string, rows []string) {
	t.Helper()
	pages := PageContents(t, data)
	if len(pages) < 2 {
		t.Fatalf("expected the table to span multiple pages, got %d page(s)", len(pages))
	}
	lit := func(s string) string { return "(" + s + ")" }

	if !strings.Contains(pages[0], lit(intro)) {
		t.Errorf("expected %q on page 1", intro)
	}
	if !strings.Contains(pages[0], lit(rows[0])) {
		t.Errorf("expected first table row %q on page 1 (table must start on the first page)", rows[0])
	}
	for i, p := range pages {
		if got := strings.Count(p, lit(header)); got != 1 {
			t.Errorf("page %d: expected table header %q once, got %d", i+1, header, got)
		}
	}
	all := strings.Join(pages, "\n")
	for _, r := range rows {
		if got := strings.Count(all, lit(r)); got != 1 {
			t.Errorf("expected row %q exactly once, got %d", r, got)
		}
	}
}
