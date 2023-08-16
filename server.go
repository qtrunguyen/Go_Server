package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	gindump "github.com/tpkeeper/gin-dump"
	controller "gitlab.com/pracmaticreviews/golang-gin-poc/Controller"
	middlewares "gitlab.com/pracmaticreviews/golang-gin-poc/Middlewares"
	service "gitlab.com/pracmaticreviews/golang-gin-poc/Service"
)

var (
	videoService    service.VideoService       = service.New()
	VideoController controller.VideoController = controller.New(videoService)
)

func setupLogOutput() {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func main() {

	setupLogOutput()
	server := gin.New()

	server.Use(gin.Recovery(), middlewares.Logger(), middlewares.BasicAuth(), gindump.Dump())

	server.GET("/videos", func(context *gin.Context) {
		context.JSON(200, VideoController.FindAll())
	})

	server.POST("/videos", func(context *gin.Context) {
		err := VideoController.Save(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	server.DELETE("/videos/:title", func(c *gin.Context) {
		err := VideoController.Delete(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	server.Run(":8080")
}
