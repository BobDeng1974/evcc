package core

import (
	"io/ioutil"
	"log"

	"github.com/andig/evcc/api"
)

var Logger api.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
