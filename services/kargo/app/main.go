package main

import (
	"fmt"
	burrito_common "github.com/totomz/template-burrito/common/burrito-common"
	"github.com/totomz/template-burrito/common/httpserver"
	"github.com/totomz/template-burrito/services/kargo"
	"strconv"
)

func main() {
	service := kargo.NewService()

	server := httpserver.NewHttpServer(service)

	port, _ := strconv.Atoi(burrito_common.GetenvOrDefault("BIND_PORT", "8443"))

	println(fmt.Sprintf("Listening on 0.0.0:%v", port))

	server.StartAsync("0.0.0.0", port)

	// start server
	println(fmt.Sprintf("server started"))

	// Wait forever
	<-server.StopChan
	println(fmt.Sprintf("bye"))
}
