package exec

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
)

func truish(s string) bool {
	return s == "1" || strings.ToLower(s) == "true" || strings.ToLower(s) == "on"
}

var re = regexp.MustCompile("\\${(\\w+)(:([a-zA-Z0-9%.]+))?}")

// replaceFormatted replaces all occurrances of ${key} with val from the kv map.
// All keys of kv must exist inside the string to apply replacements to
func replaceFormatted(s string, kv map[string]interface{}) (string, error) {
	matches := re.FindAllStringSubmatch(s, -1)

	for len(matches) > 0 {
		for _, m := range matches {
			key := m[1]
			val, ok := kv[key]
			if !ok {
				return "", errors.New("could not find match for " + m[0])
			}

			// apply format
			format := m[3]
			if format != "" {
				val = fmt.Sprintf(format, val)
			}

			// update string
			literalMatch := m[0]
			s = strings.ReplaceAll(s, literalMatch, fmt.Sprintf("%v", val))
		}

		// update matches
		matches = re.FindAllStringSubmatch(s, -1)
	}

	return s, nil
}

func contextWithTimeout(timeout time.Duration) context.Context {
	ctx := context.Background()
	if timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, timeout)
	}
	return ctx
}

func execWithStringResult(ctx context.Context, script string) (string, error) {
	Logger.Println("exec script: " + script)

	args, err := shellquote.Split(script)
	if err != nil {
		return "", err
	}

	if len(args) < 1 {
		return "", errors.New("exec: missing script")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	Logger.Println("exec result: " + strings.TrimSpace(string(b)))

	return strings.TrimSpace(string(b)), nil
}

func execWithFloatResult(ctx context.Context, script string) (float64, error) {
	s, err := execWithStringResult(ctx, script)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}
