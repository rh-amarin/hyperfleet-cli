#!/bin/bash
# Select and save Kubernetes context and namespace
source "$(dirname "$(realpath "$0")")/hf.lib.sh"

hf_require_kubectl

show_status() {
    echo -e "${BOLD}Current Settings${NC}"
    echo -e "  Context:   ${CYAN}${HF_KUBE_CONTEXT:-<not set>}${NC}"
    echo -e "  Namespace: ${CYAN}${HF_KUBE_NAMESPACE:-<not set>}${NC}"
    echo ""
}

select_context() {
    echo -e "${BOLD}Available Contexts${NC}"
    local contexts
    contexts=$(kubectl config get-contexts -o name)

    local i=1
    local ctx_array=()
    while IFS= read -r ctx; do
        [[ -z "$ctx" ]] && continue
        ctx_array+=("$ctx")
        if [[ "$ctx" == "$HF_KUBE_CONTEXT" ]]; then
            echo -e "  ${GREEN}$i)${NC} $ctx ${GREEN}(current)${NC}"
        else
            echo -e "  $i) $ctx"
        fi
        ((i++))
    done <<< "$contexts"

    echo ""
    read -rp "Select context [1-$((i-1))]: " choice

    if [[ "$choice" =~ ^[0-9]+$ ]] && (( choice >= 1 && choice < i )); then
        local selected="${ctx_array[$((choice-1))]}"
        hf_set_context "$selected"
    else
        hf_warn "Invalid selection"
        return 1
    fi
}

select_namespace() {
    [[ -z "$HF_KUBE_CONTEXT" ]] && hf_die "No context set. Select a context first."

    echo -e "${BOLD}Available Namespaces${NC} (in $HF_KUBE_CONTEXT)"
    local namespaces
    namespaces=$(hf_kubectl get namespaces -o jsonpath='{.items[*].metadata.name}')

    local i=1
    local ns_array=()
    for ns in $namespaces; do
        ns_array+=("$ns")
        if [[ "$ns" == "$HF_KUBE_NAMESPACE" ]]; then
            echo -e "  ${GREEN}$i)${NC} $ns ${GREEN}(current)${NC}"
        else
            echo -e "  $i) $ns"
        fi
        ((i++))
    done

    echo ""
    read -rp "Select namespace [1-$((i-1))]: " choice

    if [[ "$choice" =~ ^[0-9]+$ ]] && (( choice >= 1 && choice < i )); then
        local selected="${ns_array[$((choice-1))]}"
        hf_set_namespace "$selected"
    else
        hf_warn "Invalid selection"
        return 1
    fi
}

do_set() {
    local context="${1:-}"
    local namespace="${2:-}"

    [[ -z "$context" ]] && hf_die "Usage: hf.kube.context.sh set <context> [namespace]"

    # Verify context exists
    if ! kubectl config get-contexts "$context" &>/dev/null; then
        hf_die "Context '$context' not found"
    fi
    hf_set_context "$context"

    if [[ -n "$namespace" ]]; then
        # Verify namespace exists
        if ! hf_kubectl get namespace "$namespace" &>/dev/null; then
            hf_die "Namespace '$namespace' not found in context '$context'"
        fi
        hf_set_namespace "$namespace"
    fi
}

do_clear() {
    hf_clear_context
    hf_clear_namespace
}

case "${1:-}" in
    status|"")
        show_status
        ;;
    context)
        select_context
        ;;
    namespace|ns)
        select_namespace
        ;;
    select)
        select_context && echo "" && select_namespace
        ;;
    set)
        do_set "${2:-}" "${3:-}"
        ;;
    clear)
        do_clear
        ;;
    *)
        hf_usage "[status|context|namespace|select|set|clear]"
        echo "Commands:"
        echo "  status     Show current context and namespace (default)"
        echo "  context    Interactively select a context"
        echo "  namespace  Interactively select a namespace"
        echo "  select     Select both context and namespace"
        echo "  set <ctx> [ns]  Set context and optionally namespace"
        echo "  clear      Clear saved context and namespace"
        ;;
esac
