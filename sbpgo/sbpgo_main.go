package main

import (
  "github.com/sethpollen/sbp_linux_config/sbpgo"
  "fmt"
  "log"
  "os"
  "os/user"
  "path"
)

// Single entry point for all of my Go programs. This makes it easier to install
// the suite. Each program is pretty small, so the overall binary size
// continues to be dominated by the 2 MiB of Go runtime.

func backHome() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalln(err)
    return ""
	}
	return path.Join(user.HomeDir, ".back")
}

func main() {
	if len(os.Args) < 2 {
    fmt.Fprintln(os.Stderr, "No subcommand")
    os.Exit(1)
  }
  var subcommand = os.Args[1]
  os.Args = os.Args[1:]

  switch subcommand {

  case "back":
    sbpgo.BackMain(backHome(), true)

  case "format_percent":
    sbpgo.FormatPercentMain()

  case "network_usage":
    sbpgo.NetworkUsageMain()

  default:
    fmt.Fprintln(os.Stderr, "Unrecognized subcommand:", subcommand)
    os.Exit(1)
  }
}
