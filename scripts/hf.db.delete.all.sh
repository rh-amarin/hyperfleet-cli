#!/bin/bash
# Delete all records from adapter_statuses, node_pools and clusters tables
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config db-host db-port db-name db-user

hf_require_psql

[[ -z "$HF_DB_HOST" ]] && hf_die "DB host not configured. Run hf.db.config.sh to set up database connection."
[[ -z "$HF_DB_USER" ]] && hf_die "DB user not configured. Run hf.db.config.sh to set up database connection."
[[ -z "$HF_DB_NAME" ]] && hf_die "DB name not configured. Run hf.db.config.sh to set up database connection."

export PGHOST="$HF_DB_HOST"
export PGPORT="$HF_DB_PORT"
export PGDATABASE="$HF_DB_NAME"
export PGUSER="$HF_DB_USER"
[[ -n "$HF_DB_PASSWORD" ]] && export PGPASSWORD="$HF_DB_PASSWORD"

tables=(adapter_statuses node_pools clusters)

echo -e "${BOLD}Current record counts:${NC}"
echo ""
for table in "${tables[@]}"; do
    count=$(psql -t -A -c "SELECT COUNT(*) FROM $table" 2>/dev/null || echo "error")
    printf "  %-20s %s\n" "$table" "$count"
done
echo ""

echo -e "${RED}${BOLD}WARNING: This will delete ALL records from the tables above!${NC}"
echo ""
read -p "Type 'yes' to confirm: " confirmation

if [[ "$confirmation" != "yes" ]]; then
    hf_info "Operation cancelled"
    exit 0
fi

echo ""
hf_info "Deleting all records..."

for table in "${tables[@]}"; do
    result=$(psql -c "DELETE FROM $table" 2>&1)
    exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        hf_info "$table: $result"
    else
        hf_error "$table: $result"
        exit 1
    fi
done

hf_info "Done"
