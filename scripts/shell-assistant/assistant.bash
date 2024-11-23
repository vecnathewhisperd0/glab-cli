_glab_duo_bash() {
    if [[ -n "$READLINE_LINE" ]]; then
        echo -en "\rGenerating command...\r"
        
        local command=$(glab duo ask --shell "$READLINE_LINE" | head -n1 | tr -d '\r\n')
            
        echo -en "\r                      \r"
            
        if [[ -n "$command" ]]; then
            READLINE_LINE="$command"
            READLINE_POINT=${#READLINE_LINE}
        fi
    fi
}

bind -x '"\ee": _glab_duo_bash'
