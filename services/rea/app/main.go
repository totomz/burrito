package main

import (
	"flag"
	"log/slog"
	"os"

	common "github.com/totomz/burrito/common"
	"github.com/totomz/burrito/services/rea"
)

func main() {
	common.SetDefaultLogger()

	serviceName := ""
	flag.StringVar(&serviceName, "deity", "", "the name of the God to generate")

	flag.Parse()

	slog.Info("generating service", "name", serviceName)
	seed := rea.ServiceSeed{
		ServiceName: serviceName,
	}

	err := rea.CreateServiceDirectory(seed)
	if err != nil {
		slog.Error("error generating service folder", "error", err)
		os.Exit(11)
	}

	slog.Info("service directory created", "path", seed.ServicePath())

	err = rea.CloneTemplate(seed)
	if err != nil {
		slog.Error("generation failed", "error", err)
	}

}
