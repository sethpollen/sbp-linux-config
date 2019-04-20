// Library for constructing prompt strings of the specific form that I like.
package sbpgo

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"
	"unicode/utf8"
)

// Collects information during construction of a prompt string.
type PromptEnv struct {
	// If nil, the current date/time will be omitted from the prompt string.
	Now            *time.Time
	Home           string
	Pwd            string
	Hostname       string
	ShortHostname  string
	RunningOverSsh bool
	TmuxStatus     *TmuxStatus
	// Text to include in the prompt, along with the PWD.
	Info string
	// A short string to place before the final $ in the prompt.
	Flag StyledString
	// Exit code of the last process run in the shell.
	ExitCode int
	// Maximum number of characters which prompt may occupy horizontally.
	Width int
	// Environment variables which should be emitted to the shell which uses this
	// prompt.
	EnvironMod EnvironMod
}

func GetPwd() string {
	// If possible, get the pwd from $PWD, as this usually does the right thing
	// with symlinks (i.e. it shows the path you used to get here, not the
	// actual physical path). If $PWD fails, fall back on os.Getwd().
	pwd := os.Getenv("PWD")
	if len(pwd) == 0 {
		pwd, _ = os.Getwd()
	}
	return pwd
}

// Generates a PromptEnv based on current environment variables. The maximum
// number of characters which the prompt may occupy must be passed as 'width'.
// If 'now' is nil, the current date/time will be omitted from the prompt
// string.
func NewPromptEnv(
	pwd string,
	width int,
	exitCode int,
	now *time.Time) *PromptEnv {

	var self = new(PromptEnv)
	self.Now = now

	user, err := user.Current()
	if err != nil {
		self.Home = ""
	} else {
		self.Home = user.HomeDir
	}

	self.Pwd = pwd
	self.Hostname, _ = os.Hostname()
	self.ShortHostname = strings.SplitN(self.Hostname, ".", 2)[0]
	self.RunningOverSsh = (os.Getenv("SSH_TTY") != "")
	self.TmuxStatus = GetTmuxStatus()

	self.Info = ""
	self.ExitCode = exitCode
	self.Width = width
	self.EnvironMod = *NewEnvironMod()

	return self
}

type Prompt struct {
	Prompt StyledString
	Title  string
}

// Generates a shell prompt string.
func (self *PromptEnv) makePrompt(
	pwdMod func(in StyledString) StyledString) Prompt {
	// Construct the prompt text which must precede the PWD.
	var promptBeforePwd StyledString
	var title string

	promptBeforePwd = Stylize("╭╴", Cyan, nil, true)

	// Date and time, if supplied.
	if self.Now != nil {
		promptBeforePwd =
			append(promptBeforePwd,
				Stylize(self.Now.Format("01/02 15:04"), White, Blue, true)...)
	}
  
  // TODO: just testing this separator style.
  promptBeforePwd = append(promptBeforePwd, Stylize("▌", Blue, Magenta, false)...)

	// Hostname.
	if self.RunningOverSsh {
		promptBeforePwd = append(promptBeforePwd, Stylize("(", Yellow, nil, false)...)
		title += "("
	}

	promptBeforePwd = append(promptBeforePwd,
		Stylize(self.ShortHostname, White, Magenta, true)...)
	title += self.ShortHostname

	tmuxSession := self.TmuxStatus.AttachedSession()
	if tmuxSession != "" {
		promptBeforePwd = append(promptBeforePwd, Stylize("|", Yellow, nil, true)...)
		promptBeforePwd = append(promptBeforePwd, Stylize(tmuxSession, Magenta, nil, true)...)
		title += "|" + tmuxSession
	}

	if self.RunningOverSsh {
		promptBeforePwd = append(promptBeforePwd, Stylize(")", Yellow, nil, false)...)
		title += ")"
	}

	tmuxSessions := self.TmuxStatus.Sessions()
	if len(tmuxSessions) > 0 {
		attention := false
		for _, a := range tmuxSessions {
			if a {
				attention = true
				break
			}
		}
		if attention {
			// Show a bold ! to indicate "bell".
			promptBeforePwd = append(promptBeforePwd, Stylize("!", Yellow, nil, true)...)
		} else {
			// Show a subtle % to indicate "running".
			promptBeforePwd = append(promptBeforePwd, Stylize("%%", Yellow, nil, false)...)
		}
	}

	// Info (if we got one).
	if self.Info != "" {
		promptBeforePwd = append(promptBeforePwd, Stylize(" [", White, nil, false)...)
		promptBeforePwd = append(promptBeforePwd,
			Stylize(self.Info, White, nil, true)...)
		promptBeforePwd = append(promptBeforePwd, Stylize("]", White, nil, false)...)
		title += " [" + self.Info + "]"
	}

	// Construct the prompt text which must follow the PWD.
	var promptAfterPwd StyledString

	// Exit code.
	if self.ExitCode != 0 {
		promptAfterPwd = Stylize(fmt.Sprintf("[%d]", self.ExitCode), Red, nil, true)
	}

	// Determine how much space is left for the PWD.
	var pwdWidth = self.Width - len(promptBeforePwd) - len(promptAfterPwd)
	if pwdWidth < 0 {
		pwdWidth = 0
	}
	var pwdOnItsOwnLine = false
	if pwdWidth < 20 &&
		utf8.RuneCountInString(self.Pwd) > pwdWidth &&
		self.Width >= pwdWidth {
		// Don't cram the PWD into a tiny space; put it on its own line.
		pwdWidth = self.Width
		pwdOnItsOwnLine = true
	}

	var pwdPrompt = self.formatPwd(pwdMod, pwdWidth)
	title += " " + pwdPrompt.PlainString()

	// Build the complete prompt string.
	var fullPrompt StyledString = promptBeforePwd
	if pwdOnItsOwnLine {
		fullPrompt = append(fullPrompt, Unstyled(" ")...)
		fullPrompt = append(fullPrompt, promptAfterPwd...)
		fullPrompt = append(fullPrompt, Stylize("\n│ ", Cyan, nil, true)...)
		fullPrompt = append(fullPrompt, pwdPrompt...)
	} else {
		fullPrompt = append(fullPrompt, Unstyled(" ")...)
		fullPrompt = append(fullPrompt, pwdPrompt...)
		fullPrompt = append(fullPrompt, promptAfterPwd...)
	}
	fullPrompt = append(fullPrompt, Stylize("\n╰╴", Cyan, nil, true)...)
	fullPrompt = append(fullPrompt, self.Flag...)

	switch ShellTypeFlag() {
	case "posix":
		fullPrompt = append(fullPrompt, Stylize("$ ", Yellow, nil, true)...)
	case "fish":
		fullPrompt = append(fullPrompt, Stylize("~> ", Yellow, nil, true)...)
	}

	return Prompt{fullPrompt, title}
}

