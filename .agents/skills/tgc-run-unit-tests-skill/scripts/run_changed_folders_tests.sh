#!/bin/bash
set -e

# Default to current directory if TGC_DIR not set
TGC_DIR="${TGC_DIR:-$(pwd)}"
cd "$TGC_DIR"

echo "Checking for changes in $TGC_DIR..."

# Get changed files (staged and unstaged)
# Also include untracked files to be thorough
CHANGED_FILES=$(git diff --name-only HEAD; git ls-files --others --exclude-standard)

if [ -z "$CHANGED_FILES" ]; then
    echo "No changes detected."
    exit 0
fi

echo "Changed files:"
echo "$CHANGED_FILES"
echo "----------------"

# Extract top level folders
FOLDERS=$(echo "$CHANGED_FILES" | cut -d/ -f1 | sort -u)

for FOLDER in $FOLDERS; do
    # Check if folder exists and is a directory
    if [ -d "$FOLDER" ]; then
        # Check if it has tests
        if find "$FOLDER" -name "*_test.go" -print -quit | grep -q .; then
            echo "🚀 Running unit tests for $FOLDER..."
            make test-local TEST="./$FOLDER/..."
        else
            echo "ℹ️ No tests found in $FOLDER, skipping."
        fi
    else
        echo "ℹ️ $FOLDER is not a directory, skipping."
    fi
done
