#!/bin/bash
set -eo pipefail

function html_min() {
    # > {{
    # }} <
    # }} {{
    # > <
    tr -d '\n' \
        | sed -E 's/([>\}\}])[[:space:]]+([<\{\{])/\1\2/g' \
        | tr -s ' '
}

cat > static.go << EOF
package patrol
var indexHTML = \`$(cat index.html | html_min)\`
var stylesCSS = \`$(tailwindcss build | postcss)\`
EOF
