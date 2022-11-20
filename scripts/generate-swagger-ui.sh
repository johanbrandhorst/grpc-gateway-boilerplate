#!/bin/sh

set -e

[[ -z "$SWAGGER_UI_VERSION" ]] && echo "missing \$SWAGGER_UI_VERSION" && exit 1

SWAGGER_UI_GIT="https://github.com/swagger-api/swagger-ui.git"
CACHE_DIR="./.cache/swagger-ui/$SWAGGER_UI_VERSION"
GEN_DIR="./third_party/OpenAPI"

escape_str() {
  echo "$1" | sed -e 's/[]\/$*.^[]/\\&/g'
}

# do caching if there's no cache yet
if [[ ! -d "$CACHE_DIR" ]]; then
  mkdir -p "$CACHE_DIR"
  tmp="$(mktemp -d)"
  git clone --depth 1 --branch "$SWAGGER_UI_VERSION" "$SWAGGER_UI_GIT" "$tmp"
  cp -r "$tmp/dist/"* "$CACHE_DIR"
  cp -r "$tmp/LICENSE" "$CACHE_DIR"
  rm -rf "$tmp"
fi

# populate swagger.json
tmp="    urls: ["
for i in $(find "$GEN_DIR" -name "*.swagger.json"); do
  escaped_gen_dir="$(escape_str "$GEN_DIR/")"
  path="$(echo $i | sed -e "s/$escaped_gen_dir//g")"
  tmp="$tmp{\"url\":\"$path\",\"name\":\"$path\"},"
done
# delete last characters from $tmp
tmp=$(echo "$tmp" | sed 's/.$//')
tmp="$tmp],"

# recreate swagger-ui, delete all except swagger.json
find "$GEN_DIR" -type f -not -name "*.swagger.json" -delete
mkdir -p "$GEN_DIR"
cp -r "$CACHE_DIR/"* "$GEN_DIR"

# replace the default URL
line="$(cat "$GEN_DIR/swagger-initializer.js" | grep -n "url" | cut -f1 -d:)"
escaped_tmp="$(escape_str "$tmp")"
sed -i'' -e "$line s/^.*$/$escaped_tmp/" "$GEN_DIR/swagger-initializer.js"
rm -f "$GEN_DIR/swagger-initializer.js-e"
