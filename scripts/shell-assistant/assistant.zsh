function _glab_duo_zsh() {
    # Show status message
    echo -en "\rGenerating command...\r"
    
    # Get command suggestion using shell command
    local command=$(glab duo ask --shell "$BUFFER" 2>/dev/null | head -n1 | tr -d '\r\n')
        
    # Clear the status message with spaces and reset cursor
    echo -en "\r                      \r"
        
    if [[ -n "$command" ]]; then
        BUFFER="$command"
        CURSOR=${#BUFFER}
    fi
}

zle -N _glab_duo_zsh
bindkey '\ee' _glab_duo_zsh
