#!/bin/bash
# Script to generate provider documentation using tfplugindocs
# Run this when adding new resources/data sources or changing schemas

set -e

echo "Generating provider documentation..."
tfplugindocs generate --provider-name openwrt
echo "Documentation generated successfully."

echo ""
echo "NOTE: If you added custom content to docs (examples, guides, etc.),"
echo "you may need to restore it after generation."
echo "See README.md for customization guidelines."