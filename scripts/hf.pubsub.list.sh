#!/bin/bash
# List Pub/Sub topics and their subscriptions for the current GCP project
# Usage: hf.pubsub.list.sh [filter_term]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config gcp-project

FILTER="${1:-}"

hf_require_gcloud
hf_require_jq

hf_info "Listing Pub/Sub topics and subscriptions for project: $HF_GCP_PROJECT"
if [[ -n "$FILTER" ]]; then
  hf_info "Filtering by: $FILTER"
fi

# Fetch all topics and subscriptions in just two API calls
TOPICS_JSON=$(gcloud pubsub topics list --project "$HF_GCP_PROJECT" --format="json(name)" 2>/dev/null || echo "[]")
SUBS_JSON=$(gcloud pubsub subscriptions list --project "$HF_GCP_PROJECT" --format="json(name,topic)" 2>/dev/null || echo "[]")

if [[ "$TOPICS_JSON" == "[]" && "$SUBS_JSON" == "[]" ]]; then
  echo "No topics or subscriptions found."
  exit 0
fi

# Use jq to group subscriptions by topic, apply filtering, and format as plain text
jq -r -n --arg filter "$FILTER" \
  --argjson topics "$TOPICS_JSON" \
  --argjson subs "$SUBS_JSON" '
  
  # Group subscriptions by their topic
  reduce $subs[] as $sub (
    # Initialize dictionary with all topics (empty arrays)
    reduce $topics[] as $t ({}; .[$t.name] = []);
    # Add subscription to its topic array
    if .[$sub.topic] != null then
      .[$sub.topic] += [$sub.name]
    else
      .[$sub.topic] = [$sub.name]
    end
  )
  | to_entries | sort_by(.key)
  | .[]
  | .key as $t_full | ($t_full | split("/") | last) as $t_short
  | .value as $s_full_arr
  | (if $filter != "" then
       if ($t_short | contains($filter)) then
         # Topic matches filter, show topic and all its subscriptions
         { show: true, topic: $t_short, subs: ($s_full_arr | map(split("/") | last)) }
       else
         # Topic does not match, check if any subscriptions match
         ($s_full_arr | map(split("/") | last) | map(select(contains($filter)))) as $matched_subs
         | if ($matched_subs | length > 0) then
             { show: true, topic: $t_short, subs: $matched_subs }
           else
             { show: false }
           end
       end
     else
       # No filter, show everything
       { show: true, topic: $t_short, subs: ($s_full_arr | map(split("/") | last)) }
     end)
  | select(.show == true)
  | "\(.topic)",
    (.subs[] | "    \(.)")
'
