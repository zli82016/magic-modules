#!/bin/bash
set -e

RESOURCE_NAME=$1
GENERATED_TEST_FILE=$2

if [ $# -lt 2 ]; then
    echo "Usage: $0 <ResourceName> <GeneratedTestFilePath>"
    echo "Example: $0 DialogflowCXAgent test/services/dialogflowcx/dialogflowcx_agent_generated_test.go"
    exit 1
fi

TGC_DIR="${TGC_DIR:-$(pwd)}" # Assume running from downstream root or provided

# Find latest metadata file (skipping empty or very small files)
METADATA_FILE=$(find "$TGC_DIR/test" -name "tests_metadata_*.json" -maxdepth 1 -size +1k | sort -r | head -n 1)

if [ -z "$METADATA_FILE" ]; then
    echo "Error: No tests_metadata_*.json found in $TGC_DIR/test/"
    exit 1
fi

echo "Using metadata file: $METADATA_FILE"

# Extract test names from metadata
echo "Extracting tests from metadata for $RESOURCE_NAME..."
EXPECTED_TESTS=$(grep -E "\"test_name\": \"TestAcc${RESOURCE_NAME}" "$METADATA_FILE" | sed -E 's/.*"test_name": "([^"]+)".*/\1/' | sort -u)

if [ -z "$EXPECTED_TESTS" ]; then
    echo "No tests found for $RESOURCE_NAME in metadata."
    exit 0
fi

echo "Checking against $GENERATED_TEST_FILE..."
MISSING=0
for TEST in $EXPECTED_TESTS; do
    if ! grep -q "$TEST" "$GENERATED_TEST_FILE"; then
        echo "❌ Missing test: $TEST"
        MISSING=$((MISSING+1))
    fi
done

if [ $MISSING -eq 0 ]; then
    echo "✅ All tests present in generated file."
else
    echo "Total missing tests: $MISSING"
    exit 1
fi
