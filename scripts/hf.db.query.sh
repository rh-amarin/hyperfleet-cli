#!/bin/bash
# Execute PostgreSQL queries using configured connection
# Usage: hf.db.query.sh <query>
#        hf.db.query.sh -f <file.sql>
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config db-host db-port db-name db-user

hf_require_psql

# Check if we have database configuration
[[ -z "$HF_DB_HOST" ]] && hf_die "DB host not configured. Run hf.db.config.sh to set up database connection."
[[ -z "$HF_DB_USER" ]] && hf_die "DB user not configured. Run hf.db.config.sh to set up database connection."
[[ -z "$HF_DB_NAME" ]] && hf_die "DB name not configured. Run hf.db.config.sh to set up database connection."

# Build connection string
export PGHOST="$HF_DB_HOST"
export PGPORT="$HF_DB_PORT"
export PGDATABASE="$HF_DB_NAME"
export PGUSER="$HF_DB_USER"

# Set password if configured (otherwise psql will prompt or use .pgpass)
[[ -n "$HF_DB_PASSWORD" ]] && export PGPASSWORD="$HF_DB_PASSWORD"

# Parse arguments
if [[ $# -eq 0 ]]; then
    hf_usage "<query> | -f <file.sql>"
    echo "Execute PostgreSQL queries using configured connection"
    echo ""
    echo "Options:"
    echo "  -f <file>     Execute SQL from file"
    echo "  -c <query>    Execute SQL query (default if not -f)"
    echo "  -t            Tuples only mode (no headers)"
    echo "  -A            Unaligned output mode"
    echo "  -q            Quiet mode"
    echo ""
    echo "Examples:"
    echo "  $0 'SELECT * FROM users LIMIT 10'"
    echo "  $0 -f schema.sql"
    echo "  $0 -t -A 'SELECT id FROM users' | while read id; do echo \$id; done"
    exit 1
fi

# Check if it's a file execution
if [[ "$1" == "-f" ]]; then
    [[ -z "$2" ]] && hf_die "Usage: $0 -f <file.sql>"
    [[ ! -f "$2" ]] && hf_die "File not found: $2"
    hf_info "Executing SQL from file: $2"
    psql -f "$2"
elif [[ "$1" == "-c" ]]; then
    # -c flag already provided, pass through
    psql "$@"
else
    # No flag provided, assume it's a query and add -c
    psql -c "$@"
fi
