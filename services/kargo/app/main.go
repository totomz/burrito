package main

import (
	"fmt"
	"github.com/croccocode/golibz/env"
	"github.com/totomz/template-burrito/common/httpserver"
	"github.com/totomz/template-burrito/services/kargo"
	"strconv"
)

func main() {
	// The service has your business logic
	service := kargo.NewService()

	// Enable http metrics and traces, with some default labels
	// Passing nil to NewHttpServer() to disable the metrics is not supported, I'm sorry
	metrics := httpserver.StandardMetrics{
		Namespace: "kargo",
		ConstLabels: map[string]string{
			"env": "prod",
		},
	}

	server := httpserver.NewHttpServer(service, &metrics)

	port, _ := strconv.Atoi(env.GetenvOrDefault("BIND_PORT", "8443"))

	println(fmt.Sprintf("Listening on 0.0.0:%v", port))

	server.StartAsync("0.0.0.0", port)

	// start server
	println(fmt.Sprintf("server started"))

	// Wait forever
	<-server.StopChan
	println(fmt.Sprintf("bye"))
}
