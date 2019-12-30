package core

import (
	"io/ioutil"
	"log"

	"github.com/andig/evcc/api"
)

// Logger is the package-local logger
var Logger api.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
