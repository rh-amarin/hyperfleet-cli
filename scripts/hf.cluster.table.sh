#!/bin/bash
# List all clusters in table format
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

hf_require_jq

# Color codes
GREEN=$(printf '\033[32m')
RED=$(printf '\033[31m')
YELLOW=$(printf '\033[33m')
RESET=$(printf '\033[0m')

hf_get "/clusters" | jq -r '
  # Collect all unique condition types across all clusters
  ([.items[].status.conditions[].type] | unique) as $types |

  # Store items for later use
  .items as $items |

  # Build header row
  (["ID", "NAME", "GEN"] + $types | @tsv),

  # Build separator row
  (["---", "---", "---"] + ($types | map("---")) | @tsv),

  # Build data rows
  ($items[] | . as $cluster |
    [.id, .name, (.generation // 0 | tostring)] + [$types[] as $t | ($cluster.status.conditions | map(select(.type == $t)) | .[0].status |
      if . == "True" then "\u0001"
      elif . == "False" then "\u0002"
      elif . == "Unknown" then "\u0003"
      elif . == "" or . == null then "-"
      else . end)]
    | @tsv
  )
' | awk -v green="$GREEN" -v red="$RED" -v yellow="$YELLOW" -v reset="$RESET" '
BEGIN { FS = "\t" }
{
  row[NR] = $0
  n = split($0, f, "\t")
  if (n > ncols) ncols = n
  for (i = 1; i <= n; i++) {
    w = length(f[i])
    if (w > cw[i]) cw[i] = w
  }
}
END {
  for (r = 1; r <= NR; r++) {
    n = split(row[r], f, "\t")
    for (i = 1; i <= ncols; i++) {
      cell = (i <= n) ? f[i] : ""
      if      (cell == "\001") display = green "●" reset
      else if (cell == "\002") display = red   "●" reset
      else if (cell == "\003") display = yellow "●" reset
      else                     display = cell
      pad = cw[i] - length(cell)
      if (i < ncols) printf "%s%*s  ", display, pad, ""
      else           printf "%s", display
    }
    printf "\n"
  }
}'
