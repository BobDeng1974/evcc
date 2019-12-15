package core

import (
	"log"
	"os"
	"strings"
)

var Log *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

type logger interface {
	Printf(format string, v ...interface{})
}

func LogEnabled() bool {
	env := strings.TrimSpace(os.Getenv("ULM_DEBUG"))
	return env != "" && env != "0"
}
