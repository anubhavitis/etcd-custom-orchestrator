package controller

import (
	"job-allocator/src/etcdClient"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlerEtcd(c *gin.Context) {
	// Retrieve the query parameter named "param" from the request
	param := c.Param("param")

	// Check if the query parameter is empty
	if param == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or empty 'param' query parameter"})
		return
	}

	name := "job-" + param
	newJob := etcdClient.Job{Name: name}

	etcdClient.UpdateJobsConfig(newJob)
}
