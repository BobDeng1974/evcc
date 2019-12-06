package ulm

import (
	"os"
	"strings"
)

func LogEnabled() bool {
	env := strings.TrimSpace(os.Getenv("ULM_DEBUG"))
	return env != "" && env != "0"
}
