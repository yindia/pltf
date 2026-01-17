#!/bin/bash

# Script to generate terraform-docs for all modules
# Usage: ./scripts/generate-terraform-docs.sh

set -e  # Exit on error

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the project root (one level up from scripts/)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MODULES_DIR="$PROJECT_ROOT/modules"

# Check if modules directory exists
if [ ! -d "$MODULES_DIR" ]; then
    echo "Error: Modules directory not found at $MODULES_DIR"
    exit 1
fi

# Check if terraform-docs is installed
if ! command -v terraform-docs &> /dev/null; then
    echo "Error: terraform-docs is not installed"
    echo "Please install it from: https://github.com/terraform-docs/terraform-docs"
    exit 1
fi

# Counter for processed modules
success_count=0
fail_count=0

echo "Generating terraform-docs for all modules in $MODULES_DIR"
echo "=================================================="
echo ""

# Find all directories in modules/ and process each one
for module_dir in "$MODULES_DIR"/*/; do
    # Check if it's actually a directory (glob will still expand even if no matches)
    [ -d "$module_dir" ] || continue
    
    # Get just the module name (last part of path)
    module_name=$(basename "$module_dir")
    
    # Skip if it's not a valid module directory (e.g., skip files)
    [ -f "$module_dir/module.yaml" ] || continue
    
    echo "Processing module: $module_name"
    
    # Run terraform-docs command
    if terraform-docs markdown table \
        --output-file $module_name.md \
        --output-mode inject \
        "$module_dir"; then
        echo "✓ Successfully generated docs for $module_name"
        cp "$module_dir/$module_name.md" docs/references/modules/
        ((success_count++))
    else
        echo "✗ Failed to generate docs for $module_name"
        ((fail_count++))
    fi
    echo ""
done

echo "=================================================="
echo "Summary:"
echo "  Success: $success_count"
echo "  Failed:  $fail_count"
echo "  Total:   $((success_count + fail_count))"

# Exit with error if any modules failed
if [ $fail_count -gt 0 ]; then
    exit 1
fi


