#!/bin/bash
# Delete rows from PostgreSQL table
# Usage: hf.db.delete.sh <table> [id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config db-host db-port db-name db-user

hf_require_psql

# Check if we have database configuration
[[ -z "$HF_DB_HOST" ]] && hf_die "DB host not configured. Run hf.db.config.sh to set up database connection."
[[ -z "$HF_DB_USER" ]] && hf_die "DB user not configured. Run hf.db.config.sh to set up database connection."
[[ -z "$HF_DB_NAME" ]] && hf_die "DB name not configured. Run hf.db.config.sh to set up database connection."

# Parse arguments
if [[ $# -eq 0 ]]; then
    hf_usage "<table> [id]"
    echo "Delete rows from a PostgreSQL table"
    echo ""
    echo "Arguments:"
    echo "  table         Table name (required)"
    echo "  id            Row ID to delete (optional)"
    echo ""
    echo "Examples:"
    echo "  $0 users 123           # Delete user with id=123"
    echo "  $0 sessions            # Delete all rows from sessions table"
    exit 1
fi

table="$1"
row_id="${2:-}"

# Build connection string
export PGHOST="$HF_DB_HOST"
export PGPORT="$HF_DB_PORT"
export PGDATABASE="$HF_DB_NAME"
export PGUSER="$HF_DB_USER"
[[ -n "$HF_DB_PASSWORD" ]] && export PGPASSWORD="$HF_DB_PASSWORD"

# Build DELETE query and confirmation message
if [[ -n "$row_id" ]]; then
    # Delete specific row
    query="DELETE FROM $table WHERE id = '$row_id'"
    echo -e "${YELLOW}You are about to delete:${NC}"
    echo "  Table: $table"
    echo "  Row ID: $row_id"
    echo ""

    # Show the row that will be deleted
    echo -e "${BOLD}Row to be deleted:${NC}"
    psql -c "SELECT * FROM $table WHERE id = '$row_id'" 2>/dev/null || hf_warn "Could not preview row (table might not exist or id not found)"
    echo ""
else
    # Delete all rows
    query="DELETE FROM $table"

    # Count rows
    row_count=$(psql -t -A -c "SELECT COUNT(*) FROM $table" 2>/dev/null || echo "unknown")

    echo -e "${RED}${BOLD}WARNING: You are about to delete ALL rows!${NC}"
    echo "  Table: $table"
    echo "  Row count: $row_count"
    echo ""
fi

# Ask for confirmation
read -p "Are you sure you want to proceed? (yes/no): " confirmation

if [[ "$confirmation" != "yes" ]]; then
    hf_info "Delete operation cancelled"
    exit 0
fi

# Execute DELETE
echo ""
hf_info "Executing DELETE..."
result=$(psql -c "$query" 2>&1)
exit_code=$?

if [[ $exit_code -eq 0 ]]; then
    echo "$result"
    hf_info "Delete completed successfully"
else
    hf_error "Delete failed:"
    echo "$result"
    exit 1
fi
