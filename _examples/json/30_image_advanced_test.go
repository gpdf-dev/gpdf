package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_30_ImageAdvanced(t *testing.T) {
	imgData := testutil.TestImagePNG(t, 300, 200, color.RGBA{R: 100, G: 149, B: 237, A: 255})
	imgB64 := base64.StdEncoding.EncodeToString(imgData)

	smallImg := testutil.TestImagePNG(t, 80, 50, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	smallB64 := base64.StdEncoding.EncodeToString(smallImg)

	alignImg := testutil.TestImagePNG(t, 100, 60, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	alignB64 := base64.StdEncoding.EncodeToString(alignImg)

	alphaImg := testutil.TestImagePNGAlpha(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	alphaB64 := base64.StdEncoding.EncodeToString(alphaImg)

	gradientImg := testutil.TestImagePNGGradientAlpha(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	gradientB64 := base64.StdEncoding.EncodeToString(gradientImg)

	fileImgData := testutil.TestImagePNG(t, 150, 100, color.RGBA{R: 251, G: 188, B: 4, A: 255})
	fileImgB64 := base64.StdEncoding.EncodeToString(fileImgData)
	_ = testutil.WriteTestImageFile(t, fileImgData, "yellow.png")

	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Advanced Image Features", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "FitMode Comparison", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "FitContain (default):", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "height": "30mm", "fit": "contain"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "FitStretch:", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "height": "30mm", "fit": "stretch"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "FitOriginal:", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "fit": "original"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "FitCover:", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "height": "30mm", "fit": "cover"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Image Alignment", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "text", "content": "AlignLeft", "style": {"size": 9}},
					{"type": "image", "image": {"src": "%s", "width": "30mm", "align": "left"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "AlignCenter", "style": {"size": 9}},
					{"type": "image", "image": {"src": "%s", "width": "30mm", "align": "center"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "AlignRight", "style": {"size": 9}},
					{"type": "image", "image": {"src": "%s", "width": "30mm", "align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "PNG Transparency (Alpha Channel)", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Checkerboard alpha:", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Gradient alpha:", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "File Path Loading", "style": {"size": 14, "bold": true}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Image loaded from file path:", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "50mm"}}
				]}
			]}}
		]
	}`,
		imgB64, imgB64,
		smallB64, imgB64,
		alignB64, alignB64, alignB64,
		alphaB64, gradientB64,
		fileImgB64,
	))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "30_image_advanced.pdf", doc)
}
