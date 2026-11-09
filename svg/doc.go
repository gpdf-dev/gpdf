// Package svg converts SVG vector graphics into PDF content-stream
// operators, using only the Go standard library.
//
// It is used by gpdf to draw SVG images as resolution-independent vector
// graphics (PDF Form XObjects) instead of rasterizing them. The supported
// subset covers what icon sets, logos and charts exported from common tools
// typically use:
//
//   - Shapes: path (all commands, including arcs), rect (with rx/ry),
//     circle, ellipse, line, polyline, polygon
//   - Structure: svg (viewBox, width/height, preserveAspectRatio "none" or
//     the default xMidYMid meet), g, defs, symbol, use, switch
//   - Painting: fill, stroke, stroke-width, stroke-linecap,
//     stroke-linejoin, stroke-miterlimit, stroke-dasharray,
//     stroke-dashoffset, fill-rule, opacity, fill-opacity, stroke-opacity,
//     display, visibility, color/currentColor
//   - Styling: presentation attributes, the style attribute, and <style>
//     sheets with simple tag, .class and #id selectors
//   - Transforms: matrix, translate, scale, rotate, skewX, skewY
//   - Colors: #rgb, #rrggbb, #rgba, #rrggbbaa, rgb(), rgba() and the CSS
//     named colors
//
// Not supported (silently ignored): text, embedded raster images, clip
// paths, masks, filters, markers and patterns. Gradients are approximated
// with a solid fill using the color of their first stop. Group opacity is
// approximated by multiplying it into the opacity of each descendant.
//
// One SVG user unit (px) is mapped to one PDF point, matching gpdf's 72 DPI
// treatment of raster images; absolute units (mm, cm, in, pt, pc) are
// converted to points.
package svg
