#!/bin/bash

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <input_file> <output_file>"
  exit 1
fi

input_file="$1"
output_file="$2"

jq_filter="pick(.[].episodeId, .[].seriesId, .[].date, .[].eventType, .[].downloadId? // empty, .[].id, .[].data.downloadClient? // empty, .[].data.downloadClientName? // empty)"

# Use jq to filter the JSON and write to the output file
jq "$jq_filter" "$input_file" > "$output_file"