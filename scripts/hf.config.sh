#!/bin/bash
# Central HyperFleet configuration management
# Usage: hf.config.sh [show|set|clear|doctor|bootstrap|env] [args...]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"

SCRIPT_DIR="$(dirname "$(realpath "$0")")"
ENV_LIST_FILE="$HF_CONFIG_DIR/environment-list"
ENV_ACTIVE_FILE="$HF_CONFIG_DIR/environment-active"

# Derive flat property list from registry (used by env activate/count)
_CONFIG_KEYS=()
for _entry in "${HF_CONFIG_REGISTRY[@]}"; do
    _hf_parse "$_entry"
    _CONFIG_KEYS+=("$_HF_E_KEY")
done

# ============================================================================
# Environment helpers
# ============================================================================

_env_exists() {
    [[ -f "$ENV_LIST_FILE" ]] && grep -qx "$1" "$ENV_LIST_FILE" 2>/dev/null
}

_get_active_env() {
    [[ -f "$ENV_ACTIVE_FILE" ]] && cat "$ENV_ACTIVE_FILE"
}

_get_env_prop_count() {
    local env="$1" count=0
    for key in "${_CONFIG_KEYS[@]}"; do
        [[ -f "$HF_CONFIG_DIR/$env.$key" ]] && ((count++))
    done
    echo "$count"
}

# ============================================================================
# show — unified display, optional env-name for override view
# ============================================================================

show_config() {
    local env_name="${1:-}"
    local active
    active=$(_get_active_env)

    echo -e "${BOLD}HyperFleet Configuration${NC}"
    echo ""
    echo "  Config dir: $HF_CONFIG_DIR"
    if [[ -n "$env_name" ]]; then
        _env_exists "$env_name" || hf_die "Environment '$env_name' not found in $ENV_LIST_FILE"
        echo -e "  Environment: ${CYAN}$env_name${NC}"
    fi
    [[ -n "$active" ]] && echo -e "  Active: ${GREEN}$active${NC}"
    echo ""

    local current_section=""
    for entry in "${HF_CONFIG_REGISTRY[@]}"; do
        _hf_parse "$entry"
        local key="$_HF_E_KEY"
        local value="" suffix=""

        # Environment override takes priority when viewing an env
        if [[ -n "$env_name" ]] && [[ -f "$HF_CONFIG_DIR/$env_name.$key" ]]; then
            value=$(<"$HF_CONFIG_DIR/$env_name.$key")
            suffix="  \033[0;36m[$env_name]\033[0m"
        elif [[ -f "$HF_CONFIG_DIR/$key" ]]; then
            value=$(<"$HF_CONFIG_DIR/$key")
        else
            value="$_HF_E_DEFAULT"
        fi

        # Section header on change
        if [[ "$_HF_E_SECTION" != "$current_section" ]]; then
            [[ -n "$current_section" ]] && echo ""
            echo -e "${BOLD}${_HF_E_SECTION}${NC}"
            current_section="$_HF_E_SECTION"
        fi

        # Display value (mask sensitive)
        if [[ "$_HF_E_FLAGS" == *s* ]]; then
            local display
            if [[ -n "$value" ]]; then display="<set>"; else display="<not set>"; fi
            printf "  %-35s %s" "$key" "$display"
        else
            printf "  %-35s %s" "$key" "${value:-<not set>}"
        fi
        [[ -n "$suffix" ]] && printf "%b" "$suffix"
        printf "\n"
    done
    echo ""
}

# ============================================================================
# set / clear — generic, validated by registry
# ============================================================================

set_config() {
    local key="$1" value="$2"
    [[ -z "$key" ]] && hf_die "Usage: hf.config.sh set <key> <value>"
    [[ -z "$value" ]] && hf_die "Usage: hf.config.sh set <key> <value>"
    hf_config_set "$key" "$value"
}

clear_config() {
    local key="$1"
    [[ -z "$key" ]] && hf_die "Usage: hf.config.sh clear <key>"

    if [[ "$key" == "all" ]]; then
        for entry in "${HF_CONFIG_REGISTRY[@]}"; do
            _hf_parse "$entry"
            rm -f "$HF_CONFIG_DIR/$_HF_E_KEY"
        done
        hf_info "All config cleared"
    else
        hf_config_clear "$key"
    fi
}

# ============================================================================
# doctor — scan scripts for hf_require_config and report readiness
# ============================================================================

