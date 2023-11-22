package main

import (
	"context"
	"io"
	"net/http"
	"os"

	//"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	controller "videoAPI/Controller"
	middlewares "videoAPI/Middlewares"
	service "videoAPI/Service"
	_ "videoAPI/docs"
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

	r.Use(gin.Recovery(), middlewares.Logger()) // , middlewares.BasicAuth()

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/videos", func(context *gin.Context) {

		err := VideoController.Save(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	r.POST("/signup", func(context *gin.Context) {
		VideoController.SignUp(context)
	})

	r.POST("/login", func(context *gin.Context) {
		VideoController.LogIn(context)
	})

	r.GET("/videos", func(context *gin.Context) {

		err := VideoController.HandleVideoSearchAndPaginate(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	r.GET("/videos/all", func(context *gin.Context) {
		err := VideoController.FindAll(context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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

// @title  Video API
func main() {

	setupLogOutput()

	client, err := setupMongoDB()
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	setupRedis()

	videoService = service.NewMongoVideoService(client, "trungdb", "trungcl", "usercl", redisClient)
	VideoController = controller.New(videoService)

	server := setupRouter()

	server.Run(":8080")
}
