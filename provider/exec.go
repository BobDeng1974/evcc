package provider

import (
	"context"
	"log"
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

		log.Println("exec result: " + strings.TrimSpace(string(b)))
		return strings.TrimSpace(string(b)), nil
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
