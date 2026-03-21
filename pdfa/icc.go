package pdfa

import (
	"bytes"
	"encoding/binary"
	"math"
)

// sRGBICCProfile returns a minimal sRGB ICC v2 profile suitable for PDF/A.
// The profile is generated programmatically in pure Go with no external
// dependencies or embedded files.
func sRGBICCProfile() []byte {
	return buildSRGBProfile()
}

// s15Fixed16 converts a float64 to ICC s15Fixed16Number.
func s15Fixed16(f float64) uint32 {
	return uint32(int32(math.Round(f * 65536)))
}

func buildSRGBProfile() []byte {
	// sRGB color space parameters (IEC 61966-2-1)
	// XYZ matrix columns (D50 adapted):
	rX, rY, rZ := 0.4360747, 0.2225045, 0.0139322
	gX, gY, gZ := 0.3850649, 0.7168786, 0.0971045
	bX, bY, bZ := 0.1430804, 0.0606169, 0.7141733
	// D50 white point
	wpX, wpY, wpZ := 0.9504559, 1.0, 1.0890578

	// Build tags
	type tagEntry struct {
		sig  [4]byte
		data []byte
	}

	// XYZ type data helper
	xyzType := func(x, y, z float64) []byte {
		var buf bytes.Buffer
		buf.WriteString("XYZ ")                                 // type signature
		_ = binary.Write(&buf, binary.BigEndian, uint32(0))     // reserved
		_ = binary.Write(&buf, binary.BigEndian, s15Fixed16(x)) // X
		_ = binary.Write(&buf, binary.BigEndian, s15Fixed16(y)) // Y
		_ = binary.Write(&buf, binary.BigEndian, s15Fixed16(z)) // Z
		return buf.Bytes()
	}

	// Curve type with a single gamma value (u8Fixed8Number).
	// count=1 means the single entry is interpreted as a gamma exponent.
	curveGamma := func(gamma float64) []byte {
		var buf bytes.Buffer
		buf.WriteString("curv")                                     // type signature
		_ = binary.Write(&buf, binary.BigEndian, uint32(0))         // reserved
		_ = binary.Write(&buf, binary.BigEndian, uint32(1))         // count = 1 means gamma
		_ = binary.Write(&buf, binary.BigEndian, uint16(gamma*256)) // u8Fixed8Number
		// Pad to 4-byte boundary
		for buf.Len()%4 != 0 {
			buf.WriteByte(0)
		}
		return buf.Bytes()
	}

	// Text description type (profileDescriptionTag)
	descType := func(text string) []byte {
		var buf bytes.Buffer
		buf.WriteString("desc")                                       // type signature
		_ = binary.Write(&buf, binary.BigEndian, uint32(0))           // reserved
		_ = binary.Write(&buf, binary.BigEndian, uint32(len(text)+1)) // ASCII length incl NUL
		buf.WriteString(text)
		buf.WriteByte(0) // NUL terminator
		// Unicode localizable (empty)
		_ = binary.Write(&buf, binary.BigEndian, uint32(0)) // unicode language code
		_ = binary.Write(&buf, binary.BigEndian, uint32(0)) // unicode count
		// ScriptCode (empty)
		_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // scriptcode code
		buf.WriteByte(0)                                     // scriptcode count
		buf.Write(make([]byte, 67))                          // scriptcode string (67 bytes)
		// Pad to 4-byte boundary
		for buf.Len()%4 != 0 {
			buf.WriteByte(0)
		}
		return buf.Bytes()
	}

	// Text type (for copyright)
	textType := func(text string) []byte {
		var buf bytes.Buffer
		buf.WriteString("text")                             // type signature
		_ = binary.Write(&buf, binary.BigEndian, uint32(0)) // reserved
		buf.WriteString(text)
		buf.WriteByte(0) // NUL
		for buf.Len()%4 != 0 {
			buf.WriteByte(0)
		}
		return buf.Bytes()
	}

	gamma22 := curveGamma(2.2)

	tags := []tagEntry{
		{sig: [4]byte{'d', 'e', 's', 'c'}, data: descType("sRGB IEC61966-2.1")},
		{sig: [4]byte{'w', 't', 'p', 't'}, data: xyzType(wpX, wpY, wpZ)},
		{sig: [4]byte{'r', 'X', 'Y', 'Z'}, data: xyzType(rX, rY, rZ)},
		{sig: [4]byte{'g', 'X', 'Y', 'Z'}, data: xyzType(gX, gY, gZ)},
		{sig: [4]byte{'b', 'X', 'Y', 'Z'}, data: xyzType(bX, bY, bZ)},
		{sig: [4]byte{'r', 'T', 'R', 'C'}, data: gamma22},
		{sig: [4]byte{'g', 'T', 'R', 'C'}, data: gamma22},
		{sig: [4]byte{'b', 'T', 'R', 'C'}, data: gamma22},
		{sig: [4]byte{'c', 'p', 'r', 't'}, data: textType("Public Domain")},
	}

	// Calculate offsets
	headerSize := 128
	tagTableSize := 4 + len(tags)*12 // count(4) + entries(12 each)
	dataOffset := headerSize + tagTableSize
	// Align data start to 4 bytes
	if dataOffset%4 != 0 {
		dataOffset += 4 - dataOffset%4
	}

	// Calculate tag offsets and total size
	offsets := make([]int, len(tags))
	currentOffset := dataOffset
	for i, tag := range tags {
		offsets[i] = currentOffset
		currentOffset += len(tag.data)
		// Align to 4 bytes
		if currentOffset%4 != 0 {
			currentOffset += 4 - currentOffset%4
		}
	}
	totalSize := currentOffset

	// Build profile into a buffer
	buf := new(bytes.Buffer)
	buf.Grow(totalSize)

	// === Header (128 bytes) ===
	_ = binary.Write(buf, binary.BigEndian, uint32(totalSize)) // Profile size
	buf.WriteString("gpdf")                                     // Preferred CMM Type
	_ = binary.Write(buf, binary.BigEndian, uint32(0x02100000)) // Version 2.1.0
	buf.WriteString("mntr")                                     // Device class: monitor
	buf.WriteString("RGB ")                                     // Color space: RGB
	buf.WriteString("XYZ ")                                     // PCS: XYZ
	// Date/time (12 bytes) - 2024-01-01 00:00:00
	_ = binary.Write(buf, binary.BigEndian, uint16(2024)) // year
	_ = binary.Write(buf, binary.BigEndian, uint16(1))    // month
	_ = binary.Write(buf, binary.BigEndian, uint16(1))    // day
	_ = binary.Write(buf, binary.BigEndian, uint16(0))    // hour
	_ = binary.Write(buf, binary.BigEndian, uint16(0))    // minute
	_ = binary.Write(buf, binary.BigEndian, uint16(0))    // second
	buf.WriteString("acsp")                                // File signature (always "acsp")
	buf.WriteString("APPL")                                // Primary platform
	_ = binary.Write(buf, binary.BigEndian, uint32(0))     // Profile flags
	buf.WriteString("gpdf")                                // Device manufacturer
	buf.WriteString("sRGB")                                // Device model
	_ = binary.Write(buf, binary.BigEndian, uint64(0))     // Device attributes
	_ = binary.Write(buf, binary.BigEndian, uint32(0))     // Rendering intent (perceptual)
	// PCS illuminant (D50 XYZ)
	_ = binary.Write(buf, binary.BigEndian, s15Fixed16(0.9642))
	_ = binary.Write(buf, binary.BigEndian, s15Fixed16(1.0))
	_ = binary.Write(buf, binary.BigEndian, s15Fixed16(0.8249))
	buf.WriteString("gpdf")     // Profile creator
	buf.Write(make([]byte, 16)) // Profile ID (MD5, 16 bytes) - zeros
	buf.Write(make([]byte, 28)) // Reserved (28 bytes)

	// === Tag table ===
	_ = binary.Write(buf, binary.BigEndian, uint32(len(tags))) // Tag count
	for i, tag := range tags {
		buf.Write(tag.sig[:])
		_ = binary.Write(buf, binary.BigEndian, uint32(offsets[i]))
		_ = binary.Write(buf, binary.BigEndian, uint32(len(tag.data)))
	}

	// Pad to data start
	for buf.Len() < dataOffset {
		buf.WriteByte(0)
	}

	// === Tag data ===
	for _, tag := range tags {
		buf.Write(tag.data)
		// Pad to 4-byte alignment
		for buf.Len()%4 != 0 {
			buf.WriteByte(0)
		}
	}

	return buf.Bytes()
}
