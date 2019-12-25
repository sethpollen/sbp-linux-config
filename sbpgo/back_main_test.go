// Covers both back_main.go and future.go by invoking back_main as a child
// process.

package sbpgo_test

import (
  "bytes"
  "os/exec"
  "strings"
  "testing"
  "time"
)

const backMain = "./linux_amd64_stripped/back_main_for_test"

func call(t *testing.T,
          argv []string,
          expectSuccess bool,
          expectedStdout string,
          expectedStderr string) {
  c := exec.Command(backMain, argv...)

  argvStr := strings.Join(argv, " ")

  var stdout bytes.Buffer
  var stderr bytes.Buffer
  c.Stdout = &stdout
  c.Stderr = &stderr

  err := c.Run()
  if err != nil {
    _, ok := err.(*exec.ExitError)
    if !ok {
      t.Errorf("[%s] Unexpected exec failure: %v", argvStr, err)
      return
    }
    if expectSuccess {
      t.Errorf("[%s] Unexpected command failure: %v", argvStr, err)
    }
  } else if !expectSuccess {
    t.Errorf("[%s] Unexpected command success", argvStr)
  }

  if bytes.Compare(stdout.Bytes(), []byte(expectedStdout)) != 0 {
    t.Errorf("[%s] Unexpected stdout:\n%s", argvStr, stdout.Bytes())
  }
  if bytes.Compare(stderr.Bytes(), []byte(expectedStderr)) != 0 {
    t.Errorf("[%s] Unexpected stderr:\n%s", argvStr, stderr.Bytes())
  }
}

func TestHelp(t *testing.T) {
  call(t, []string{}, false,
       "", "No subcommand. Try one of these:\n" +
           "  ls ls_nostar start peek poll reclaim kill\n")
}

func TestBasicWorkflow(t *testing.T) {
  call(t, []string{"ls"}, true, "", "")
  call(t, []string{"start", "job", "echo foo"}, true, "", "")
  time.Sleep(100 * time.Millisecond)

  // We can see that the job has completed.
  call(t, []string{"ls"}, true, "job *\n", "")
  call(t, []string{"peek", "job"}, true, "foo\n", "")
  call(t, []string{"poll", "job"}, true, "", "")

  // Clean up.
  call(t, []string{"reclaim", "job"}, true, "", "")

  // The job is now gon.e
  call(t, []string{"peek", "job"}, false, "", "Job does not exist: job\n")
}

func TestJobPassedAsMultiplePieces(t *testing.T) {
  call(t, []string{"start", "job", "echo", "foo;", "and echo bar"}, true,
       "", "")
  time.Sleep(100 * time.Millisecond)

  // The job pieces should have been stitched together and then evaluated in
  // a shell.
  call(t, []string{"peek", "job"}, true, "foo\nbar\n", "")

  // Clean up.
  call(t, []string{"reclaim", "job"}, true, "", "")
}

func TestKill(t *testing.T) {
  call(t, []string{"start", "job", "echo foo; and sleep 100000"}, true, "", "")
  time.Sleep(100 * time.Millisecond)

  // The job produced some output but it still running.
  call(t, []string{"peek", "job"}, true, "foo\n", "")
  call(t, []string{"poll", "job"}, false, "", "Job still running: job\n")
  call(t, []string{"reclaim", "job"}, false, "", "Job still running: job\n")
  call(t, []string{"ls"}, true, "job\n", "")

  // Kill it.
  call(t, []string{"kill", "job"}, true, "", "")

  // We can still see the output it produced before it died.
  call(t, []string{"peek", "job"}, true, "foo\n", "")

  // Clean up.
  call(t, []string{"reclaim", "job"}, true, "", "")
}

func TestKillCompletedJob(t *testing.T) {
  call(t, []string{"start", "job", "echo foo"}, true, "", "")
  time.Sleep(100 * time.Millisecond)

  // The job has completed
  call(t, []string{"poll", "job"}, true, "", "")

  // Kill it anyway. This is a successful no-op.
  call(t, []string{"kill", "job"}, true, "", "")

  // Clean up.
  call(t, []string{"reclaim", "job"}, true, "", "")
}

func TestJobNotFound(t *testing.T) {
  call(t, []string{"peek", "job"}, false, "", "Job does not exist: job\n")
  call(t, []string{"kill", "job"}, false, "", "Job does not exist: job\n")
  call(t, []string{"poll", "job"}, false, "", "Job does not exist: job\n")
  call(t, []string{"reclaim", "job"}, false, "", "Job does not exist: job\n")
}

func TestEmptyJob(t *testing.T) {
  // It's OK to pass nothing as the job program.
  call(t, []string{"start", "job"}, true, "", "")
  time.Sleep(100 * time.Millisecond)

  // Evaluating an empty string in a shell produces no output.
  call(t, []string{"peek", "job"}, true, "", "")

  // Cleanup.
  call(t, []string{"reclaim", "job"}, true, "", "")
}

func TestLs(t *testing.T) {
  call(t, []string{"start", "a", "sleep 100000"}, true, "", "")
  call(t, []string{"start", "b"}, true, "", "")
  call(t, []string{"start", "c", "sleep 100000"}, true, "", "")
  call(t, []string{"start", "d"}, true, "", "")
  time.Sleep(100 * time.Millisecond)

  call(t, []string{"ls"}, true, "b *\nd *\na\nc\n", "")
  call(t, []string{"ls_nostar"}, true, "b\nd\na\nc\n", "")

  // Clean up.
  call(t, []string{"kill", "a"}, true, "", "")
  call(t, []string{"kill", "c"}, true, "", "")

  call(t, []string{"reclaim", "a"}, true, "", "")
  call(t, []string{"reclaim", "b"}, true, "", "")
  call(t, []string{"reclaim", "c"}, true, "", "")
  call(t, []string{"reclaim", "d"}, true, "", "")
}

func TestMissingArgs(t *testing.T) {
  call(t, []string{"start"}, false, "", "No job specified\n")
  call(t, []string{"peek"}, false, "", "No job specified\n")
  call(t, []string{"poll"}, false, "", "No job specified\n")
  call(t, []string{"reclaim"}, false, "", "No job specified\n")
  call(t, []string{"kill"}, false, "", "No job specified\n")
}

func TestTooManyArgs(t *testing.T) {
  call(t, []string{"ls", "foo"}, false, "", "Too many args: foo\n")
  call(t, []string{"peek", "job", "foo"}, false, "", "Too many args: foo\n")
  call(t, []string{"poll", "job", "foo"}, false, "", "Too many args: foo\n")
  call(t, []string{"reclaim", "job", "foo"}, false, "", "Too many args: foo\n")
  call(t, []string{"kill", "job", "foo"}, false, "", "Too many args: foo\n")
}

func TestBogusSubcommand(t *testing.T) {
  // These are not valud subcommands.
  call(t, []string{"fork", "foo"}, false, "", "Unrecognized subcommand: fork\n")
  call(t, []string{"join", "foo"}, false, "", "Unrecognized subcommand: join\n")
}
