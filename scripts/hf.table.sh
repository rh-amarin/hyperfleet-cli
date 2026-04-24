#!/bin/bash
# List all clusters with their nodepools in a combined table
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

hf_require_jq

GREEN=$(printf '\033[32m')
RED=$(printf '\033[31m')
YELLOW=$(printf '\033[33m')
RESET=$(printf '\033[0m')

CLUSTERS_JSON=$(hf_get "/clusters")

ALL_CLUSTER_IDS=$(echo "$CLUSTERS_JSON" | jq -r '.items[].id')

COMBINED=""
for CID in $ALL_CLUSTER_IDS; do
  NP_JSON=$(hf_get "/clusters/${CID}/nodepools" 2>/dev/null)
  CLUSTER_STATUSES=$(hf_get "/clusters/${CID}/statuses" 2>/dev/null)
  CLUSTER_AC=$(echo "$CLUSTER_STATUSES" | jq -r '.items // [] | length')
  CLUSTER_ROW=$(echo "$CLUSTERS_JSON" | jq -r --arg cid "$CID" --arg ac "${CLUSTER_AC:-0}" '
    .items[] | select(.id == $cid) |
    ((.status.conditions // [] | map(select(.type == "Ready")) | .[0].status) // "-") as $ready |
    ["C", .id, .name, (.generation // 0 | tostring), $ready, $ac,
     (.status.conditions // [] | map(select(.type != "Ready")) | map(.type + "=" + .status) | join(","))] | @tsv
  ')
  NP_IDS=$(echo "$NP_JSON" | jq -r '(.items // [])[].id' 2>/dev/null)
  NP_ROWS=""
  for NPID in $NP_IDS; do
    NP_STATUSES=$(hf_get "/clusters/${CID}/nodepools/${NPID}/statuses" 2>/dev/null)
    NP_AC=$(echo "$NP_STATUSES" | jq -r '.items // [] | length')
    ROW=$(echo "$NP_JSON" | jq -r --arg npid "$NPID" --arg ac "${NP_AC:-0}" '
      .items[] | select(.id == $npid) |
      ((.status.conditions // [] | map(select(.type == "Ready")) | .[0].status) // "-") as $ready |
      ["N", ("  " + .id), ("  " + .name), (.generation // 0 | tostring), $ready, $ac,
       (.status.conditions // [] | map(select(.type != "Ready")) | map(.type + "=" + .status) | join(","))] | @tsv
    ')
    NP_ROWS+="${ROW}"$'\n'
  done
  COMBINED+="${CLUSTER_ROW}"$'\n'
  if [[ -n "$NP_ROWS" ]]; then
    COMBINED+="${NP_ROWS}"
  fi
done

ALL_COND_TYPES=$(echo "$COMBINED" | awk -F'\t' '{
  split($7, pairs, ",")
  for (i in pairs) {
    split(pairs[i], kv, "=")
    if (kv[1] != "") types[kv[1]] = 1
  }
} END {
  n = asorti(types, sorted)
  for (i = 1; i <= n; i++) printf "%s\n", sorted[i]
}')

COND_HEADER=""
COND_SEP=""
while IFS= read -r t; do
  [[ -z "$t" ]] && continue
  COND_HEADER+="\t${t}"
  COND_SEP+="\t---"
done <<< "$ALL_COND_TYPES"

{
  printf "ID\tNAME\tGEN\tREADY\tADAPTERS%s\n" "$COND_HEADER"
  printf "---\t---\t---\t-----\t--------%s\n" "$COND_SEP"

  while IFS=$'\t' read -r kind id name gen ready adapters conds; do
    [[ -z "$kind" ]] && continue
    if [[ "$ready" == "True" ]]; then
      READY_CELL="\x01"
    elif [[ "$ready" == "False" ]]; then
      READY_CELL="\x02"
    elif [[ "$ready" == "Unknown" ]]; then
      READY_CELL="\x03"
    else
      READY_CELL="-"
    fi
    ROW="${id}\t${name}\t${gen}\t${READY_CELL}\t${adapters}"
    while IFS= read -r t; do
      [[ -z "$t" ]] && continue
      VAL=$(echo "$conds" | tr ',' '\n' | awk -F= -v t="$t" '$1 == t {print $2}')
      if [[ "$VAL" == "True" ]]; then
        ROW+="\t\x01"
      elif [[ "$VAL" == "False" ]]; then
        ROW+="\t\x02"
      elif [[ "$VAL" == "Unknown" ]]; then
        ROW+="\t\x03"
      else
        ROW+="\t-"
      fi
    done <<< "$ALL_COND_TYPES"
    printf "%b\n" "$ROW"
  done <<< "$COMBINED"
} | awk -v green="$GREEN" -v red="$RED" -v yellow="$YELLOW" -v reset="$RESET" '
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
