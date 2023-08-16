package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	entity "gitlab.com/pracmaticreviews/golang-gin-poc/Entity"
	service "gitlab.com/pracmaticreviews/golang-gin-poc/Service"
)

type VideoController interface {
	FindAll() []entity.Video
	Save(context *gin.Context) error
	Delete(context *gin.Context) error
}

type controller struct {
	service service.VideoService
}

func New(newService service.VideoService) VideoController {
	return &controller{
		service: newService,
	}
}

func (c *controller) FindAll() []entity.Video {
	return c.service.FindAll()
}

func (c *controller) Save(context *gin.Context) error {
	var video entity.Video
	if err := context.ShouldBindJSON(&video); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	c.service.Save(video)
	context.JSON(http.StatusOK, gin.H{"message": "Video saved"})
	return nil
}

func (c *controller) Delete(context *gin.Context) error {
	title := context.Param("title")
	deletedVideo := c.service.Delete(title)
	if deletedVideo.Title == "" {
		context.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return nil
	}
	context.JSON(http.StatusOK, gin.H{"message": "Video deleted"})
	return nil
}
