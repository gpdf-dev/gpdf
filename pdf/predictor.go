package pdf

import "fmt"

// predictorParams holds the /DecodeParms entries that affect predictor decoding
// (ISO 32000-1 §7.4.4.4). Zero value means "no predictor".
type predictorParams struct {
	predictor        int
	colors           int
	bitsPerComponent int
	columns          int
}

// readPredictorParams extracts predictor settings from a /DecodeParms dict,
// applying the defaults defined by the specification.
func (r *Reader) readPredictorParams(d Dict) (predictorParams, error) {
	p := predictorParams{predictor: 1, colors: 1, bitsPerComponent: 8, columns: 1}
	if d == nil {
		return p, nil
	}

	fields := []struct {
		key string
		dst *int
	}{
		{"Predictor", &p.predictor},
		{"Colors", &p.colors},
		{"BitsPerComponent", &p.bitsPerComponent},
		{"Columns", &p.columns},
	}
	for _, f := range fields {
		if _, ok := d[Name(f.key)]; !ok {
			continue
		}
		v, err := r.intFromDict(d, f.key)
		if err != nil {
			return p, err
		}
		*f.dst = v
	}

	if p.colors < 1 || p.bitsPerComponent < 1 || p.columns < 1 {
		return p, fmt.Errorf("pdf: invalid /DecodeParms (Colors=%d BitsPerComponent=%d Columns=%d)",
			p.colors, p.bitsPerComponent, p.columns)
	}
	return p, nil
}

// applyPredictor reverses the predictor applied before compression.
// Predictor 1 (or absent) is a no-op, 2 is the TIFF predictor, and 10–15 are
// the PNG predictors, which record the filter type per row.
func applyPredictor(data []byte, p predictorParams) ([]byte, error) {
	switch {
	case p.predictor <= 1:
		return data, nil
	case p.predictor == 2:
		return applyTIFFPredictor(data, p)
	case p.predictor >= 10:
		return applyPNGPredictor(data, p)
	default:
		return nil, fmt.Errorf("pdf: unsupported /Predictor %d", p.predictor)
	}
}

// rowLength returns the number of bytes in one unfiltered row.
func (p predictorParams) rowLength() int {
	bits := p.columns * p.colors * p.bitsPerComponent
	return (bits + 7) / 8
}

// pixelBytes returns the distance in bytes to the preceding pixel, at least 1.
func (p predictorParams) pixelBytes() int {
	n := (p.colors*p.bitsPerComponent + 7) / 8
	if n < 1 {
		return 1
	}
	return n
}

// applyPNGPredictor reverses PNG row filtering (RFC 2083 §6). Each encoded row
// is one filter-type byte followed by rowLen filtered bytes.
func applyPNGPredictor(data []byte, p predictorParams) ([]byte, error) {
	rowLen := p.rowLength()
	if rowLen < 1 {
		return nil, fmt.Errorf("pdf: predictor row length is zero")
	}
	bpp := p.pixelBytes()

	out := make([]byte, 0, len(data))
	prev := make([]byte, rowLen)
	row := make([]byte, rowLen)

	for pos := 0; pos < len(data); pos += 1 + rowLen {
		filter := data[pos]
		end := pos + 1 + rowLen
		if end > len(data) {
			// Tolerate a truncated final row rather than discarding the
			// rows already decoded — some producers pad inconsistently.
			end = len(data)
		}
		n := copy(row, data[pos+1:end])
		for i := n; i < rowLen; i++ {
			row[i] = 0
		}

		if err := unfilterPNGRow(filter, row, prev, bpp); err != nil {
			return nil, err
		}

		out = append(out, row[:n]...)
		copy(prev, row)
	}
	return out, nil
}

// unfilterPNGRow reverses one row's filter in place.
func unfilterPNGRow(filter byte, row, prev []byte, bpp int) error {
	switch filter {
	case 0: // None
	case 1: // Sub
		for i := bpp; i < len(row); i++ {
			row[i] += row[i-bpp]
		}
	case 2: // Up
		for i := range row {
			row[i] += prev[i]
		}
	case 3: // Average
		for i := range row {
			left := 0
			if i >= bpp {
				left = int(row[i-bpp])
			}
			row[i] += byte((left + int(prev[i])) / 2)
		}
	case 4: // Paeth
		for i := range row {
			var left, upLeft byte
			if i >= bpp {
				left = row[i-bpp]
				upLeft = prev[i-bpp]
			}
			row[i] += paeth(left, prev[i], upLeft)
		}
	default:
		return fmt.Errorf("pdf: unsupported PNG predictor filter type %d", filter)
	}
	return nil
}

// paeth is the PNG Paeth predictor function.
func paeth(a, b, c byte) byte {
	p := int(a) + int(b) - int(c)
	pa, pb, pc := abs(p-int(a)), abs(p-int(b)), abs(p-int(c))
	if pa <= pb && pa <= pc {
		return a
	}
	if pb <= pc {
		return b
	}
	return c
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// applyTIFFPredictor reverses TIFF predictor 2, which stores each component as
// the difference from the same component of the previous pixel.
func applyTIFFPredictor(data []byte, p predictorParams) ([]byte, error) {
	if p.bitsPerComponent != 8 {
		return nil, fmt.Errorf("pdf: TIFF predictor supports 8 bits per component only, got %d", p.bitsPerComponent)
	}

	rowLen := p.rowLength()
	out := make([]byte, len(data))
	copy(out, data)
	for start := 0; start < len(out); start += rowLen {
		end := start + rowLen
		if end > len(out) {
			end = len(out)
		}
		for i := start + p.colors; i < end; i++ {
			out[i] += out[i-p.colors]
		}
	}
	return out, nil
}
