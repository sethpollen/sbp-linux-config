# My personal oh-my-zsh theme.

clear_cwd_file() {
  # When the shell exits, clear the remembered cwd.
  rm -f ~/.cwd
}

# Prompt colors.
cyan="%{$fg_bold[cyan]%}"
yellow="%{$fg_bold[yellow]%}"
red="%{$fg_bold[red]%}"
white="%{$fg_bold[white]%}"
no_color="%{$reset_color%}"

# Standard function for building prompt strings. The result is exported to the
# PROMPT variable.
# Arguments:
#   --maxlen=INTEGER
#       Specifies the maximum number of characters for the resulting string.
#       The PWD may be truncated to make it fit, but no other truncation
#       efforts will occur. Defaults to $COLUMNS.
#   --info=STRING
#       Optionally provides an info string to print before the PWD. If this is
#       non-empty, it will be surrounded by square brackets and printed out
#       in white.
#   --pwd=STRING
#       Optionally specifies the string to print as the PWD. If omitted, the
#       complete current PWD is used.
build_prompt() {
  # Parse args.
  local info=
  local pwd="%~"
  local maxlen=
  for arg in "$@"; do
    case "$arg" in
      --maxlen=*)
        maxlen="${arg#--maxlen=}" ;;
      --info=*)
        info="${arg#--info=}" ;;
      --pwd=*)
        pwd="${arg#--pwd=}" ;;
    esac
  done

  # Check if we need a default maxlen.
  if [ -z "$maxlen" ]; then
    if [ $COLUMNS ]; then
      maxlen=$COLUMNS
    else
      maxlen=9999
    fi
  fi

  # Make sure we know our hostname.
  if [ -z "$HOST" ]; then
    export HOST=$(hostname)
  fi

  # Dress up the info, if we got one.
  if [ ! -z "$info" ]; then
    # Add a trailing space to set it off from the PWD.
    info="[${info}] "
  fi

  # Compute how much space we have for the PWD. We take off 12 for the date
  # and time, then the number of characters in the hostname, 1 for the space
  # after the hostname, then the number of characters in the info, then four
  # more for the exit status.
  local pwd_maxlen=$((maxlen - 12 - $#HOST - 1 - $#info - 4))
  if [[ $pwd_maxlen -lt 2 ]]; then
    # We need at least 2 spots for the "..".
    pwd_maxlen=2
  fi

  # Build up the prompt.
  PROMPT="${cyan}%D{%m/%d %H:%M} %m "
  if [ ! -z "$info" ]; then
    PROMPT="${PROMPT}${white}${info}${cyan}"
  fi
  PROMPT="${PROMPT}%${pwd_maxlen}<..<${pwd}%<<%(?.. ${red}[%?])
${yellow}>${no_color} "

  # Make sure the PROMPT variable is exported to the outer ZSH environment.
  export PROMPT
}

# Standard function for building titlebar strings. The result is exported to the
# appropriate oh-my-zsh variables.
# Arguments:
#   --maxlen=INTEGER
#       Specifies the maximum number of characters for the resulting string.
#       The PWD may be truncated to make it fit, but no other truncation
#       efforts will occur.
#   --info=STRING
#       Optionally provides an info string to print before the PWD. If this is
#       non-empty, it will be surrounded by square brackets.
#   --pwd=STRING
#       Optionally specifies the string to print as the PWD. If omitted, the
#       complete current PWD is used.
build_title_bar() {
  # Parse args.
  local info=
  local pwd="%~"
  local maxlen=
  for arg in "$@"; do
    case "$arg" in
      --maxlen=*)
        maxlen="${arg#--maxlen=}" ;;
      --info=*)
        info="${arg#--info=}" ;;
      --pwd=*)
        pwd="${arg#--pwd=}" ;;
    esac
  done

  # Check if we need a default maxlen.
  if [ -z "$maxlen" ]; then
    maxlen=9999
  fi

  # Make sure we know our hostname.
  if [ -z "$HOST" ]; then
    export HOST=$(hostname)
  fi

  # Dress up the info, if we got one.
  if [ ! -z "$info" ]; then
    # Add a trailing space to set it off from the PWD.
    info="[${info}] "
  fi

  # Compute how much space we have for the PWD. We take off the number of
  # characters in the hostname, 1 for the space after the hostname, then
  # the number of characters in the info.
  local pwd_maxlen=$((maxlen - $#HOST - 1 - $#info))
  if [[ $pwd_maxlen -lt 2 ]]; then
    # We need at least 2 spots for the "..".
    pwd_maxlen=2
  fi

  # Build up the title bar string.
  local title_bar="%m "
  if [ ! -z "$info" ]; then
    title_bar="${title_bar}${info}"
  fi
  title_bar="${title_bar}%${pwd_maxlen}<..<${pwd}%<<"

  # Export to oh-my-zsh.
  export ZSH_THEME_TERM_TAB_TITLE_IDLE=$title_bar
  export ZSH_THEME_TERM_TITLE_IDLE=$title_bar
}

# Overridable function to set up prompt and title bar before each command.
set_up_terminal() {
  # For the prompt, let the default length be $COLUMNS.
  build_prompt

  # For the title bars, cap the length at 70 characters.
  build_title_bar --maxlen=70
}

# Register hooks.
autoload -U add-zsh-hook
add-zsh-hook zshexit clear_cwd_file
add-zsh-hook precmd set_up_terminal
