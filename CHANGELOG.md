# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Fixed
- Reader now supports compressed object streams (`/ObjStm`) and predictor-encoded streams, so PDFs written with a cross-reference stream — the default for most modern producers — can be opened, overlaid and merged instead of failing with `pdf: object N not found in xref` (#35)
  - `pdf/objstm.go`: new `/ObjStm` reader — decodes the stream body, parses the `N` object-number/offset header pairs up to `/First`, and resolves objects from it. Decoded streams are cached so sibling objects do not re-inflate the same body, and each object is parsed from its own bounded slice of the body.
  - `pdf/reader.go`: `parseXRefEntries` now records type-2 cross-reference entries, `GetObject` resolves them through the object stream, and `MaxObjectNumber` counts them — so `gpdf.Open`, `gpdf.Merge` and `Overlay` all work on these files. Entries from a newer xref section keep winning over an older `/Prev` section.
  - `pdf/predictor.go`: `/DecodeParms` predictor support (PNG predictors 10–15 and the TIFF predictor 2). Real cross-reference streams are almost always written with `/Predictor 12`, so without this the decoded xref entries were garbage.
  - `pdf/objstm_test.go`, `pdf/predictor_test.go`: coverage for object-stream reads, merging an object-stream PDF, stale xref indices, newest-section-wins precedence, and each predictor filter type
- Incremental update xref entries are now exactly 20 bytes, so the output of `Open` → `Overlay` → `Save` is no longer reported as damaged by strict readers
  - `pdf/modifier.go`: `writeIncrementalXRef` wrote `"%010d %05d n \r\n"` — 21 bytes. Readers index the table by fixed-width offset, so the extra byte shifted every following entry and forced a cross-reference reconstruction (`qpdf --check`: `invalid xref entry`). The non-incremental writer in `pdf/xref.go` was already correct; its doc comment showed the wrong width and has been corrected.
  - `pdf/modifier_extra_test.go`: regression test asserting every appended xref entry line is 20 bytes

### Changed
- `_validation` now resolves again: `github.com/hhrutter/lzw` was pinned to `v1.0.4`, a version that no longer exists upstream, and the `go.sum` entry for `github.com/hhrutter/tiff v1.0.4` no longer matched the module. The indirect requirements are realigned with what `pdfcpu v0.9.1` declares (`lzw v1.0.0`, `tiff v1.0.1`); the recorded checksums are verified against `sum.golang.org`.
  - `_validation/validation_test.go`: `TestQPDF_ObjectStreamRoundTrip` converts a generated PDF with `qpdf --object-streams=generate`, then merges and overlays it and validates the result with pdfcpu (skipped when `qpdf` is not installed)

## [1.0.11] - 2026-05-18

### Fixed
- `ResolvedFont.ID` now uses the font registration key instead of the resolved family name, so per-family overrides (e.g. registering a CJK font under a custom key) are honored at render time (#30)
  - `template/fontresolver.go`: the resolver stores the input registration key on `ResolvedFont.ID` so downstream font lookups in the renderer match the registered font, not the resolved family fallback
  - `template/fontresolver_test.go`: regression coverage for registration-key vs family-name ID
  - Golden refresh: `_examples/testdata/golden/32_cjk_text.pdf` and `32d_cjk_mixed.pdf` regenerated to reflect the corrected font ID embedded in the content stream

## [1.0.10] - 2026-05-07

### Added
- Per-column horizontal text alignment for tables (#26)
  - Builder: `template.ColumnAlign(aligns ...document.TextAlign)` — sets the horizontal alignment for each column in both the header and body. Columns without a provided alignment fall back to the default left alignment.
  - JSON / GoTemplate schema: `table.columnAlign: ["left", "center", "right"]`
- Example tests: `_examples/{builder,json,gotemplate}/38_table_column_align_test.go` with shared golden

## [1.0.9] - 2026-05-06

### Fixed
- AutoRow/Row no longer splits across page boundaries — a row that does not fit in the remaining space on the current page now moves as a whole to the next page, instead of placing the columns that fit on the current page and re-rendering the full row on the next (#24)
  - `template/grid.go`: `RowBuilder.build()` now sets `BreakInside=BreakAvoid` on the horizontal Box so rows are treated as atomic layout units
  - `document/layout/block.go`: `layoutVerticalChild` falls back to a normal split when a `BreakAvoid` child is the first node on a fresh page, preventing an infinite loop of empty pages when a row is taller than a full page

## [1.0.8] - 2026-05-06

### Added
- Padding support on Text elements following the CSS box model (#23)
  - Builder: `template.TextPadding(Edges)`
  - JSON / GoTemplate schema: `text.style.padding` (uniform, e.g. `"10mm"`) and `text.style.paddings` (CSS shorthand `[top, right, bottom, left]`, 1–4 values)
  - Layout engine: `Style.Padding` and `Style.Border` participate in text flow — wrap width shrinks to the inner content area, lines are offset by the top/left inset, and `BgColor` fills the padded box (`document/layout/flow.go`)
- Example tests: `_examples/{builder,json,gotemplate}/36_text_padding_test.go` with shared golden

## [1.0.7] - 2026-04-29

### Added
- Borders and backgrounds for tables, images, and boxes
  - Builder: `Border(opts...)` with `BorderWidth`, `BorderWidths`, `BorderColor`, `BorderLine`; applied via `WithTableBorder`, `WithImageBorder`, `WithBoxBorder`, `WithTextBorder`, plus `WithImageBackground` / table & box background options
  - JSON / GoTemplate schema: `SchemaBorder { width, widths, color, style }`; `image.border` / `image.background`; `table.border` / `table.cellBorder` / `table.borderCollapse` / `table.background`
  - Layout engine support: borders participate in box-model sizing and pagination (`document/layout/{block,engine,paging}.go`)
- Example tests: `_examples/{builder,json,gotemplate}/35_border_test.go` with shared golden

## [1.0.6] - 2026-04-20

### Added
- Minimum display size constraints for images and QR codes — raise layout overflow to the next page when the target box would render below `minWidth` / `minHeight` (#19)
  - Builder: `MinDisplayWidth(v)` / `MinDisplayHeight(v)` options on Image and QR
  - JSON / GoTemplate schema: `image.minWidth` / `image.minHeight` / `qr.minWidth` / `qr.minHeight`
  - Layout engine propagates overflow when the constraint is violated (`document/layout/block.go`)
- Example tests: `_examples/{builder,json,gotemplate}/34_image_min_size_test.go` with shared golden

## [1.0.5] - 2026-04-19

### Fixed
- Text alignment precision for Standard 14 fonts — right/center alignment now lands at the true container edge instead of drifting up to ~17pt for large bold text
  - `template/fontresolver.go`: `MeasureString` and `LineBreak` now use Adobe Core 14 AFM advance widths when no TTF is registered, instead of the `charCount × size × 0.5` approximation
  - `Resolve` now normalizes an empty `FontFamily` to `Helvetica`, matching the PDF renderer's default
  - Metrics (Ascender / Descender / CapHeight) from AFM are returned for Standard 14 fonts — previously hard-coded 0.8 / -0.2 / 0.7 fallback

### Added
- `pdf/font/standard14.go` — Adobe Standard 14 font constants, `IsStandard14`, `Standard14Metrics`, `Standard14Width`, `NewStandard14Font`
- `pdf/font/standard14_data.go` — AFM-derived width tables and metrics for Helvetica / Times / Courier / Symbol / ZapfDingbats families (14 fonts, printable ASCII coverage)
- Tests: `pdf/font/standard14_test.go`, `template/fontresolver_test.go`

## [1.0.4] - 2026-04-07

### Added
- AcroForm flatten support — flatten form fields into static content (#17)

## [1.0.3] - 2026-03-23

### Fixed
- Multi-page table support — tables inside Row/Col now automatically split across pages
  - `layoutHorizontal` propagates child overflow to the paginator
  - Table headers repeat on each continuation page (existing `layoutTable` logic)

## [1.0.2] - 2026-03-23

### Added
- PDF document merging — combine pages from multiple PDFs into one (#11)
  - `pdf.MergePDFs()`: Core merge engine with object reference remapping
  - `gpdf.Merge()`: High-level facade with `Source`, `PageRange`, `WithMergeMetadata()`
  - `pdf.Writer.AddRawPage()`, `PageTreeRef()`: Raw page insertion support
  - Merge examples: basic merge, page range extraction, metadata, merge + overlay, issue #11 scenario

## [1.0.1] - 2026-03-22

### Added
- RFC 3161 timestamping for digital signatures

## [1.0.0] - 2026-03-22

### Added
- Existing PDF overlay — open, read, and modify existing PDFs
  - `pdf.Reader`: PDF parser with XRef table/stream parsing, page tree traversal, object caching
  - `pdf.Modifier`: Incremental Update engine (non-destructive append to existing PDF)
  - `template.ExistingDocument`: High-level API with `Overlay()`, `EachPage()`, `Save()`
  - `gpdf.Open()`: Facade entry point for opening existing PDFs
  - `render.OverlayRenderer`: Content stream capture for overlay rendering
- Overlay examples: text watermark, page numbers, stamps, confidential header, facade usage

## [0.9.0] - 2026-03-05

### Added
- Absolute positioning for placing elements at exact XY coordinates
- `textIndent` and `cellVAlign` support in JSON/GoTemplate schema
- Comprehensive English documentation for gpdf core
- CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md
- GitHub Issue templates (bug report, feature request) and Pull Request template
- CHANGELOG.md
- GoDoc enrichment with `doc.go` files, missing comments, and example tests
- Test coverage improved to 92.0%

### Changed
- Moved Benchmark section after Features in all READMEs
- Unified architecture diagrams to English across all README translations
- Reduced cyclomatic complexity of `applySchemaStyle`

### Fixed
- Stabilized golden tests by using version-independent Producer metadata

## [0.8.0] - 2026-03-03

### Added
- Image fit modes (contain, cover, fill, none)
- Image embedding from file paths
- PNG alpha transparency support
- JSON schema and Go template examples for all features

### Changed
- Restructured `_examples/` into `builder/`, `json/`, `gotemplate/`, `component/` subdirectories
- Unified golden files across builder/json/gotemplate into shared directory
- Reduced cyclomatic complexity in `layoutImage` and `parseColor`

## [0.7.0] - 2026-03-02

### Added
- Reusable components (Invoice, Report, Letter templates)
- Fuzz testing for all packages
- PDF output validation with pdfcpu

## [0.6.0] - 2026-03-02

### Added
- Go template integration (`gpdf.FromGoTemplate`)
- JSON schema generation (`gpdf.FromJSON`)

### Fixed
- UTF-8 to WinAnsiEncoding conversion in PDF literal strings

## [0.5.0] - 2026-03-02

### Added
- Layer 1: PDF Primitives (Writer, XRef, Font, Stream, Image)
- Layer 2: Document Model (Node, Box, Style, Layout Engine)
- Layer 3: Template API (Builder, 12-column Grid, Components)
- CJK support (TrueType + CMap + subsetting)
- Tables with headers, column widths, striped rows, vertical alignment
- Headers & Footers with page numbers
- Multiple units (pt, mm, cm, in, em, %)
- Color spaces (RGB, Grayscale, CMYK)
- JPEG/PNG image embedding
- Document metadata (title, author, subject, creator)
- QR code generation with error correction levels
- Barcode generation (Code 128)
- Text decorations (underline, strikethrough, letter spacing, text indent)
- Lists (bulleted and numbered)
- Buildinfo package with version in PDF Producer metadata
- Benchmarks (10-30x faster than alternatives)
- CI/CD with GitHub Actions
- Multi-language READMEs (EN, JA, ZH, KO, ES, PT)

### Fixed
- Reed-Solomon coefficient order in QR code encoder
- binary.Write return value handling for errcheck lint

[Unreleased]: https://github.com/gpdf-dev/gpdf/compare/v1.0.11...HEAD
[1.0.11]: https://github.com/gpdf-dev/gpdf/compare/v1.0.10...v1.0.11
[1.0.10]: https://github.com/gpdf-dev/gpdf/compare/v1.0.9...v1.0.10
[1.0.9]: https://github.com/gpdf-dev/gpdf/compare/v1.0.8...v1.0.9
[1.0.8]: https://github.com/gpdf-dev/gpdf/compare/v1.0.7...v1.0.8
[1.0.7]: https://github.com/gpdf-dev/gpdf/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/gpdf-dev/gpdf/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/gpdf-dev/gpdf/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/gpdf-dev/gpdf/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/gpdf-dev/gpdf/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/gpdf-dev/gpdf/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/gpdf-dev/gpdf/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/gpdf-dev/gpdf/compare/v0.9.0...v1.0.0
[0.9.0]: https://github.com/gpdf-dev/gpdf/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/gpdf-dev/gpdf/compare/v0.5.0...v0.8.0
[0.7.0]: https://github.com/gpdf-dev/gpdf/releases/tag/v0.7.0
[0.6.0]: https://github.com/gpdf-dev/gpdf/releases/tag/v0.6.0
[0.5.0]: https://github.com/gpdf-dev/gpdf/releases/tag/v0.5.0
