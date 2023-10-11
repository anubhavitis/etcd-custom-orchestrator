package main

import (
	"etcd_test/controller"
	"etcd_test/etcdClient"
	"flag"
)

func init() {
	go etcdClient.Setup()

}

func main() {
	port := flag.String("port", ":8080", "HTTP server port")
	flag.Parse()

	controller.RunServer(port)
}
