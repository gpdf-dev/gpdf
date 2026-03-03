package gotemplate_test

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_30_ImageAdvanced(t *testing.T) {
	imgData := testutil.TestImagePNG(t, 300, 200, color.RGBA{R: 100, G: 149, B: 237, A: 255})
	smallImg := testutil.TestImagePNG(t, 80, 50, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	alignImg := testutil.TestImagePNG(t, 100, 60, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	alphaImg := testutil.TestImagePNGAlpha(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	gradientImg := testutil.TestImagePNGGradientAlpha(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	fileImgData := testutil.TestImagePNG(t, 150, 100, color.RGBA{R: 251, G: 188, B: 4, A: 255})
	_ = testutil.WriteTestImageFile(t, fileImgData, "yellow.png")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.FitModeLabel}}", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ContainLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.ImgB64}}", "width": "60mm", "height": "30mm", "fit": "contain"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.StretchLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.ImgB64}}", "width": "60mm", "height": "30mm", "fit": "stretch"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.OriginalLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.SmallImgB64}}", "fit": "original"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.CoverLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.ImgB64}}", "width": "60mm", "height": "30mm", "fit": "cover"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.AlignLabel}}", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.AlignLeftLabel}}", "style": {"size": 9}},
					{"type": "image", "image": {"src": "{{.AlignImgB64}}", "width": "30mm", "align": "left"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.AlignCenterLabel}}", "style": {"size": 9}},
					{"type": "image", "image": {"src": "{{.AlignImgB64}}", "width": "30mm", "align": "center"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.AlignRightLabel}}", "style": {"size": 9}},
					{"type": "image", "image": {"src": "{{.AlignImgB64}}", "width": "30mm", "align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.TransparencyLabel}}", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.CheckerLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.AlphaB64}}", "width": "60mm"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.GradientLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.GradientB64}}", "width": "60mm"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.FilePathLabel}}", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.FileLoadLabel}}", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "{{.FileImgB64}}", "width": "50mm"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":             "Advanced Image Features",
		"FitModeLabel":      "FitMode Comparison",
		"ContainLabel":      "FitContain (default):",
		"StretchLabel":      "FitStretch:",
		"OriginalLabel":     "FitOriginal:",
		"CoverLabel":        "FitCover:",
		"AlignLabel":        "Image Alignment",
		"AlignLeftLabel":    "AlignLeft",
		"AlignCenterLabel":  "AlignCenter",
		"AlignRightLabel":   "AlignRight",
		"TransparencyLabel": "PNG Transparency (Alpha Channel)",
		"CheckerLabel":      "Checkerboard alpha:",
		"GradientLabel":     "Gradient alpha:",
		"FilePathLabel":     "File Path Loading",
		"FileLoadLabel":     "Image loaded from file path:",
		"ImgB64":            base64.StdEncoding.EncodeToString(imgData),
		"SmallImgB64":       base64.StdEncoding.EncodeToString(smallImg),
		"AlignImgB64":       base64.StdEncoding.EncodeToString(alignImg),
		"AlphaB64":          base64.StdEncoding.EncodeToString(alphaImg),
		"GradientB64":       base64.StdEncoding.EncodeToString(gradientImg),
		"FileImgB64":        base64.StdEncoding.EncodeToString(fileImgData),
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "30_image_advanced.pdf", doc)
}
