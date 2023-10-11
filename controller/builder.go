package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func RunServer(port *string) {
	router := gin.Default()
	router.GET("/:param", HandlerEtcd)
	if err := router.Run(*port); err != nil {
		fmt.Println("Error running server: ", err)
	}
}
