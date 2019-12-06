package exec

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
)

var re = regexp.MustCompile("\\${(\\w+)(:([a-zA-Z0-9%.]+))?}")

// TODO replace multiple
func replaceFormatted(s string, kv map[string]interface{}) (string, error) {
	matches := re.FindAllStringSubmatch(s, -1)

	for len(matches) > 0 {
		for _, m := range matches {
			log.Println(m)
			k := m[1]
			if v, ok := kv[k]; ok {
				format := m[3]
				if format != "" {
					v = fmt.Sprintf(format, v)
				}

				// update string
				lit := m[0]
				s = strings.ReplaceAll(s, lit, fmt.Sprintf("%v", v))
			} else {
				return "", errors.New("could not find match for " + m[0])
			}
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
	args, err := shellquote.Split(script)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func execWithFloatResult(ctx context.Context, script string) (float64, error) {
	s, err := execWithStringResult(ctx, script)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}
