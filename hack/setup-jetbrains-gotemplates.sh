#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOKOALA_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${GOKOALA_ROOT}"

traverse_templates() {
  # HTML templates
  find * -type f -iname "*.go.html" -print0 | while IFS= read -r -d '' file; do
    echo "Processing file: $file"
    echo "<file url=\"file://\$PROJECT_DIR\$/$file\" dialect=\"HTML\" />" >> ".idea/templateLanguages.xml"
  done
  # JSON templates
  find * -type f -iname "*.go.json" -print0 | while IFS= read -r -d '' file; do
    echo "Processing file: $file"
    echo "<file url=\"file://\$PROJECT_DIR\$/$file\" dialect=\"JSON\" />" >> ".idea/templateLanguages.xml"
  done
  # TileJSON templates
  find * -type f -iname "*.go.tilejson" -print0 | while IFS= read -r -d '' file; do
    echo "Processing file: $file"
    echo "<file url=\"file://\$PROJECT_DIR\$/$file\" dialect=\"JSON\" />" >> ".idea/templateLanguages.xml"
  done
  # XML templates
  find * -type f -iname "*.go.xml" -print0 | while IFS= read -r -d '' file; do
    echo "Processing file: $file"
    echo "<file url=\"file://\$PROJECT_DIR\$/$file\" dialect=\"XML\" />" >> ".idea/templateLanguages.xml"
  done
}

mkdir -p ".idea/"

cat << EOF > ".idea/templateLanguages.xml"
<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="TemplateDataLanguageMappings">
EOF

traverse_templates

cat << EOF >> ".idea/templateLanguages.xml"
  </component>
</project>
EOF
