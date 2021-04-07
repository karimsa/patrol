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

PATH="$PATH:$(dirname $0)/../node_modules/.bin"

mkdir -p dist

cat index.html | html_min > dist/index.html

css=`mktemp`
tailwindcss build | css_min > dist/styles.css
