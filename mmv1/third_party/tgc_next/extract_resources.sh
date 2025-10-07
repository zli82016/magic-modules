#!/bin/bash

# This script reads a file containing Terraform resource map entries (like the
# provided resources.txt) and extracts the resource name before the colon.

INPUT_FILE="handwritten_resources.txt"
OUTPUT_FILE="handwritten_resource_list_extracted.txt"

# Check if the input file exists
if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found."
    exit 1
fi

echo "Extracting core resource names from $INPUT_FILE (excluding IAM entries)..."

# awk breakdown:
# - FNR > 0: Ensures we only process content lines (not empty lines)
# - -F:":" : Sets the field separator to a colon. The resource name will be in the first field ($1).
# - print $1: Prints the first field.
# - sed: Uses a regular expression to clean up the output:
#   - s/^[[:space:]]*// : Removes leading whitespace.
#   - s/"//g : Removes all double quote characters (").
# - grep -v: Filters out lines ending with _iam_binding, _iam_member, or _iam_policy.
awk 'FNR > 0 { print $1 }' "$INPUT_FILE" | \
    sed -E 's/^[[:space:]]*//; s/"//g; s/://' | \
    grep -vE '(_iam_policy|_iam_binding|_iam_member)$' > "$OUTPUT_FILE"

echo "Extraction complete. Filtered resource list saved to $OUTPUT_FILE"

exit 0
