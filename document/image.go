package document

// ImageSource carries the raw image data along with its format and
// intrinsic pixel dimensions.
type ImageSource struct {
	// Data holds the raw image bytes (e.g., JPEG or PNG encoded).
	Data []byte
	// Format identifies the image encoding.
	Format ImageFormat
	// Width is the intrinsic pixel width.
	Width int
	// Height is the intrinsic pixel height.
	Height int
}

// ImageFormat identifies the encoding of an image's raw data.
type ImageFormat int

const (
	// ImageJPEG indicates JPEG encoding.
	ImageJPEG ImageFormat = iota
	// ImagePNG indicates PNG encoding.
	ImagePNG
	// ImageSVG indicates SVG vector graphics.
	ImageSVG
)

// Image is a leaf document node that renders an image within the
// document layout. The FitMode controls how the image is scaled
// relative to its layout bounds.
type Image struct {
	// Source holds the image data and metadata.
	Source ImageSource
	// ImgStyle controls spacing and decorative properties.
	ImgStyle Style
	// FitMode determines how the image is scaled within its bounds.
	FitMode ImageFitMode
	// DisplayWidth is the explicit display width set by template options.
	DisplayWidth Value
	// DisplayHeight is the explicit display height set by template options.
	DisplayHeight Value
}

// ImageFitMode controls how an image is scaled within its layout bounds.
type ImageFitMode int

const (
	// FitContain scales the image to fit entirely within the bounds
	// while preserving aspect ratio. The image may be smaller than the
	// bounds in one dimension.
	FitContain ImageFitMode = iota
	// FitCover scales the image to cover the bounds completely while
	// preserving aspect ratio. Parts of the image may be clipped.
	FitCover
	// FitStretch scales the image to exactly fill the bounds, potentially
	// distorting the aspect ratio.
	FitStretch
	// FitOriginal uses the image's intrinsic pixel dimensions converted
	// to points (at 72 DPI).
	FitOriginal
)

// NodeType returns NodeImage.
func (img *Image) NodeType() NodeType { return NodeImage }

// Children returns nil because an image is a leaf node.
func (img *Image) Children() []DocumentNode { return nil }

// Style returns the image's visual style.
func (img *Image) Style() Style { return img.ImgStyle }
