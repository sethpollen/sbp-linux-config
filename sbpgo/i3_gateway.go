// Library for manipulating a running i3 desktop via i3-msg.

package sbpgo

import (
  "encoding/json"
  "fmt"
  "os"
  "os/exec"
  "sort"
  "strconv"
  "strings"
)

type Workspace struct {
  Num int
  Name string
  Focused bool
}

// Gets the current list of workspaces.
func getWorkspaces() ([]Workspace, error) {
  cmd := exec.Command("i3-msg", "-t", "get_workspaces")
  result, err := cmd.Output()
  if err != nil {
    return nil, err
  }

  var rawJson interface{}
  err = json.Unmarshal(result, &rawJson)
  if err != nil {
    return nil, err
  }

  jsonList, ok := rawJson.([]interface{})
  if !ok {
    return nil, fmt.Errorf("get_workspaces result is not a list")
  }

  var workspaces []Workspace
  for _, w := range jsonList {
    jsonMap, ok := w.(map[string]interface{})
    if !ok {
      return nil, fmt.Errorf("get_workspaces result is not a list of maps")
    }

    rawNum, ok := jsonMap["num"]
    if !ok {
      return nil, fmt.Errorf("get_workspaces did not return num")
    }
    num, ok := rawNum.(float64)
    if !ok {
      return nil, fmt.Errorf("get_workspaces num is not a number")
    }

    rawName, ok := jsonMap["name"]
    if !ok {
      return nil, fmt.Errorf("get_workspaces did not return name")
    }
    name, ok := rawName.(string)
    if !ok {
      return nil, fmt.Errorf("get_workspaces name is not a string")
    }

    rawFocused, ok := jsonMap["focused"]
    if !ok {
      return nil, fmt.Errorf("get_workspace did not return focused")
    }
    focused, ok := rawFocused.(bool)
    if !ok {
      return nil, fmt.Errorf("get_workspace focused is not a bool")
    }

    workspaces = append(workspaces, Workspace{int(num), name, focused})
  }

  return workspaces, nil
}

// Issues a command via i3-msg.
func issueI3Commands(cmds ...string) error {
  return exec.Command("i3-msg", "--quiet", strings.Join(cmds, " ; ")).Run()
}

func getCurrentWorkspace(ws []Workspace) (*Workspace, error) {
  for _, w := range ws {
    if w.Focused {
      return &w, nil
    }
  }
  return nil, fmt.Errorf("No workspace currently focused")
}

// Gets the smallest unused workspace number.
func nextFreeWorkspaceNumber(ws []Workspace) int {
  var usedNums []int
  for _, w := range ws {
    usedNums = append(usedNums, w.Num)
  }
  sort.Ints(usedNums)

  // Find the first unused positive number.
  num := 1
  for _, usedNum := range usedNums {
    if num != usedNum {
      return num
    }
    num++
  }
  return num
}

// Parses the workspace number out of 'name'. Returns -1 if there doesn't appear
// to be a workspace number present.
func parseWorkspaceNumber(name string) int {
  firstPart := strings.SplitN(name, ":", 2)[0]
  num, err := strconv.Atoi(firstPart)
  if err != nil {
    return -1
  }
  // Zero is not a valid workspace number.
  if num <= 0 {
    return -1
  }
  return num
}

// Removes the leading workspace number and colon. Returns an empty string if
// the name is just a number (no colon).
func removeWorkspaceNumber(name string) string {
  parts := strings.SplitN(name, ":", 2)

  _, err := strconv.Atoi(parts[0])
  if err != nil {
    // The first part isn't a number, so don't remove anything.
    return name
  }

  // Return the second part, or an empty string if there is no second part.
  if len(parts) == 1 {
    return ""
  }
  return parts[1]
}

func makeWorkspaceName(num int, rest string) string {
  if len(rest) == 0 {
    return fmt.Sprintf("%d", num)
  }
  return fmt.Sprintf("%d:%s", num, rest)
}

func RenameCurrentWorkspace() error {
  selection, err := Dmenu("New workspace name:", nil)
  if err != nil {
    return err
  }
  if len(selection) == 0 {
    // User aborted.
    return nil
  }

  ws, err := getWorkspaces()
  if err != nil {
    return err
  }

  current, err := getCurrentWorkspace(ws)
  if err != nil {
    return err
  }

  if parseWorkspaceNumber(selection) <= 0 {
    // The user didn't specify a number. Just keep the workspace's existing
    // number.
    num := parseWorkspaceNumber(current.Name)
    if num <= 0 {
      // Something weird is happening. Just use a default.
      num = 1
    }
    selection = fmt.Sprintf("%d:%s", num, selection)
  }

  return issueI3Commands(fmt.Sprintf("rename workspace \"%s\" to \"%s\"",
                                     current.Name, selection))
}

func SwitchToNewWorkspace() error {
  ws, err := getWorkspaces()
  if err != nil {
    return err
  }

  num := nextFreeWorkspaceNumber(ws)
  return issueI3Commands(fmt.Sprintf("workspace number %d", num))
}

// Entry point.
func I3GatewayMain() {
  if len(os.Args) < 2 {
    fmt.Fprintln(os.Stderr, "No subcommand")
    os.Exit(1)
  }
  var subcommand = os.Args[1]

  var err error
  switch subcommand {

  case "rename":
    err = RenameCurrentWorkspace()

  case "switch_new":
    err = SwitchToNewWorkspace()

  default:
    fmt.Fprintln(os.Stderr, "Unrecognized subcommand:", subcommand)
    os.Exit(1)
  }

  if err != nil {
    fmt.Println(err)
    os.Exit(2)
  }
}