// Formats the PWD for use in a prompt. 'mod' is an arbitrary transformation
// to apply to the full PWD before it is (potentially) truncated. The return
// value always ends in a space character unless it is empty.
func (self *PromptEnv) formatPwd(
	mod func(in StyledString) StyledString, width int) StyledString {
	// Perform tilde collapsing on the PWD.
	var home = self.Home
	if strings.HasSuffix(home, "/") {
		home = home[:len(home)-1]
	}
	var pwd = self.Pwd
	if strings.HasPrefix(pwd, home) {
		pwd = "~" + pwd[len(home):]
	}
	if pwd == "" {
		pwd = "/"
	}

	var styledPwd StyledString = Stylize(pwd, Cyan, nil, true)

	if mod != nil {
		styledPwd = mod(styledPwd)
	}

	// TODO: Dim slashes in the PWD by darkening the FG color. Do the same for
	// the … character below, and maybe for the SSH parentheses and tmux flags.

	var pwdRunes = utf8.RuneCountInString(styledPwd.PlainString())
	// Subtract 1 in case we have to include the ellipsis character.
	// Subtract another 1 for the space character. Subtract another 1
	// for I-don't-know-what reason. We just have to, or the terminal
	// inserts a blank line after the PWD.
	var start = pwdRunes - (width - 3)
	if start > 0 {
		// Truncate the PWD.
		if start >= pwdRunes {
			// There is no room for the PWD at all.
			styledPwd = make(StyledString, 0)
		} else {
			styledPwd = styledPwd[start:]
			var withEllipsis StyledString = Stylize("…", Cyan, nil, false)
			withEllipsis = append(withEllipsis, styledPwd...)
			styledPwd = withEllipsis
		}
	}

	if len(styledPwd) > 0 {
		styledPwd = append(styledPwd, Unstyled(" ")...)
	}

	return styledPwd
}

// Renders all the information from this PromptEnv into a shell script which
// may be sourced. The following variables will be set:
//   PROMPT
//   TERM_TITLE
//   ... plus any other variables set in self.EnvironMod.
func (self *PromptEnv) ToScript(
	pwdMod func(in StyledString) StyledString) string {
	// Start by making a copy of the custom EnvironMod.
	var mod = self.EnvironMod
	// Now add our variables to it.
	var prompt = self.makePrompt(pwdMod)
	mod.SetVar("PROMPT", prompt.Prompt.AnsiString(ShellTypeFlag() == "posix"))
	mod.SetVar("TERM_TITLE", prompt.Title)
	// Include the Info string separately, since it is sometimes useful
	// on its own (i.e. as the name of the current repo).
	mod.SetVar("INFO", self.Info)
	return mod.ToScript()
}
