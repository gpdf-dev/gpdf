package svg

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// maxDepth bounds element nesting (and <use> indirection) so that hostile
// input cannot exhaust the stack.
const maxDepth = 256

// maxElements bounds the number of parsed elements.
const maxElements = 200000

// node is a parsed SVG element.
type node struct {
	name     string            // local element name, e.g. "path"
	attrs    map[string]string // local attribute name -> value
	children []*node
	text     string // character data (used by <style>)
}

// Image is a parsed SVG document ready to be rendered.
type Image struct {
	// Width and Height are the intrinsic size in points.
	Width, Height float64

	root    *node
	ids     map[string]*node
	css     []cssRule
	viewBox [4]float64 // minX, minY, width, height
	noAR    bool       // preserveAspectRatio="none"
}

// Parse parses SVG data. It returns an error when the data is not
// well-formed XML or has no <svg> root element.
func Parse(data []byte) (*Image, error) {
	root, err := parseTree(data)
	if err != nil {
		return nil, err
	}
	img := &Image{root: root, ids: make(map[string]*node)}
	img.index(root, 0)
	img.css = parseCSS(collectStyleText(root))
	img.setupViewport()
	return img, nil
}

func parseTree(data []byte) (*node, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity
	// Accept any declared encoding; non-UTF-8 content is passed through.
	dec.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }

	tb := &treeBuilder{}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("svg: %w", err)
		}
		if err := tb.token(tok); err != nil {
			return nil, err
		}
	}
	if tb.root == nil || tb.root.name != "svg" {
		return nil, errors.New("svg: missing <svg> root element")
	}
	return tb.root, nil
}

// treeBuilder assembles XML tokens into a node tree.
type treeBuilder struct {
	stack []*node
	root  *node
	count int
}

func (tb *treeBuilder) token(tok xml.Token) error {
	switch t := tok.(type) {
	case xml.StartElement:
		tb.count++
		if tb.count > maxElements || len(tb.stack) >= maxDepth {
			return errors.New("svg: document too large or too deeply nested")
		}
		n := &node{name: t.Name.Local, attrs: make(map[string]string, len(t.Attr))}
		for _, a := range t.Attr {
			n.attrs[a.Name.Local] = a.Value
		}
		if len(tb.stack) > 0 {
			parent := tb.stack[len(tb.stack)-1]
			parent.children = append(parent.children, n)
		} else if tb.root == nil {
			tb.root = n
		}
		tb.stack = append(tb.stack, n)
	case xml.EndElement:
		if len(tb.stack) > 0 {
			tb.stack = tb.stack[:len(tb.stack)-1]
		}
	case xml.CharData:
		if len(tb.stack) > 0 && tb.stack[len(tb.stack)-1].name == "style" {
			tb.stack[len(tb.stack)-1].text += string(t)
		}
	}
	return nil
}

// index records elements by id.
func (img *Image) index(n *node, depth int) {
	if depth > maxDepth {
		return
	}
	if id := n.attrs["id"]; id != "" {
		if _, dup := img.ids[id]; !dup {
			img.ids[id] = n
		}
	}
	for _, c := range n.children {
		img.index(c, depth+1)
	}
}

func collectStyleText(root *node) string {
	var b strings.Builder
	var walk func(n *node, depth int)
	walk = func(n *node, depth int) {
		if depth > maxDepth {
			return
		}
		if n.name == "style" {
			b.WriteString(n.text)
			b.WriteByte('\n')
		}
		for _, c := range n.children {
			walk(c, depth+1)
		}
	}
	walk(root, 0)
	return b.String()
}

// setupViewport resolves the intrinsic size and the viewBox.
func (img *Image) setupViewport() {
	a := img.root.attrs
	vb, hasVB := parseViewBox(a["viewBox"])

	w, wOK := parseAbsLength(a["width"])
	h, hOK := parseAbsLength(a["height"])
	switch {
	case wOK && hOK:
	case wOK && hasVB:
		h = w * vb[3] / vb[2]
	case hOK && hasVB:
		w = h * vb[2] / vb[3]
	case hasVB:
		w, h = vb[2], vb[3]
	default:
		// CSS default replaced-element size.
		if !wOK {
			w = 300
		}
		if !hOK {
			h = 150
		}
	}
	if !hasVB {
		vb = [4]float64{0, 0, w, h}
	}
	img.Width, img.Height = w, h
	img.viewBox = vb
	img.noAR = strings.HasPrefix(strings.TrimSpace(a["preserveAspectRatio"]), "none")
}

func parseViewBox(s string) ([4]float64, bool) {
	nums := parseNumberList(s)
	if len(nums) != 4 || !(nums[2] > 0 && nums[2] <= maxSize) || !(nums[3] > 0 && nums[3] <= maxSize) {
		return [4]float64{}, false
	}
	return [4]float64{nums[0], nums[1], nums[2], nums[3]}, true
}

// parseAbsLength parses a width/height attribute into points. Percentages
// and invalid or non-positive values report false.
func parseAbsLength(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasSuffix(s, "%") {
		return 0, false
	}
	v, ok := parseLength(s, 0)
	return v, ok && v > 0 && v <= maxSize
}

// maxSize bounds intrinsic sizes and viewBox dimensions; larger values are
// treated as invalid.
const maxSize = 1e6
