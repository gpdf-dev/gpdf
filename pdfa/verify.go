package pdfa

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// Conformance holds parsed PDF/A conformance information from a PDF.
type Conformance struct {
	Part              int    // 1 or 2 (pdfaid:part)
	Level             string // "B" or "A" (pdfaid:conformance)
	XMP               *XMPInfo
	HasOutputIntents  bool
	HasGTSPDFA1       bool
	HasConditionID    bool
	HasDestProfile    bool
	HasMetadataStream bool
	ICCInfo           *ICCProfileInfo
	ForbiddenElements []string // list of forbidden elements found
	FontsEmbedded     bool     // true if font embedding indicators are found
}

// XMPInfo holds parsed XMP metadata values.
type XMPInfo struct {
	Title       string
	Creator     string // dc:creator (author)
	Subject     string
	CreatorTool string // xmp:CreatorTool
	Producer    string // pdf:Producer
	CreateDate  string
	ModifyDate  string
	PDFAPart    int
	PDFAConf    string
	HasXPacket  bool // xpacket begin/end present
}

// ICCProfileInfo holds parsed ICC profile header information.
type ICCProfileInfo struct {
	Size         uint32
	Version      string // e.g. "2.1.0"
	DeviceClass  string // e.g. "mntr"
	ColorSpace   string // e.g. "RGB "
	HasACSP      bool   // "acsp" file signature
	TagCount     int
	RequiredTags []string // list of found required tags
	MissingTags  []string // list of missing required tags
}

// ParseConformance extracts and validates PDF/A conformance information from raw PDF bytes.
// It never returns an error for missing elements; instead, missing elements are reported
// as violations via the Validate method.
func ParseConformance(pdfData []byte) *Conformance {
	s := string(pdfData)
	c := &Conformance{}

	// 1. Parse XMP metadata (may be absent in non-conformant PDFs)
	xmp, err := parseXMP(s)
	if err == nil {
		c.XMP = xmp
		c.Part = xmp.PDFAPart
		c.Level = xmp.PDFAConf
	}

	// 2. OutputIntent checks
	c.HasOutputIntents = strings.Contains(s, "/OutputIntents")
	c.HasGTSPDFA1 = strings.Contains(s, "/GTS_PDFA1")
	c.HasConditionID = strings.Contains(s, "/OutputConditionIdentifier")
	c.HasDestProfile = strings.Contains(s, "/DestOutputProfile")
	c.HasMetadataStream = strings.Contains(s, "/Subtype /XML")

	// 3. ICC profile validation
	c.ICCInfo = parseICCProfile(pdfData)

	// 4. Forbidden elements check
	c.ForbiddenElements = checkForbiddenElements(s, c.Part)

	// 5. Font embedding check
	c.FontsEmbedded = checkFontEmbedding(s)

	return c
}

// Validate checks all PDF/A conformance requirements and returns a list of violations.
func (c *Conformance) Validate() []string {
	var violations []string

	// XMP checks
	if c.XMP == nil {
		violations = append(violations, "XMP metadata not found")
	} else {
		if !c.XMP.HasXPacket {
			violations = append(violations, "XMP missing xpacket header/footer")
		}
		if c.Part == 0 {
			violations = append(violations, "pdfaid:part not found or invalid")
		}
		if c.Level == "" {
			violations = append(violations, "pdfaid:conformance not found")
		}
		if c.XMP.Title == "" {
			violations = append(violations, "XMP dc:title is empty (recommended)")
		}
		if c.XMP.Producer == "" {
			violations = append(violations, "XMP pdf:Producer is empty")
		}
		if c.XMP.CreatorTool == "" {
			violations = append(violations, "XMP xmp:CreatorTool is empty")
		}
		if c.XMP.CreateDate == "" {
			violations = append(violations, "XMP xmp:CreateDate is empty")
		}
		if c.XMP.ModifyDate == "" {
			violations = append(violations, "XMP xmp:ModifyDate is empty")
		}
	}

	// OutputIntent checks
	if !c.HasOutputIntents {
		violations = append(violations, "/OutputIntents missing from catalog")
	}
	if !c.HasGTSPDFA1 {
		violations = append(violations, "/GTS_PDFA1 output intent subtype missing")
	}
	if !c.HasConditionID {
		violations = append(violations, "/OutputConditionIdentifier missing")
	}
	if !c.HasDestProfile {
		violations = append(violations, "/DestOutputProfile missing")
	}
	if !c.HasMetadataStream {
		violations = append(violations, "Metadata stream (/Subtype /XML) missing")
	}

	// ICC profile checks
	if c.ICCInfo == nil {
		violations = append(violations, "ICC color profile not found")
	} else {
		if !c.ICCInfo.HasACSP {
			violations = append(violations, "ICC profile missing 'acsp' signature")
		}
		if c.ICCInfo.ColorSpace != "RGB " {
			violations = append(violations, fmt.Sprintf("ICC color space = %q, want 'RGB '", c.ICCInfo.ColorSpace))
		}
		if c.ICCInfo.DeviceClass != "mntr" {
			violations = append(violations, fmt.Sprintf("ICC device class = %q, want 'mntr'", c.ICCInfo.DeviceClass))
		}
		if c.Part == 1 && !strings.HasPrefix(c.ICCInfo.Version, "2.") {
			violations = append(violations, fmt.Sprintf("PDF/A-1b requires ICC v2.x, got %s", c.ICCInfo.Version))
		}
		for _, tag := range c.ICCInfo.MissingTags {
			violations = append(violations, fmt.Sprintf("ICC profile missing required tag: %s", tag))
		}
	}

	// Forbidden elements
	for _, elem := range c.ForbiddenElements {
		violations = append(violations, fmt.Sprintf("forbidden element found: %s", elem))
	}

	return violations
}

