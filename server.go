package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	controller "gitlab.com/pracmaticreviews/golang-gin-poc/Controller"
	middlewares "gitlab.com/pracmaticreviews/golang-gin-poc/Middlewares"
	service "gitlab.com/pracmaticreviews/golang-gin-poc/Service"
)

var (
	videoService    service.VideoService
	VideoController controller.VideoController
	redisClient     *redis.Client
)

func setupLogOutput() {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func setupMongoDB() (*mongo.Client, error) {
	uri := "mongodb://localhost:27017"
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func setupRedis() (*redis.Client, error) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Ping the Redis server to check the connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}

func setupRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery(), middlewares.Logger(), middlewares.BasicAuth())

	r.POST("/videos", func(context *gin.Context) {
		err := VideoController.Save(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	r.GET("/videos", func(context *gin.Context) {
		videos, err := VideoController.FindAll()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusOK, videos)
	})

	r.GET("/videos/:id", func(context *gin.Context) {
		err := VideoController.FindByID(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	r.DELETE("/videos/:id", func(context *gin.Context) {
		err := VideoController.Delete(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	r.PATCH("/videos/:id", func(context *gin.Context) {
		err := VideoController.Update(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})
	return r
}

func main() {

	setupLogOutput()

	client, err := setupMongoDB()
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	setupRedis()

	videoService = service.NewMongoVideoService(client, "trungdb", "trungcl", redisClient)
	VideoController = controller.New(videoService)

	server := setupRouter()

	server.Run(":8080")
}
