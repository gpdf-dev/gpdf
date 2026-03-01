package barcode

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

// Format represents a barcode symbology.
type Format int

const (
	// Code128 is the Code 128 barcode symbology, supporting ASCII 0-127.
	Code128 Format = iota
)

// Barcode holds encoded barcode data.
type Barcode struct {
	// Data is the original input string.
	Data string
	// Format is the barcode symbology used.
	Format Format
	// Pattern is the expanded bar pattern where true=bar (black) and
	// false=space (white).
	Pattern []bool
}

// quietZone is the number of module units of white space on each side.
const quietZone = 10

// Encode creates a barcode from the given data string.
func Encode(data string, format Format) (*Barcode, error) {
	switch format {
	case Code128:
		symbols, err := encodeCode128(data)
		if err != nil {
			return nil, err
		}
		pattern := code128ToPattern(symbols)
		return &Barcode{
			Data:    data,
			Format:  format,
			Pattern: pattern,
		}, nil
	default:
		return nil, fmt.Errorf("barcode: unsupported format %d", format)
	}
}

// PNG renders the barcode as a PNG image. barWidth is the pixel width per
// module unit, and height is the total pixel height of the image.
func (b *Barcode) PNG(barWidth, height int) ([]byte, error) {
	if barWidth < 1 {
		return nil, fmt.Errorf("barcode: barWidth must be >= 1")
	}
	if height < 1 {
		return nil, fmt.Errorf("barcode: height must be >= 1")
	}

	// Total width in pixels: quiet zone + pattern + quiet zone.
	totalModules := quietZone + len(b.Pattern) + quietZone
	width := totalModules * barWidth

	// Build raw image data: each row is filter_byte + RGB pixels.
	rowSize := 1 + width*3 // filter byte + 3 bytes per pixel
	rawData := make([]byte, rowSize*height)

	// Build one row.
	row := make([]byte, rowSize)
	row[0] = 0x00 // filter: None
	for x := 0; x < width; x++ {
		module := x / barWidth
		isBar := false
		if module >= quietZone && module < quietZone+len(b.Pattern) {
			isBar = b.Pattern[module-quietZone]
		}
		off := 1 + x*3
		if isBar {
			row[off] = 0   // R
			row[off+1] = 0 // G
			row[off+2] = 0 // B
		} else {
			row[off] = 255   // R
			row[off+1] = 255 // G
			row[off+2] = 255 // B
		}
	}

	// Copy the same row for all scanlines.
	for y := 0; y < height; y++ {
		copy(rawData[y*rowSize:], row)
	}

	// Compress with flate (deflate).
	var compressed bytes.Buffer
	w, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("barcode: flate init: %w", err)
	}
	if _, err := w.Write(rawData); err != nil {
		return nil, fmt.Errorf("barcode: flate write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("barcode: flate close: %w", err)
	}

	// Build PNG file.
	var buf bytes.Buffer

	// PNG signature.
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	// IHDR chunk.
	writeChunk(&buf, "IHDR", func(d *bytes.Buffer) {
		_ = binary.Write(d, binary.BigEndian, uint32(width))  // width
		_ = binary.Write(d, binary.BigEndian, uint32(height)) // height
		d.WriteByte(8)                                        // bit depth
		d.WriteByte(2)                                        // color type: RGB
		d.WriteByte(0)                                        // compression
		d.WriteByte(0)                                        // filter
		d.WriteByte(0)                                        // interlace
	})

	// IDAT chunk: zlib header (0x78 0x9C) + deflated data + zlib checksum.
	writeChunk(&buf, "IDAT", func(d *bytes.Buffer) {
		// zlib header for default compression.
		d.WriteByte(0x78)
		d.WriteByte(0x9C)
		d.Write(compressed.Bytes())
		// Adler-32 checksum of uncompressed data.
		checksum := adler32(rawData)
		_ = binary.Write(d, binary.BigEndian, checksum)
	})

	// IEND chunk.
	writeChunk(&buf, "IEND", func(d *bytes.Buffer) {})

	return buf.Bytes(), nil
}

// writeChunk writes a PNG chunk with the given type and data to w.
func writeChunk(w *bytes.Buffer, chunkType string, fill func(d *bytes.Buffer)) {
	var data bytes.Buffer
	fill(&data)

	// Length.
	binary.Write(w, binary.BigEndian, uint32(data.Len()))
	// Type + Data for CRC.
	typeBytes := []byte(chunkType)
	w.Write(typeBytes)
	w.Write(data.Bytes())
	// CRC32 over type + data.
	crc := crc32.NewIEEE()
	crc.Write(typeBytes)
	crc.Write(data.Bytes())
	binary.Write(w, binary.BigEndian, crc.Sum32())
}

// adler32 computes the Adler-32 checksum of data.
func adler32(data []byte) uint32 {
	const mod = 65521
	var a uint32 = 1
	var b uint32 = 0
	for _, d := range data {
		a = (a + uint32(d)) % mod
		b = (b + a) % mod
	}
	return (b << 16) | a
}
