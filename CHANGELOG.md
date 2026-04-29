# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

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

[Unreleased]: https://github.com/gpdf-dev/gpdf/compare/v1.0.7...HEAD
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
