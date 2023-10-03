package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	entity "gitlab.com/pracmaticreviews/golang-gin-poc/Entity"
	service "gitlab.com/pracmaticreviews/golang-gin-poc/Service"
)

type VideoController interface {
	FindAll() ([]entity.Video, error)
	Save(context *gin.Context) error
	Delete(context *gin.Context) error
	FindByID(context *gin.Context) error
	Update(context *gin.Context) error
}

type controller struct {
	service service.VideoService
}

func New(newService service.VideoService) VideoController {
	return &controller{
		service: newService,
	}
}

func (c *controller) FindAll() ([]entity.Video, error) {
	videos, err := c.service.FindAll()
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func (c *controller) Save(context *gin.Context) error {
	var video entity.Video

	if err := context.ShouldBindJSON(&video); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	if c.service.VideoExists(video.ID) {
		context.JSON(http.StatusConflict, gin.H{"error": "Video ID already exists"})
		return nil
	}

	c.service.Save(video)
	context.JSON(http.StatusOK, gin.H{"message": "Video saved"})

	return nil
}

func (c *controller) Delete(context *gin.Context) error {
	id := context.Param("id")
	err := c.service.Delete(id)
	if err != nil {
		return err
	}

	// if deletedVideo.ID == "" {
	// 	context.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
	// 	return nil
	// }

	context.JSON(http.StatusOK, gin.H{"message": "Video deleted"})

	return nil
}

func (c *controller) FindByID(context *gin.Context) error {
	id := context.Param("id")
	findVideo, _ := c.service.FindByID(id)

	if findVideo.ID == "" {
		context.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return nil
	}

	context.JSON(http.StatusOK, findVideo)

	return nil
}

func (c *controller) Update(context *gin.Context) error {
	var updateFields map[string]string
	if err := context.ShouldBindJSON(&updateFields); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil
	}

	id := context.Param("id")
	existingVideo, _ := c.service.FindByID(id)

	if existingVideo.ID == "" {
		context.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return nil
	}

	c.service.Update(&existingVideo, updateFields)

	context.JSON(http.StatusOK, gin.H{"message": "Video updated"})

	return nil
}
