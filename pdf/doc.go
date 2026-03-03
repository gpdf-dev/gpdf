// Package pdf provides low-level PDF object types, a streaming writer, and
// color utilities for producing valid PDF binary output.
//
// This package is Layer 1 of the gpdf architecture and is used by the higher
// layers (document and template) to emit raw PDF content. Most application
// code should use the template or document packages instead.
//
// # PDF Objects
//
// Every PDF value type implements the [Object] interface so it can be
// serialized with WriteTo:
//
//   - [Name] — PDF name objects such as /Type or /Font
//   - [LiteralString] — parenthesized strings e.g. (Hello)
//   - [HexString] — angle-bracket hex strings e.g. <48656C6C6F>
//   - [Integer] — integer numbers
//   - [Real] — floating-point numbers
//   - [Boolean] — true or false
//   - [Null] — the PDF null object
//   - [Dict] — key-value dictionaries
//   - [Array] — ordered sequences of objects
//   - [Stream] — binary data streams (optionally compressed)
//   - [ObjectRef] — indirect object references e.g. 1 0 R
//   - [Rectangle] — a four-element array representing a bounding box
//
// # Writer
//
// [NewWriter] creates a streaming PDF writer that manages object allocation,
// cross-reference tables, font registration, and image embedding:
//
//	w := pdf.NewWriter(out)
//	w.RegisterFont("NotoSans", fontData)
//	w.AddPage(page)
//	w.Close()
//
// # Colors
//
// The [Color] type supports RGB, Grayscale, and CMYK color spaces.
// Convenience constructors and named colors are provided:
//
//	c := pdf.RGB(0.2, 0.4, 0.8)
//	c := pdf.RGBHex(0x1A237E)
//	c := pdf.Gray(0.5)
//	c := pdf.CMYK(0, 0.5, 1, 0)
//
// Pre-defined colors: [Black], [White], [Red], [Green], [Blue], [Yellow],
// [Cyan], [Magenta].
package pdf
