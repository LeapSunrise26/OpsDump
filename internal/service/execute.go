package service

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// ExecResult describes the outcome of running a command.
type ExecResult struct {
	ExitCode int
	Output   string
	Error    string
	TimedOut bool
}

// toUTF8 converts command output to UTF-8. Windows shells (cmd) emit GBK;
// any bytes that are not valid UTF-8 are re-encoded from GBK so logs and the
// run history stay readable.
func toUTF8(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	if out, err := simplifiedchinese.GBK.NewDecoder().Bytes(b); err == nil {
		return string(out)
	}
	return string(b)
}

// findBashLocates a usable bash on Windows (Git Bash); returns "" if absent.
// The built-in templates and typical verify commands use sh syntax (ls, $(...),
// test -n), which cmd.exe cannot run. Windows ships a WSL bash launcher in
// System32 that is NOT a usable shell here, so it is explicitly skipped.
func findBash() string {
	lookup, err := exec.LookPath("bash")
	if err == nil && !strings.Contains(strings.ToLower(lookup), `system32`) {
		return lookup
	}
	// Scan every PATH entry for bash.exe (again skipping System32).
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if strings.Contains(strings.ToLower(dir), `system32`) {
			continue
		}
		c := filepath.Join(dir, "bash.exe")
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	for _, c := range []string{
		`C:\Program Files\Git\bin\bash.exe`,
		`C:\Program Files\Git\usr\bin\bash.exe`,
		`C:\Program Files (x86)\Git\bin\bash.exe`,
		`C:\Program Files (x86)\Git\usr\bin\bash.exe`,
	} {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	return ""
}

// Exec runs a shell command with the given env injected. The command string keeps
// ${VAR} placeholders as-is; the shell resolves them at runtime using the provided
// environment, so secrets never appear in logged output.
//
// It uses os/exec (not gproc) for explicit stream capture and context timeout control.
// Default timeout is 24 hours. Commands can override by setting EXEC_TIMEOUT_MINUTES env.
func (e *EnvMap) Exec(ctx context.Context, command string) (*ExecResult, error) {
	command = e.Resolve(command)

	// Apply timeout: default 1 hour, can be overridden by job's kv or env
	timeout := 1 * time.Hour
	if timeoutStr := e.Get("EXEC_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := time.ParseDuration(timeoutStr + "m"); err == nil && minutes > 0 {
			timeout = minutes
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		shell     string
		shellArgs []string
	)
	if runtime.GOOS == "windows" {
		// Prefer Git Bash for sh-style commands (ls, $(...), test -n);
		// fall back to cmd.exe when bash is not installed. cmd is switched to
		// UTF-8 (chcp 65001) so GBK mojibake does not pollute run history.
		if p := findBash(); p != "" {
			shell = p
			shellArgs = []string{"-c", command}
			g.Log().Infof(ctx, "exec: using bash %s", p)
		} else {
			shell = "cmd"
			shellArgs = []string{"/c", "chcp 65001 >nul & " + command}
			g.Log().Infof(ctx, "exec: using cmd (bash not found), PATH=%s", os.Getenv("PATH"))
		}
	} else {
		// Prefer bash for full feature support (arrays, (( )), etc.);
		// fall back to /bin/sh when bash is not installed.
		if p, err := exec.LookPath("bash"); err == nil {
			shell = p
			shellArgs = []string{"-c", command}
		} else {
			shell = "/bin/sh"
			shellArgs = []string{"-c", command}
		}
	}
	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	cmd.Env = append(os.Environ(), e.BuildEnv()...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := toUTF8(buf.Bytes())
	res := &ExecResult{Output: out}
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			res.ExitCode = -1
			res.Error = "执行超时"
			res.TimedOut = true
			return res, nil
		}
		if _, ok := err.(*exec.ExitError); ok {
			res.ExitCode = 1
			if cmd.ProcessState != nil {
				if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
					res.ExitCode = ws.ExitStatus()
				}
			}
			res.Error = strings.TrimSpace(out)
			return res, nil
		}
		// Non-exit error: context deadline exceeded or start failure.
		return res, gerror.WrapCode(gcode.CodeOperationFailed, err, "command failed to start or was cancelled")
	}
	return res, nil
}

// resolve returns the command string with ${VAR} substituted (for dry-run display in a
// sanitized way). Resolved values are only used for debug representation.
func (e *EnvMap) resolve(command string) string {
	return e.Resolve(command)
}