// --- XMP parsing ---

// xmpRDF is a minimal struct to parse the RDF inside XMP.
type xmpRDF struct {
	Descriptions []xmpDescription `xml:"Description"`
}

type xmpDescription struct {
	// Dublin Core
	Title       *xmpAlt `xml:"title"`
	Creator     *xmpSeq `xml:"creator"`
	Description *xmpAlt `xml:"description"`
	// XMP Basic
	CreatorTool string `xml:"CreatorTool"`
	CreateDate  string `xml:"CreateDate"`
	ModifyDate  string `xml:"ModifyDate"`
	// PDF properties
	Producer string `xml:"Producer"`
	// PDF/A identification
	Part        int    `xml:"part"`
	Conformance string `xml:"conformance"`
}

type xmpAlt struct {
	Items []xmpLI `xml:"Alt>li"`
}

type xmpSeq struct {
	Items []xmpLI `xml:"Seq>li"`
}

type xmpLI struct {
	Value string `xml:",chardata"`
}

func parseXMP(pdfStr string) (*XMPInfo, error) {
	info := &XMPInfo{}

	// Find xpacket boundaries
	beginIdx := strings.Index(pdfStr, "<?xpacket begin=")
	endIdx := strings.Index(pdfStr, "<?xpacket end=")
	if beginIdx < 0 || endIdx < 0 {
		return nil, fmt.Errorf("xpacket boundaries not found")
	}
	info.HasXPacket = true

	// Extract XMP between xpacket markers
	xmpRegion := pdfStr[beginIdx : endIdx+len("<?xpacket end=\"w\"?>")]

	// Find <rdf:RDF ...> ... </rdf:RDF>
	rdfStart := strings.Index(xmpRegion, "<rdf:RDF")
	rdfEnd := strings.Index(xmpRegion, "</rdf:RDF>")
	if rdfStart < 0 || rdfEnd < 0 {
		return nil, fmt.Errorf("rdf:RDF not found in XMP")
	}
	rdfXML := xmpRegion[rdfStart : rdfEnd+len("</rdf:RDF>")]

	// Parse with xml.Decoder
	var rdf xmpRDF
	decoder := xml.NewDecoder(strings.NewReader(rdfXML))
	if err := decoder.Decode(&rdf); err != nil {
		return nil, fmt.Errorf("XML decode: %w", err)
	}

	// Extract values from all Description elements
	for _, desc := range rdf.Descriptions {
		if desc.Title != nil && len(desc.Title.Items) > 0 && desc.Title.Items[0].Value != "" {
			info.Title = desc.Title.Items[0].Value
		}
		if desc.Creator != nil && len(desc.Creator.Items) > 0 && desc.Creator.Items[0].Value != "" {
			info.Creator = desc.Creator.Items[0].Value
		}
		if desc.Description != nil && len(desc.Description.Items) > 0 && desc.Description.Items[0].Value != "" {
			info.Subject = desc.Description.Items[0].Value
		}
		if desc.CreatorTool != "" {
			info.CreatorTool = desc.CreatorTool
		}
		if desc.CreateDate != "" {
			info.CreateDate = desc.CreateDate
		}
		if desc.ModifyDate != "" {
			info.ModifyDate = desc.ModifyDate
		}
		if desc.Producer != "" {
			info.Producer = desc.Producer
		}
		if desc.Part > 0 {
			info.PDFAPart = desc.Part
		}
		if desc.Conformance != "" {
			info.PDFAConf = desc.Conformance
		}
	}

	return info, nil
}

