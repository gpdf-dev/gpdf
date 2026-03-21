# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

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

[Unreleased]: https://github.com/gpdf-dev/gpdf/compare/v1.0.0...HEAD
[0.9.0]: https://github.com/gpdf-dev/gpdf/compare/v0.8.0...v1.0.0
[0.8.0]: https://github.com/gpdf-dev/gpdf/compare/v0.5.0...v0.8.0
[0.7.0]: https://github.com/gpdf-dev/gpdf/releases/tag/v0.7.0
[0.6.0]: https://github.com/gpdf-dev/gpdf/releases/tag/v0.6.0
[0.5.0]: https://github.com/gpdf-dev/gpdf/releases/tag/v0.5.0
