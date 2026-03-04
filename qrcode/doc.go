// Package qrcode implements QR code encoding per ISO 18004.
//
// It supports the full range of QR code versions (1-40), four error
// correction levels, and automatic version/mask selection. The encoded
// QR code can be rendered as a PNG image for embedding in PDF documents.
//
// # Encoding
//
// Use [Encode] to encode a data string into a QR code:
//
//	qr, err := qrcode.Encode("https://gpdf.dev", qrcode.LevelM)
//
// # Rendering
//
// Call [QRCode.PNG] to render the QR code as a PNG image with a given
// scale (pixels per module):
//
//	pngData, err := qr.PNG(10) // 10 pixels per module
//
// # Error Correction Levels
//
// Four error correction levels are available, trading data capacity for
// recovery capability:
//
//   - [LevelL] — ~7% recovery (maximum data capacity)
//   - [LevelM] — ~15% recovery (default)
//   - [LevelQ] — ~25% recovery
//   - [LevelH] — ~30% recovery (maximum error resilience)
package qrcode
