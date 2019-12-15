package core

import (
	"os"
	"strings"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

func LogEnabled() bool {
	env := strings.TrimSpace(os.Getenv("ULM_DEBUG"))
	return env != "" && env != "0"
}
