package provider

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/andig/ulm/api"
	"github.com/kballard/go-shellquote"
)

type Exec struct{}

func (e *Exec) StringValue(script string) api.StringProvider {
	args, err := shellquote.Split(script)
	if err != nil {
		panic(err)
	} else if len(args) < 1 {
		panic("exec: missing script")
	}

	// return func to access cached value
	return func(ctx context.Context) (string, error) {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(b)), nil
	}
}

func (e *Exec) IntValue(script string) api.IntProvider {
	exec := e.StringValue(script)

	// return func to access cached value
	return func(ctx context.Context) (int64, error) {
		s, err := exec(ctx)
		if err != nil {
			return 0, err
		}

		return strconv.ParseInt(s, 10, 64)
	}
}

func (e *Exec) FloatValue(script string) api.FloatProvider {
	exec := e.StringValue(script)

	// return func to access cached value
	return func(ctx context.Context) (float64, error) {
		s, err := exec(ctx)
		if err != nil {
			return 0, err
		}

		return strconv.ParseFloat(s, 64)
	}
}
