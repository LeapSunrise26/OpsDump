// Package service implements business logic for ops-dump.
package service

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// EnvMap holds KEY=VALUE loaded from the .env file (runtime injected values).
type EnvMap struct {
	values map[string]string
	file   string
}

// NewEnv reads a .env file into an EnvMap. Missing file yields an empty env without error.
func NewEnv(file string) *EnvMap {
	e := &EnvMap{values: map[string]string{}, file: file}
	if file == "" {
		return e
	}
	f, err := os.Open(file)
	if err != nil {
		return e
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		// Strip quotes only when they wrap the entire value. Quotes inside a
		// command value, such as echo "hello world", must be preserved.
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') ||
			(v[0] == '\'' && v[len(v)-1] == '\'')) {
			v = v[1 : len(v)-1]
		}
		e.values[k] = v
	}
	return e
}

// Get returns the value of a key.
func (e *EnvMap) Get(key string) string {
	return e.values[key]
}

// Set overrides a value (used by CLI flags like for dry-run).
func (e *EnvMap) Set(key, val string) {
	e.values[key] = val
}

// varRe matches ${KEY} or $KEY (KEY = alphanumeric/underscore/dot).
var varRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_.]*)\}|\$([A-Za-z_][A-Za-z0-9_.]*)`)

// Resolve replaces every ${KEY} or $KEY with the loaded or process environment value.
// Undefined keys are left as-is so an operator can spot a misconfiguration.
func (e *EnvMap) Resolve(s string) string {
	return varRe.ReplaceAllStringFunc(s, func(match string) string {
		// Extract key: group 1 for ${KEY}, group 2 for $KEY.
		sub := varRe.FindStringSubmatch(match)
		key := sub[1]
		if key == "" {
			key = sub[2]
		}
		if value, ok := e.values[key]; ok {
			return value
		}
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return match // undefined — preserve original form ($4 stays $4)
	})
}

// ResolveForRecord resolves non-sensitive values while masking credentials.
func (e *EnvMap) ResolveForRecord(s string) string {
	return varRe.ReplaceAllStringFunc(s, func(match string) string {
		sub := varRe.FindStringSubmatch(match)
		key := sub[1]
		if key == "" {
			key = sub[2]
		}
		value, ok := e.values[key]
		if !ok {
			value, ok = os.LookupEnv(key)
		}
		if !ok {
			return match // undefined — preserve original form
		}
		if isSensitiveKey(key) {
			return "***"
		}
		return value
	})
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "password") ||
		strings.Contains(key, "passwd") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.HasSuffix(key, "_key")
}

// BuildEnv returns the environment as []string ("KEY=VAL") for os/exec.
func (e *EnvMap) BuildEnv() []string {
	out := make([]string, 0, len(e.values))
	for k, v := range e.values {
		out = append(out, k+"="+v)
	}
	return out
}
