package qrcode

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"hash/crc32"
)

// QRCode represents an encoded QR code.
type QRCode struct {
	Matrix  [][]bool             // Module grid (true = dark).
	Version int                  // QR version (1-40).
	ECLevel ErrorCorrectionLevel // Error correction level.
	Mask    int                  // Applied mask pattern (0-7).
}

// Encode encodes the given data string into a QR code at the specified
// error correction level. It automatically selects the smallest version
// and best mask pattern.
func Encode(data string, level ErrorCorrectionLevel) (*QRCode, error) {
	if len(data) == 0 {
		return nil, errors.New("qrcode: data must not be empty")
	}

	raw := []byte(data)
	m := detectMode(raw)

	version, err := selectVersion(raw, m, level)
	if err != nil {
		return nil, err
	}

	codewords, err := encodeData(raw, level, version)
	if err != nil {
		return nil, err
	}

	// Choose best mask.
	maskID := chooseBestMask(codewords, version, level)

	// Build final matrix.
	mat := buildMatrix(codewords, version, level, maskID)

	// Copy module data.
	grid := make([][]bool, mat.size)
	for i := range grid {
		grid[i] = make([]bool, mat.size)
		copy(grid[i], mat.modules[i])
	}

	return &QRCode{
		Matrix:  grid,
		Version: version,
		ECLevel: level,
		Mask:    maskID,
	}, nil
}

// Size returns the number of modules per side.
func (qr *QRCode) Size() int {
	return len(qr.Matrix)
}

// PNG renders the QR code as a PNG image. Each module is rendered as
// scale x scale pixels. A 4-module quiet zone is added on all sides.
func (qr *QRCode) PNG(scale int) ([]byte, error) {
	if scale < 1 {
		scale = 1
	}

	qrSize := len(qr.Matrix)
	quiet := 4
	imgSize := (qrSize + quiet*2) * scale
	width := imgSize
	height := imgSize

	// Build raw pixel data (filter byte + RGB pixels per row).
	rowLen := 1 + width*3 // filter byte + pixels
	raw := make([]byte, height*rowLen)

	for py := 0; py < height; py++ {
		rowOff := py * rowLen
		raw[rowOff] = 0 // no filter
		for px := 0; px < width; px++ {
			// Map pixel to module.
			mx := px/scale - quiet
			my := py/scale - quiet

			dark := false
			if mx >= 0 && mx < qrSize && my >= 0 && my < qrSize {
				dark = qr.Matrix[my][mx]
			}

			off := rowOff + 1 + px*3
			if dark {
				raw[off] = 0
				raw[off+1] = 0
				raw[off+2] = 0
			} else {
				raw[off] = 0xFF
				raw[off+1] = 0xFF
				raw[off+2] = 0xFF
			}
		}
	}

	// Compress with deflate (zlib wrapper for PNG).
	var compressed bytes.Buffer
	compressed.WriteByte(0x78) // zlib header: CM=8, CINFO=7
	compressed.WriteByte(0x01) // FCHECK (no dict, level 0)

	fw, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(raw); err != nil {
		return nil, err
	}
	if err := fw.Close(); err != nil {
		return nil, err
	}

	// Adler-32 checksum of uncompressed data.
	adler := adler32(raw)
	_ = binary.Write(&compressed, binary.BigEndian, adler)

	// Build PNG file.
	var buf bytes.Buffer

	// PNG signature.
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	// IHDR chunk.
	writeChunk(&buf, "IHDR", func(w *bytes.Buffer) {
		_ = binary.Write(w, binary.BigEndian, uint32(width))
		_ = binary.Write(w, binary.BigEndian, uint32(height))
		w.WriteByte(8) // bit depth
		w.WriteByte(2) // color type: RGB
		w.WriteByte(0) // compression method
		w.WriteByte(0) // filter method
		w.WriteByte(0) // interlace method
	})

	// IDAT chunk.
	writeChunk(&buf, "IDAT", func(w *bytes.Buffer) {
		w.Write(compressed.Bytes())
	})

	// IEND chunk.
	writeChunk(&buf, "IEND", func(w *bytes.Buffer) {})

	return buf.Bytes(), nil
}

// writeChunk writes a PNG chunk with type, data, and CRC.
func writeChunk(buf *bytes.Buffer, chunkType string, dataFn func(*bytes.Buffer)) {
	var data bytes.Buffer
	dataFn(&data)

	// Length.
	_ = binary.Write(buf, binary.BigEndian, uint32(data.Len()))
	// Type.
	buf.WriteString(chunkType)
	// Data.
	buf.Write(data.Bytes())
	// CRC (over type + data).
	crc := crc32.NewIEEE()
	crc.Write([]byte(chunkType))
	crc.Write(data.Bytes())
	_ = binary.Write(buf, binary.BigEndian, crc.Sum32())
}

// adler32 computes the Adler-32 checksum.
func adler32(data []byte) uint32 {
	const mod = 65521
	a, b := uint32(1), uint32(0)
	for _, d := range data {
		a = (a + uint32(d)) % mod
		b = (b + a) % mod
	}
	return (b << 16) | a
}
