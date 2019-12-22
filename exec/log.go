package exec

import (
	"io/ioutil"
	"log"

	"github.com/andig/ulm/api"
)

var Logger api.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
