package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"os"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_30_ImageAdvanced(t *testing.T) {
	// Create test images and encode as base64 for JSON embedding.
	imgData := testutil.TestImagePNG(t, 200, 120, color.RGBA{R: 100, G: 149, B: 237, A: 255})
	imgB64 := base64.StdEncoding.EncodeToString(imgData)

	alphaImg := testutil.TestImagePNGAlpha(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	alphaB64 := base64.StdEncoding.EncodeToString(alphaImg)

	gradientImg := testutil.TestImagePNGGradientAlpha(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	gradientB64 := base64.StdEncoding.EncodeToString(gradientImg)

	// Write an image to a file for file path loading.
	fileImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	tmpDir := t.TempDir()
	filePath := tmpDir + "/green.png"
	if err := os.WriteFile(filePath, fileImg, 0644); err != nil {
		t.Fatal(err)
	}

	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "Advanced Image Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "Advanced Image Features (JSON)", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},

			{"row": {"cols": [
				{"span": 12, "text": "FitMode Comparison", "style": {"size": 14, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "fit: contain (default)", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "height": "30mm", "fit": "contain"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "fit: stretch", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "height": "30mm", "fit": "stretch"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "fit: original", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "fit": "original"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "fit: cover", "style": {"size": 9}},
					{"type": "spacer", "height": "1mm"},
					{"type": "image", "image": {"src": "%s", "width": "60mm", "height": "30mm", "fit": "cover"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},

			{"row": {"cols": [
				{"span": 12, "text": "Image Alignment", "style": {"size": 14, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "text", "content": "align: left", "style": {"size": 9}},
					{"type": "image", "image": {"src": "%s", "width": "30mm", "align": "left"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "align: center", "style": {"size": 9}},
					{"type": "image", "image": {"src": "%s", "width": "30mm", "align": "center"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "align: right", "style": {"size": 9}},
					{"type": "image", "image": {"src": "%s", "width": "30mm", "align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},

			{"row": {"cols": [
				{"span": 12, "text": "PNG Transparency (Alpha Channel)", "style": {"size": 14, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
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
				{"span": 12, "text": "File Path Loading", "style": {"size": 14, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
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
		imgB64, imgB64, imgB64, imgB64,
		imgB64, imgB64, imgB64,
		alphaB64, gradientB64,
		filePath,
	))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "30_image_advanced.pdf", doc)
}
