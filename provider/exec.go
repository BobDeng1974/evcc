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

func (e *Exec) StringProvider(script string) api.StringProvider {
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

func (e *Exec) IntProvider(script string) api.IntProvider {
	exec := e.StringProvider(script)

	// return func to access cached value
	return func(ctx context.Context) (int64, error) {
		s, err := exec(ctx)
		if err != nil {
			return 0, err
		}

		return strconv.ParseInt(s, 10, 64)
	}
}

func (e *Exec) FloatProvider(script string) api.FloatProvider {
	exec := e.StringProvider(script)

	// return func to access cached value
	return func(ctx context.Context) (float64, error) {
		s, err := exec(ctx)
		if err != nil {
			return 0, err
		}

		return strconv.ParseFloat(s, 64)
	}
}

func (e *Exec) BoolProvider(script string) api.BoolProvider {
	exec := e.StringProvider(script)

	// return func to access cached value
	return func(ctx context.Context) (bool, error) {
		s, err := exec(ctx)
		if err != nil {
			return false, err
		}

		return truish(s), nil
	}
}

func (e *Exec) IntSetter(param, script string) api.IntSetter {
	// return func to access cached value
	return func(ctx context.Context, i int64) error {
		cmd, err := replaceFormatted(script, map[string]interface{}{
			param: i,
		})
		if err != nil {
			return err
		}

		exec := e.StringProvider(cmd)
		if _, err := exec(ctx); err != nil {
			return err
		}

		return nil
	}
}

func (e *Exec) BoolSetter(param, script string) api.BoolSetter {
	// return func to access cached value
	return func(ctx context.Context, b bool) error {
		cmd, err := replaceFormatted(script, map[string]interface{}{
			param: b,
		})
		if err != nil {
			return err
		}

		exec := e.StringProvider(cmd)
		if _, err := exec(ctx); err != nil {
			return err
		}

		return nil
	}
}
