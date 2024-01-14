package main

import (
	"flag"
	"job-allocator/src/controller"
	"job-allocator/src/etcdClient"
)

func init() {
	go etcdClient.Setup()

}

func main() {
	port := flag.String("port", ":8080", "HTTP server port")
	flag.Parse()

	controller.RunServer(port)
}
