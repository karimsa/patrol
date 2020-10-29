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

function css_min() {
    if test "$NODE_ENV" = "production"; then
        postcss
    else
        tr -d '`'
    fi
}

cat > static.go << EOF
package patrol

var indexHTML = \`$(cat index.html | html_min)\`
var stylesCSS = \`$(tailwindcss build | css_min)\`
EOF
