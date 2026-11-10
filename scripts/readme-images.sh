#!/usr/bin/env bash
# Regenerates the README hero thumbnails in docs/assets/readme/.
#
# Requirements: pdftoppm (poppler) and NotoSansJP-Regular.ttf in the
# workspace root (same location the CJK example tests use).
set -euo pipefail

cd "$(dirname "$0")/.."
OUT=docs/assets/readme
GOLDEN=_examples/testdata/golden
WIDTH=600 # rendered at 2x for a ~300px display width
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

command -v pdftoppm >/dev/null || { echo "pdftoppm not found (brew install poppler)" >&2; exit 1; }
mkdir -p "$OUT"

GPDF_SHOWCASE_OUT="$TMP" go test -count=1 -run TestShowcase_InvoiceJA ./_examples/showcase/
[ -f "$TMP/invoice_ja.pdf" ] || { echo "showcase PDF was not generated (missing font?)" >&2; exit 1; }

render() { # <pdf> <name>
	pdftoppm -png -singlefile -f 1 -l 1 -scale-to-x "$WIDTH" -scale-to-y -1 "$1" "$OUT/$2"
}

render "$GOLDEN/18_invoice.pdf" invoice
render "$TMP/invoice_ja.pdf" invoice_ja
render "$GOLDEN/19_report.pdf" report
render "$GOLDEN/29_qr_barcode_invoice.pdf" qr_barcode

pngs=$(ls "$OUT"/*.png | sed "s|^|$PWD/|" | paste -sd: -)
GPDF_FRAME_PNGS="$pngs" go test -count=1 -run TestShowcase_Frame ./_examples/showcase/

ls -la "$OUT"
