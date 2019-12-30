package cmd

import (
	"log"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/provider"
)

func stringProvider(pc *ProviderConfig) (res api.StringProvider) {
	switch pc.Type {
	case "exec", "script":
		exec := &provider.Exec{}
		res = exec.StringProvider(pc.Cmd)
	default:
		log.Fatalf("invalid provider type %s", pc.Type)
	}
	return
}

func boolProvider(pc *ProviderConfig) (res api.BoolProvider) {
	switch pc.Type {
	case "exec", "script":
		exec := &provider.Exec{}
		res = exec.BoolProvider(pc.Cmd)
	default:
		log.Fatalf("invalid provider type %s", pc.Type)
	}
	return
}

func intProvider(pc *ProviderConfig) (res api.IntProvider) {
	switch pc.Type {
	case "mqtt":
		res = mq.IntProvider(pc.Topic)
	case "exec", "script":
		exec := &provider.Exec{}
		res = exec.IntProvider(pc.Cmd)
	default:
		log.Fatalf("invalid provider type %s", pc.Type)
	}
	return
}

func floatProvider(pc *ProviderConfig) (res api.FloatProvider) {
	switch pc.Type {
	case "mqtt":
		res = mq.FloatProvider(pc.Topic)
	case "exec", "script":
		exec := &provider.Exec{}
		res = exec.FloatProvider(pc.Cmd)
	default:
		log.Fatalf("invalid provider type %s", pc.Type)
	}
	return
}

func boolSetter(param string, pc *ProviderConfig) (res api.BoolSetter) {
	switch pc.Type {
	case "exec", "script":
		exec := &provider.Exec{}
		res = exec.BoolSetter(param, pc.Cmd)
	default:
		log.Fatalf("invalid setter type %s", pc.Type)
	}
	return
}

func intSetter(param string, pc *ProviderConfig) (res api.IntSetter) {
	switch pc.Type {
	case "exec", "script":
		exec := &provider.Exec{}
		res = exec.IntSetter(param, pc.Cmd)
	default:
		log.Fatalf("invalid setter type %s", pc.Type)
	}
	return
}
