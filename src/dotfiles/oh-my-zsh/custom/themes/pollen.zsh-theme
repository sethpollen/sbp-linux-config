# My personal oh-my-zsh theme.

# Overridable function to set up prompt and title bar before each command.
set_up_terminal() {
  # Save the previous command's exit code before polluting it.
  exitcode="$?"
  # Drop one character. Otherwise, Terminator wraps even though it doesn't have
  # to.
  width="$((COLUMNS - 1))"

  # Create a temporary directory for saving sbp-prompt outputs.
  tmpDir="$(mktemp --directory)"
  promptFile="${tmpDir}/prompt"
  titleFile="${tmpDir}/title"

  sbp-prompt --exitcode="$exitcode" --width="$width" \
             --prompt_file="$promptFile" --title_file="$titleFile"

  export PROMPT="$(cat $promptFile)"
  title="$(cat $titleFile)"
  export ZSH_THEME_TERM_TAB_TITLE_IDLE="$title"
  export ZSH_THEME_TERM_TITLE_ITLE="$title"

#  rm -rf "$tmpDir"
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
