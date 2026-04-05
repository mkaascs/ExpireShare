#!/bin/bash
# scripts/gen_mocks.sh
set -e

MOCKS_DIR="internal/mocks"
INTERFACES_DIR="internal/domain/interfaces"

mkdir -p "$MOCKS_DIR"

echo "scanning interfaces in $INTERFACES_DIR..."

find "$INTERFACES_DIR" -name "*.go" | while read -r file; do
    if ! grep -q "interface" "$file"; then
        continue
    fi

    rel_path="${file#"$INTERFACES_DIR"/}"
    dir=$(dirname "$rel_path")
    base=$(basename "$file" .go)

    pkg="mocks"
    mkdir -p "$MOCKS_DIR"
    output_file="$MOCKS_DIR/${base}_mock.go"

    echo "SUCCESS: $rel_path → $output_file (pkg: $pkg)"

    mockgen \
        -source="$file" \
        -destination="$output_file" \
        -package="$pkg" \
        -imports="context=context" \
        -self_package="expire-share/internal/mocks/$(echo "$dir" | sed 's/\//_/g')" \
        2>/dev/null || {
            echo "WARNING: no interfaces found in $file"
        }
done

echo "all mocks generated in $MOCKS_DIR/"