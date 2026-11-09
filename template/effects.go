package template

import (
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// --- Rounded corners ---

// WithTextRadius rounds the corners of a text element's background
// ([BgColor]) and uniform border ([WithTextBorder]).
func WithTextRadius(v document.Value) TextOption {
	return func(s *document.Style) { s.BorderRadius = v }
}

// WithImageRadius clips the image to a rectangle with rounded corners of
// the given radius. A radius of half the image size produces a circle
// (for square images) or a pill shape.
func WithImageRadius(v document.Value) ImageOption {
	return func(cfg *imageConfig) { cfg.radius = v }
}

// WithBoxRadius rounds the corners of the box background and of a uniform
// border. The radius is clamped to half the box's width and height.
// Mixed (non-uniform) borders keep square corners.
func WithBoxRadius(v document.Value) BoxOption {
	return func(cfg *boxConfig) { cfg.radius = v }
}

// --- Opacity ---

// WithTextOpacity sets the opacity of a text element (0.0 transparent –
// 1.0 opaque), including its background and decorations. This is useful
// for watermarks and overlay text that must not hide what lies beneath.
func WithTextOpacity(v float64) TextOption {
	return func(s *document.Style) { s.Opacity = v }
}

// WithImageOpacity sets the opacity of an image (0.0 transparent – 1.0
// opaque).
func WithImageOpacity(v float64) ImageOption {
	return func(cfg *imageConfig) { cfg.opacity = v }
}

// WithBoxOpacity sets the opacity of a box and everything inside it
// (0.0 transparent – 1.0 opaque). A child with its own opacity replaces
// the box's value rather than multiplying it.
func WithBoxOpacity(v float64) BoxOption {
	return func(cfg *boxConfig) { cfg.opacity = v }
}

// --- Links ---

// WithTextLink makes a text element (or a [RichTextBuilder.Span]) a
// hyperlink to url. It does not change the text's appearance; combine it
// with [TextColor] and [Underline], or use [ColBuilder.Link].
func WithTextLink(url string) TextOption {
	return func(s *document.Style) { s.Link = url }
}

// WithImageLink makes the image a hyperlink to url.
func WithImageLink(url string) ImageOption {
	return func(cfg *imageConfig) { cfg.link = url }
}

// WithBoxLink makes the whole box area a hyperlink to url.
func WithBoxLink(url string) BoxOption {
	return func(cfg *boxConfig) { cfg.link = url }
}

// --- Page background ---

// BackgroundOption configures a page background image set with
// [WithBackgroundImage].
type BackgroundOption func(*document.PageBackground)

// BackgroundFit sets how the background image is scaled to the page.
// The default is [document.FitCover] (fill the page, cropping overflow).
func BackgroundFit(mode document.ImageFitMode) BackgroundOption {
	return func(bg *document.PageBackground) { bg.FitMode = mode }
}

// BackgroundOpacity sets the opacity of the page background (0.0
// transparent – 1.0 opaque), e.g. for a faint letterhead or watermark.
func BackgroundOpacity(v float64) BackgroundOption {
	return func(bg *document.PageBackground) { bg.Opacity = v }
}

// WithBackgroundImage draws an image (PNG, JPEG or SVG) beneath the content of
// every page, centered on the page and ignoring page margins. By default
// the image covers the whole page; use [BackgroundFit] to change that.
//
//	doc := template.New(
//	    template.WithBackgroundImage(letterhead, template.BackgroundFit(document.FitStretch)),
//	)
func WithBackgroundImage(data []byte, opts ...BackgroundOption) Option {
	return func(c *Config) {
		bg := c.ensureBackground()
		format := detectImageFormat(data)
		w, h := extractImageDimensions(data, format)
		if format == document.ImageSVG && w == 0 {
			return // unparseable SVG: no background image
		}
		bg.Image = &document.ImageSource{Data: data, Format: format, Width: w, Height: h}
		bg.FitMode = document.FitCover
		for _, opt := range opts {
			opt(bg)
		}
	}
}

// WithBackgroundColor fills every page with a solid color beneath all
// content. It can be combined with [WithBackgroundImage]; the image is
// drawn on top of the color.
func WithBackgroundColor(col pdf.Color) Option {
	return func(c *Config) {
		c.ensureBackground().Color = &col
	}
}

// ensureBackground returns the config's page background, creating a copy
// so that options never mutate a background shared with another Config.
func (c *Config) ensureBackground() *document.PageBackground {
	bg := &document.PageBackground{}
	if c.Background != nil {
		*bg = *c.Background
	}
	c.Background = bg
	return bg
}
