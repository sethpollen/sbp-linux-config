package sbpgo

import (
	"bytes"
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Gets a connection to the local memcached.
func LocalMemcache() *memcache.Client {
	return memcache.New("localhost:11211")
}

// Tries to make a version of 'path' which is relative to 'prefix'. If that
// fails, returns 'path' unchanged.
func RelativePath(path string, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return path
	}

	// Remove the prefix.
	path = path[len(prefix):]

	if (path == "") || (path == "/") {
		return "/"
	}

	// Remove a leading slash.
	if strings.HasPrefix(path, "/") {
		return path[1:]
	}

	return path
}

// Runs 'name' in 'pwd' with 'args'.  Will either send one string containing the
// command's stdout to 'outputChan' or send one error to 'errorChan'.
func EvalCommand(outputChan chan<- string, errorChan chan<- error, pwd string,
	name string, args ...string) {
	var cmd = exec.Command(name, args...)
	cmd.Dir = pwd
	text, err := cmd.Output()
	if err != nil {
		errorChan <- err
	} else {
		outputChan <- strings.TrimSpace(string(text))
	}
}

// Synchronous wrapper around EvalCommand.
func EvalCommandSync(pwd string, name string, args ...string) (string, error) {
	var outputChan = make(chan string)
	var errorChan = make(chan error)
	go EvalCommand(outputChan, errorChan, pwd, name, args...)
	select {
	case err := <-errorChan:
		return "", err
	case output := <-outputChan:
		return output, nil
	}
}

// Returns the shortest prefix of 'p' for which 'test' returns true. Returns
// an error if no prefix matched.
func SearchParents(p string, test func(p string) bool) (string, error) {
	// Build a list of prefixes, beginning with the longest.
	var prefixes []string
	for {
		prefixes = append(prefixes, p)
		var oldP = p
		p = path.Dir(p)
		if p == oldP {
			break
		}
	}

	// Search through the list backwards to find the shortest matching prefix.
	for i := len(prefixes) - 1; i >= 0; i-- {
		var prefix = prefixes[i]
		if test(prefix) {
			return prefix, nil
		}
	}

	return "", errors.New("No prefix matched")
}

// Reads all of stdin, blocking until EOF.
func ReadStdin() string {
	var buf bytes.Buffer
	io.Copy(&buf, os.Stdin)
	return buf.String()
}
