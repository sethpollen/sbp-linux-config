# My personal oh-my-zsh theme.

# Sets up prompt and title bar before each command.
set_up_terminal() {
  # Source the environment variables emitted by sbp-prompt.
  . <(sbp-prompt --exitcode="$?" --width="$COLUMNS")

  export ZSH_THEME_TERM_TAB_TITLE_IDLE="$TERM_TITLE"
  export ZSH_THEME_TERM_TITLE_IDLE="$TERM_TITLE"
}

# Print a bell character. If using the terminator terminal emulator, this should
# cause set the X window's urgency bit.
print_bell() {
  print -n "\a"
}

# Register hooks.
autoload -U add-zsh-hook
add-zsh-hook precmd print_bell

# Manually insert set_up_terminal before all other precmd hooks.
add_to_precmd_start() {
  precmd_functions=($* $precmd_functions)
}
add_to_precmd_start set_up_terminal