// --- ICC profile parsing ---

func parseICCProfile(pdfData []byte) *ICCProfileInfo {
	// Find "acsp" signature at offset 36 in ICC header.
	// Search for the pattern in the PDF stream.
	for i := 36; i < len(pdfData)-128; i++ {
		if string(pdfData[i:i+4]) != "acsp" {
			continue
		}
		// Potential ICC header starts at i-36
		headerStart := i - 36
		if headerStart < 0 {
			continue
		}
		header := pdfData[headerStart:]
		if len(header) < 132 {
			continue
		}

		profileSize := binary.BigEndian.Uint32(header[0:4])
		if profileSize < 128 || profileSize > 1<<20 { // sanity check: < 1MB
			continue
		}

		// Verify color space is "RGB " (we're looking for sRGB)
		colorSpace := string(header[16:20])
		if colorSpace != "RGB " {
			continue
		}

		info := &ICCProfileInfo{
			Size:        profileSize,
			DeviceClass: string(header[12:16]),
			ColorSpace:  colorSpace,
			HasACSP:     true,
		}

		// Version: major.minor.bugfix
		vMajor := header[8]
		vMinor := header[9] >> 4
		vBugfix := header[9] & 0x0F
		info.Version = fmt.Sprintf("%d.%d.%d", vMajor, vMinor, vBugfix)

		// Parse tag table (starts at byte 128)
		if int(profileSize) <= len(header) && profileSize >= 132 {
			tagCount := int(binary.BigEndian.Uint32(header[128:132]))
			info.TagCount = tagCount

			foundTags := make(map[string]bool)
			for t := 0; t < tagCount && 132+t*12+12 <= int(profileSize); t++ {
				offset := 132 + t*12
				sig := string(header[offset : offset+4])
				foundTags[sig] = true
			}

			requiredTags := []string{"desc", "wtpt", "rXYZ", "gXYZ", "bXYZ", "rTRC", "gTRC", "bTRC", "cprt"}
			for _, tag := range requiredTags {
				if foundTags[tag] {
					info.RequiredTags = append(info.RequiredTags, tag)
				} else {
					info.MissingTags = append(info.MissingTags, tag)
				}
			}
		}

		return info
	}

	return nil
}

// --- Forbidden elements check ---

func checkForbiddenElements(pdfStr string, part int) []string {
	// Elements forbidden in all PDF/A levels
	forbidden := []struct {
		pattern string
		name    string
		level   int // 0 = all levels, 1 = PDF/A-1 only
	}{
		{"/JavaScript", "JavaScript action", 0},
		{"/JS ", "JS action shorthand", 0},
		{"/Launch", "Launch action", 0},
		{"/Sound", "Sound action", 0},
		{"/Movie", "Movie action", 0},
		{"/RichMedia", "RichMedia annotation", 0},
		{"/EmbeddedFiles", "EmbeddedFiles (forbidden in PDF/A-1b)", 1},
	}

	var found []string
	for _, f := range forbidden {
		if f.level > 0 && part != f.level {
			continue
		}
		if strings.Contains(pdfStr, f.pattern) {
			found = append(found, f.name)
		}
	}
	return found
}

// --- Font embedding check ---

var reFontFile = regexp.MustCompile(`/FontFile[23]?\s`)

func checkFontEmbedding(pdfStr string) bool {
	// Check for presence of font file references.
	// In a PDF/A-compliant document, fonts should be embedded.
	// We check for /FontFile, /FontFile2, or /FontFile3 references.
	//
	// Note: a PDF with no text content may have no fonts at all.
	// We return true if either fonts are embedded or no fonts are referenced.

	hasFont := strings.Contains(pdfStr, "/Type /Font")
	if !hasFont {
		return true // no fonts = OK
	}

	return reFontFile.MatchString(pdfStr)
}
