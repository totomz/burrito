package main

import (
	burrito_common "github.com/totomz/template-burrito/common/burrito-common"
	"github.com/totomz/template-burrito/common/httpserver"
	"github.com/totomz/template-burrito/services/kargo"
	"strconv"
)

func main() {
	service := kargo.NewService()

	server := httpserver.NewHttpServer(service, service.Stdout(), service.Stderr())

	port, _ := strconv.Atoi(burrito_common.GetenvOrDefault("BIND_PORT", "8443"))

	service.Stdout().Printf("Listening on 0.0.0:%v", port)

	server.StartAsync("0.0.0.0", port)

	// start server
	service.Stdout().Println("server started")

	// Wait forever
	<-server.StopChan
	service.Stdout().Println("bye")
}
