// Package pdfa provides PDF/A conformance support for gpdf.
// It configures a pdf.Writer to produce PDF/A-1b or PDF/A-2b compliant output
// by injecting ICC color profiles, XMP metadata, and OutputIntent dictionaries.
package pdfa

import (
	"github.com/gpdf-dev/gpdf/pdf"
)

// Level represents a PDF/A conformance level.
type Level int

const (
	LevelA1b Level = iota // PDF/A-1b (ISO 19005-1, Level B)
	LevelA2b              // PDF/A-2b (ISO 19005-2, Level B)
)

// MetadataInfo holds document metadata for PDF/A XMP.
type MetadataInfo struct {
	Title      string
	Author     string
	Subject    string
	Creator    string
	Producer   string
	CreateDate string // ISO 8601 format, e.g. "2024-01-15T10:30:00+09:00"
	ModifyDate string // ISO 8601 format
}

// Option configures PDF/A generation.
type Option func(*config)

type config struct {
	level    Level
	metadata MetadataInfo
}

// WithLevel sets the PDF/A conformance level.
func WithLevel(level Level) Option {
	return func(c *config) { c.level = level }
}

// WithMetadata sets the document metadata for the XMP stream.
func WithMetadata(info MetadataInfo) Option {
	return func(c *config) { c.metadata = info }
}

// Apply configures a pdf.Writer for PDF/A conformance.
// It registers beforeClose hooks to write the ICC profile and XMP metadata,
// and adds OutputIntents and Metadata entries to the catalog.
func Apply(pw *pdf.Writer, opts ...Option) {
	cfg := &config{
		level: LevelA1b,
		metadata: MetadataInfo{
			Producer: "gpdf",
			Creator:  "gpdf",
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	pw.OnBeforeClose(func(pw *pdf.Writer) error {
		// 1. Write ICC profile stream.
		iccRef := pw.AllocObject()
		iccData := sRGBICCProfile()
		iccStream := pdf.Stream{
			Dict: pdf.Dict{
				pdf.Name("N"):         pdf.Integer(3),
				pdf.Name("Alternate"): pdf.Name("DeviceRGB"),
			},
			Content: iccData,
		}
		if err := pw.WriteObject(iccRef, iccStream); err != nil {
			return err
		}

		// 2. Write OutputIntent dictionary.
		outputIntentRef := pw.AllocObject()
		outputIntentDict := pdf.Dict{
			pdf.Name("Type"):                      pdf.Name("OutputIntent"),
			pdf.Name("S"):                         pdf.Name("GTS_PDFA1"),
			pdf.Name("OutputConditionIdentifier"): pdf.LiteralString("sRGB IEC61966-2.1"),
			pdf.Name("RegistryName"):              pdf.LiteralString("http://www.color.org"),
			pdf.Name("Info"):                      pdf.LiteralString("sRGB IEC61966-2.1"),
			pdf.Name("DestOutputProfile"):         iccRef,
		}
		if err := pw.WriteObject(outputIntentRef, outputIntentDict); err != nil {
			return err
		}

		// 3. Add OutputIntents to catalog.
		pw.AddCatalogEntry(pdf.Name("OutputIntents"), pdf.Array{outputIntentRef})

		// 4. Write XMP metadata stream.
		xmpRef := pw.AllocObject()
		xmpData := generateXMP(cfg.level, cfg.metadata)
		xmpStream := pdf.Stream{
			Dict: pdf.Dict{
				pdf.Name("Type"):    pdf.Name("Metadata"),
				pdf.Name("Subtype"): pdf.Name("XML"),
			},
			Content: xmpData,
		}
		if err := pw.WriteObject(xmpRef, xmpStream); err != nil {
			return err
		}

		// 5. Add Metadata to catalog.
		pw.AddCatalogEntry(pdf.Name("Metadata"), xmpRef)

		return nil
	})
}
