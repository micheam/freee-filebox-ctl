#!/usr/bin/env bash
set -e

#
#/ Usage: download_freeeapi_schema [-h] [-o OUTPUT_FILE]
#/
#/ Options:
#/   -h               show this message.
#/   -o OUTPUT_FILE   specify output file (default: STDOUT).
#/
#/ Examples:
#/    $ download_freeeapi_schema -h
#/    $ download_freeeapi_schema -o ./freeeapi/api-schema.json
#

usage() {
    grep '^#/' "${0}" | cut -c 3-
    echo ""
    exit 1
}

OUTPUT_FILE="/dev/stdout"
SCHEMA_URL="https://raw.githubusercontent.com/freee/freee-api-schema/master/v2020_06_15/open-api-3/api-schema.json"

function _main {
  echo "Downloading freeeapi OpenAPI schema..." >&2
  local TEMP_FILE
  TEMP_FILE=$(mktemp)
  curl -sSL "${SCHEMA_URL}" -o "${TEMP_FILE}"

  echo "Adjusting schema..." >&2
  # Remove application/x-www-form-urlencoded content type entries that have
  # problematic deep path references (e.g., #/paths/~1api~11~1...)
  # These cause "unexpected reference depth" errors in oapi-codegen
  jq '
    # Walk through all paths and remove problematic application/x-www-form-urlencoded references
    walk(
      if type == "object" and has("requestBody") and .requestBody.content then
        .requestBody.content |=
          if has("application/x-www-form-urlencoded") and
             .["application/x-www-form-urlencoded"].schema."$ref" and
             (.["application/x-www-form-urlencoded"].schema."$ref" | startswith("#/paths/~1")) then
            del(.["application/x-www-form-urlencoded"])
          else
            .
          end
      else
        .
      end
    )
  ' "${TEMP_FILE}" > "${OUTPUT_FILE}"

  rm -f "${TEMP_FILE}"
  echo "Schema saved to ${OUTPUT_FILE}" >&2
}

while getopts "ho:" opt; do
  case $opt in
    h) usage ;;
    o) OUTPUT_FILE="$OPTARG" ;;
    ?) usage ;;
  esac
done

shift $((OPTIND -1))

_main "$@"
