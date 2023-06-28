package main

import (
	"fmt"
	"os"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/todo-lists-app/ping-service/internal/config"
	"github.com/todo-lists-app/ping-service/internal/service"
)

var (
	BuildVersion = "dev"
	BuildHash    = "none"
	ServiceName  = "ping-service"
)

func main() {
	logs.Local().Info(fmt.Sprintf("Starting %s", ServiceName))
	logs.Local().Info(fmt.Sprintf("Version: %s, Hash: %s", BuildVersion, BuildHash))

	cfg, err := config.Build()
	if err != nil {
		_ = logs.Local().Errorf("config: %+v, err: %v", cfg, err)
		return
	}

	command := ""
	if !cfg.Local.Development {
		if len(os.Args) > 1 {
			command = os.Args[1]
		}
	}

	switch command {
	case "health":
		if err := service.NewService(cfg).Health(); err != nil {
			_ = logs.Local().Errorf("health check: %v", err)
			return
		}

	default:
		if err := service.NewService(cfg).Start(); err != nil {
			_ = logs.Local().Errorf("start service: %v", err)
			return
		}
	}
}