do_doctor() {
    echo -e "${BOLD}Config Doctor${NC}"
    echo ""

    local total=0 ready=0 missing_count=0

    for script in "$SCRIPT_DIR"/hf.*.sh; do
        [[ "$script" == *hf.lib.sh ]] && continue
        [[ "$script" == *hf.config.sh ]] && continue
        [[ "$script" == *hf.conf.env.sh ]] && continue

        local requires
        requires=$(sed -n 's/^hf_require_config //p' "$script" 2>/dev/null | head -1)
        [[ -z "$requires" ]] && continue

        ((total++))
        local script_name missing_keys
        script_name=$(basename "$script")
        missing_keys=()

        for key in $requires; do
            local value=""
            if [[ -f "$HF_CONFIG_DIR/$key" ]]; then
                value=$(<"$HF_CONFIG_DIR/$key")
            else
                for _e in "${HF_CONFIG_REGISTRY[@]}"; do
                    _hf_parse "$_e"
                    if [[ "$_HF_E_KEY" == "$key" ]]; then
                        value="$_HF_E_DEFAULT"
                        break
                    fi
                done
            fi
            [[ -z "$value" ]] && missing_keys+=("$key")
        done

        if [[ ${#missing_keys[@]} -eq 0 ]]; then
            echo -e "  ${GREEN}●${NC} $script_name"
            ((ready++))
        else
            echo -e "  ${RED}●${NC} $script_name — missing: ${missing_keys[*]}"
            ((missing_count++))
        fi
    done

    echo ""
    echo -e "  ${BOLD}$ready${NC}/$total scripts ready"

    if [[ $missing_count -gt 0 ]]; then
        echo ""
        echo "Fix with: hf.config.sh set <key> <value>"
    fi
}

# ============================================================================
# bootstrap — interactive one-command environment setup
# ============================================================================

do_bootstrap() {
    local env_name="${1:-}"

    echo -e "${BOLD}HyperFleet Environment Bootstrap${NC}"
    echo ""

    # Step 1: Environment name
    if [[ -z "$env_name" ]]; then
        read -rp "Environment name: " env_name
        [[ -z "$env_name" ]] && hf_die "Environment name is required"
    fi

    hf_info "Setting up environment: $env_name"
    echo ""

    # Step 2: Kubernetes context + namespace
    echo -e "${BOLD}Step 1: Kubernetes Context & Namespace${NC}"
    echo ""
    if command -v kubectl &>/dev/null; then
        "$SCRIPT_DIR/hf.kube.context.sh" select
        echo ""
    else
        hf_warn "kubectl not found — skipping context/namespace selection"
        echo ""
    fi

    # Step 3: API URL
    echo -e "${BOLD}Step 2: API Configuration${NC}"
    echo ""
    read -rp "API URL [${HF_API_URL:-http://localhost:8000}]: " api_url
    api_url="${api_url:-${HF_API_URL:-http://localhost:8000}}"
    hf_config_set api-url "$api_url"
    echo ""

    # Step 4: Port forwards
    echo -e "${BOLD}Step 3: Port Forwards${NC}"
    echo ""
    if command -v kubectl &>/dev/null && [[ -n "$(hf_config_value context)" ]]; then
        read -rp "Start port forwards now? [Y/n] " start_pf
        if [[ ! "$start_pf" =~ ^[nN] ]]; then
            "$SCRIPT_DIR/hf.kube.port.forward.sh" start
        else
            hf_info "Skipping port forwards"
        fi
    else
        hf_warn "No kubectl or context — skipping port forwards"
    fi
    echo ""

    # Step 5: Database configuration
    echo -e "${BOLD}Step 4: Database Configuration${NC}"
    echo ""
    read -rp "Configure database connection? [Y/n] " config_db
    if [[ ! "$config_db" =~ ^[nN] ]]; then
        "$SCRIPT_DIR/hf.db.config.sh"
    else
        hf_info "Skipping database configuration"
    fi
    echo ""

    # Step 6: Test API connectivity
    echo -e "${BOLD}Step 5: Testing API Connectivity${NC}"
    echo ""
    hf_info "Testing connection to $(hf_config_value api-url) ..."
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --http1.1 "$(hf_api_base)/clusters" 2>/dev/null)
    if [[ "$http_code" =~ ^2 ]]; then
        hf_info "API is reachable (HTTP $http_code)"
    elif [[ "$http_code" == "000" ]]; then
        hf_warn "API is not reachable — connection refused or timeout"
    else
        hf_warn "API returned HTTP $http_code"
    fi
    echo ""

    # Step 7: Save environment profile
    echo -e "${BOLD}Step 6: Saving Environment Profile${NC}"
    echo ""
    local saved=0
    for entry in "${HF_CONFIG_REGISTRY[@]}"; do
        _hf_parse "$entry"
        # Skip token from environment profiles (security)
        [[ "$_HF_E_KEY" == "token" ]] && continue
        local src_file="$HF_CONFIG_DIR/$_HF_E_KEY"
        if [[ -f "$src_file" ]]; then
            cp "$src_file" "$HF_CONFIG_DIR/$env_name.$_HF_E_KEY"
            ((saved++))
        fi
    done
    hf_info "Saved $saved properties to environment '$env_name'"

    # Step 8: Register and activate
    if ! grep -qx "$env_name" "$ENV_LIST_FILE" 2>/dev/null; then
        echo "$env_name" >>"$ENV_LIST_FILE"
        hf_info "Registered environment '$env_name'"
    fi
    echo "$env_name" >"$ENV_ACTIVE_FILE"
    hf_info "Activated environment '$env_name'"
    echo ""

    # Step 9: Doctor report as summary
    do_doctor
    echo ""
    echo "Manage environments with:"
    echo "  hf.config.sh env list              — list environments"
    echo "  hf.config.sh env show $env_name    — show this environment"
    echo "  hf.config.sh env activate $env_name — re-activate"
}

# ============================================================================
# env list — list environments with property counts
# ============================================================================

do_env_list() {
    if [[ ! -f "$ENV_LIST_FILE" ]] || [[ ! -s "$ENV_LIST_FILE" ]]; then
        hf_info "No environments configured"
        return 0
    fi

    local active
    active=$(_get_active_env)

    echo -e "${BOLD}Environments${NC}"
    echo ""

    while IFS= read -r env; do
        [[ -z "$env" ]] && continue
        local count marker="○" suffix=""
        count=$(_get_env_prop_count "$env")
        if [[ "$env" == "$active" ]]; then
            marker="${GREEN}●${NC}"
            suffix=" (active)"
        fi
        printf "  %b %-30s %s properties%s\n" "$marker" "$env" "$count" "$suffix"
    done <"$ENV_LIST_FILE"
}

# ============================================================================
# env activate — activate an environment
# ============================================================================

do_env_activate() {
    local name="$1"
    [[ -z "$name" ]] && hf_die "Usage: hf.config.sh env activate <name>"
    _env_exists "$name" || hf_die "Environment '$name' not found in $ENV_LIST_FILE"

    # Collect properties that will be overridden
    local -a override_props=()
    for key in "${_CONFIG_KEYS[@]}"; do
        [[ -f "$HF_CONFIG_DIR/$name.$key" ]] && override_props+=("$key")
    done

    if [[ ${#override_props[@]} -eq 0 ]]; then
        hf_warn "Environment '$name' has no properties defined — nothing to activate"
        return 0
    fi

    # Display what will be overridden
    echo -e "${BOLD}Activating environment: $name${NC}"
    echo ""
    echo "The following config properties will be overridden:"
    echo ""

    for key in "${override_props[@]}"; do
        local new_value current_value=""
        new_value=$(<"$HF_CONFIG_DIR/$name.$key")
        [[ -f "$HF_CONFIG_DIR/$key" ]] && current_value=$(<"$HF_CONFIG_DIR/$key")

        # Check sensitivity via registry
        local sensitive=false
        for _e in "${HF_CONFIG_REGISTRY[@]}"; do
            _hf_parse "$_e"
            if [[ "$_HF_E_KEY" == "$key" ]]; then
                [[ "$_HF_E_FLAGS" == *s* ]] && sensitive=true
                break
            fi
        done

        if [[ "$sensitive" == true ]]; then
            printf "  %-35s %s -> %s\n" "$key" "${current_value:+<set>}${current_value:-<not set>}" "<set>"
        else
            printf "  %-35s %s -> %s\n" "$key" "${current_value:-<not set>}" "$new_value"
        fi
    done

    echo ""
    read -rp "Proceed with activation? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[yY] ]]; then
        hf_info "Cancelled"
        return 0
    fi

    echo ""

    # Copy environment properties to active config
    for key in "${override_props[@]}"; do
        cp "$HF_CONFIG_DIR/$name.$key" "$HF_CONFIG_DIR/$key"
        hf_info "Set $key"
    done

    # Record the active environment
    echo "$name" >"$ENV_ACTIVE_FILE"

    echo ""
    hf_info "Environment '$name' is now active"
}

# ============================================================================
# env — dispatch env subcommands
# ============================================================================

do_env() {
    case "${1:-list}" in
        list)     do_env_list ;;
        show)     show_config "${2:-}" ;;
        activate) do_env_activate "${2:-}" ;;
        *)
            hf_usage "env [list|show|activate] [args...]"
            echo "Commands:"
            echo "  list                  List environments with property counts"
            echo "  show [name]           Show config values (with env overrides if specified)"
            echo "  activate <name>       Activate an environment"
            ;;
    esac
}

# ============================================================================
# help
# ============================================================================

show_help() {
    hf_usage "[show|set|clear|doctor|bootstrap|env] [args...]"
    echo "Commands:"
    echo "  show [env-name]         Show current configuration"
    echo "  set <key> <value>       Set a configuration value"
    echo "  clear <key>             Clear a configuration value"
    echo "  clear all               Clear all configuration"
    echo "  doctor                  Check which scripts are ready to use"
    echo "  bootstrap [env-name]    Interactive environment setup"
    echo "  env list                List environments"
    echo "  env show [name]         Show config with environment overrides"
    echo "  env activate <name>     Activate an environment"
    echo ""

    # Show environments
    if [[ -f "$ENV_LIST_FILE" ]] && [[ -s "$ENV_LIST_FILE" ]]; then
        do_env_list
        echo ""
    fi

    # Show config for the active environment
    local active
    active=$(_get_active_env)
    show_config "$active"
}

# ============================================================================
# Main
# ============================================================================

case "${1:-}" in
    "")        show_help ;;
    show)      show_config "${2:-}" ;;
    set)       set_config "${2:-}" "${3:-}" ;;
    clear)     clear_config "${2:-}" ;;
    doctor)    do_doctor ;;
    bootstrap) do_bootstrap "${2:-}" ;;
    env)       do_env "${2:-}" "${3:-}" ;;
    *)         show_help ;;
esac
