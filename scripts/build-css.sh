#!/bin/sh
set -e

PATH="$PATH:$(dirname $0)/../node_modules/.bin"
mkdir -p dist

# > {{
# }} <
# }} {{
# > <
cat index.html \
    | tr -d '\n' \
    | sed -E 's/([>\}\}])[[:space:]]+([<\{\{])/\1\2/g' \
    | tr -s ' ' > dist/index.html

css=`mktemp`
tailwindcss build \
    | (if test "$NODE_ENV" = "production"; then postcss; else tr -d '`'; fi) > dist/styles.css
