// Package barcode implements barcode encoding and rendering for PDF
// generation. Currently it supports the Code 128 symbology, which encodes
// the full ASCII character set (0-127).
//
// # Encoding
//
// Use [Encode] to create a barcode from a data string:
//
//	bc, err := barcode.Encode("INV-2026-0001", barcode.Code128)
//
// # Rendering
//
// Call [Barcode.PNG] to render the barcode as a PNG image suitable for
// embedding in a PDF. The barWidth parameter sets the pixel width per module,
// and height sets the total image height in pixels:
//
//	pngData, err := bc.PNG(2, 100) // 2px per module, 100px tall
//
// # Supported Formats
//
//   - [Code128] — Code 128 (ASCII 0-127, auto subset switching A/B/C)
package barcode
