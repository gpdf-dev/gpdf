package pdfa

import (
	"fmt"
	"strings"
	"time"
)

// generateXMP creates XMP metadata XML for PDF/A conformance.
func generateXMP(level Level, info MetadataInfo) []byte {
	part, conformance := pdfaPartConformance(level)

	createDate := info.CreateDate
	if createDate == "" {
		createDate = time.Now().Format(time.RFC3339)
	}
	modifyDate := info.ModifyDate
	if modifyDate == "" {
		modifyDate = createDate
	}

	var b strings.Builder
	b.WriteString("<?xpacket begin=\"\xEF\xBB\xBF\" id=\"W5M0MpCehiHzreSzNTczkc9d\"?>")
	b.WriteString("\n")
	b.WriteString("<x:xmpmeta xmlns:x=\"adobe:ns:meta/\">")
	b.WriteString("\n")
	b.WriteString("<rdf:RDF xmlns:rdf=\"http://www.w3.org/1999/02/22-rdf-syntax-ns#\">")
	b.WriteString("\n")

	// Dublin Core
	b.WriteString("<rdf:Description rdf:about=\"\"")
	b.WriteString(" xmlns:dc=\"http://purl.org/dc/elements/1.1/\">")
	b.WriteString("\n")
	if info.Title != "" {
		fmt.Fprintf(&b, "<dc:title><rdf:Alt><rdf:li xml:lang=\"x-default\">%s</rdf:li></rdf:Alt></dc:title>", xmlEscape(info.Title))
		b.WriteString("\n")
	}
	if info.Author != "" {
		fmt.Fprintf(&b, "<dc:creator><rdf:Seq><rdf:li>%s</rdf:li></rdf:Seq></dc:creator>", xmlEscape(info.Author))
		b.WriteString("\n")
	}
	if info.Subject != "" {
		fmt.Fprintf(&b, "<dc:description><rdf:Alt><rdf:li xml:lang=\"x-default\">%s</rdf:li></rdf:Alt></dc:description>", xmlEscape(info.Subject))
		b.WriteString("\n")
	}
	b.WriteString("</rdf:Description>\n")

	// XMP Basic
	b.WriteString("<rdf:Description rdf:about=\"\"")
	b.WriteString(" xmlns:xmp=\"http://ns.adobe.com/xap/1.0/\">")
	b.WriteString("\n")
	fmt.Fprintf(&b, "<xmp:CreatorTool>%s</xmp:CreatorTool>", xmlEscape(info.Creator))
	b.WriteString("\n")
	fmt.Fprintf(&b, "<xmp:CreateDate>%s</xmp:CreateDate>", createDate)
	b.WriteString("\n")
	fmt.Fprintf(&b, "<xmp:ModifyDate>%s</xmp:ModifyDate>", modifyDate)
	b.WriteString("\n")
	b.WriteString("</rdf:Description>\n")

	// PDF properties
	b.WriteString("<rdf:Description rdf:about=\"\"")
	b.WriteString(" xmlns:pdf=\"http://ns.adobe.com/pdf/1.3/\">")
	b.WriteString("\n")
	fmt.Fprintf(&b, "<pdf:Producer>%s</pdf:Producer>", xmlEscape(info.Producer))
	b.WriteString("\n")
	b.WriteString("</rdf:Description>\n")

	// PDF/A identification
	b.WriteString("<rdf:Description rdf:about=\"\"")
	b.WriteString(" xmlns:pdfaid=\"http://www.aiim.org/pdfa/ns/id/\">")
	b.WriteString("\n")
	fmt.Fprintf(&b, "<pdfaid:part>%d</pdfaid:part>", part)
	b.WriteString("\n")
	fmt.Fprintf(&b, "<pdfaid:conformance>%s</pdfaid:conformance>", conformance)
	b.WriteString("\n")
	b.WriteString("</rdf:Description>\n")

	b.WriteString("</rdf:RDF>\n")
	b.WriteString("</x:xmpmeta>\n")
	b.WriteString("<?xpacket end=\"w\"?>")

	return []byte(b.String())
}

func pdfaPartConformance(level Level) (int, string) {
	switch level {
	case LevelA2b:
		return 2, "B"
	default: // LevelA1b
		return 1, "B"
	}
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
