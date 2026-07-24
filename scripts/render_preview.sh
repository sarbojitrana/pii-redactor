#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

INPUT_DOCX=$1
OUTPUT_DIR=$2

if [ -z "$INPUT_DOCX" ] || [ -z "$OUTPUT_DIR" ]; then
    echo "Usage: ./render_preview.sh <input.docx> <output_directory>"
    exit 1
fi

if [ ! -f "$INPUT_DOCX" ]; then
    echo "ERROR: Input file '$INPUT_DOCX' not found."
    exit 1
fi

# 1. Dependency Resolution
# Detect the available headless office suite. 
# Linux typically uses 'libreoffice' or 'soffice', macOS uses the full path.
if command -v libreoffice >/dev/null 2>&1; then
    OFFICE_BIN="libreoffice"
elif command -v soffice >/dev/null 2>&1; then
    OFFICE_BIN="soffice"
elif [ -f "/Applications/LibreOffice.app/Contents/MacOS/soffice" ]; then
    OFFICE_BIN="/Applications/LibreOffice.app/Contents/MacOS/soffice"
else
    echo "ERROR: LibreOffice/soffice is not installed or not in the PATH."
    echo "This script requires LibreOffice to render the OOXML layout into a PDF."
    exit 1
fi

echo "Rendering visual preview using $OFFICE_BIN..."

# 2. Headless PDF Generation
# --headless: Prevents the GUI from launching (critical for CLI/server environments)
# --convert-to pdf: Invokes the PDF export filter
# --outdir: Specifies the destination directory
$OFFICE_BIN --headless \
            --invisible \
            --nodefault \
            --nologo \
            --convert-to pdf \
            --outdir "$OUTPUT_DIR" \
            "$INPUT_DOCX"

# Extract the base filename to locate the generated PDF
BASENAME=$(basename "$INPUT_DOCX" .docx)
GENERATED_PDF="$OUTPUT_DIR/$BASENAME.pdf"

if [ -f "$GENERATED_PDF" ]; then
    echo "SUCCESS: PDF preview generated at '$GENERATED_PDF'."
else
    echo "ERROR: PDF generation failed."
    exit 1
fi

# 3. (Optional) JPEG Generation
# If you strictly need a JPEG for visual QA (e.g., embedding in a web dashboard), 
# pdftoppm (from poppler-utils) is the most robust way to rasterize the PDF.
if command -v pdftoppm >/dev/null 2>&1; then
    echo "Rasterizing PDF to JPEG..."
    pdftoppm -jpeg -r 150 -singlefile "$GENERATED_PDF" "$OUTPUT_DIR/${BASENAME}_preview"
    echo "SUCCESS: JPEG preview generated at '$OUTPUT_DIR/${BASENAME}_preview.jpg'."
else
    echo "NOTE: 'pdftoppm' not found. Skipping JPEG rasterization. PDF is ready."
fi