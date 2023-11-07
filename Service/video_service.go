package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	entity "videoAPI/Entity"
)

type VideoService interface {
	Save(entity.Video) (entity.Video, error)
	Delete(string) error
	FindAll() ([]entity.Video, error)
	FindByID(string) (entity.Video, error)
	VideoExists(string) bool
	Update(*entity.Video, map[string]string) error
	SearchAndPaginate(string, string, int) ([]entity.Video, error)

	//Authorization
	CreateUser(email string, password string) error
	GetUserByEmail(email string) (User, error)
}

type videoService struct {
	client          *mongo.Client
	videoCollection *mongo.Collection
	userCollection  *mongo.Collection
	redis           *redis.Client
}

type User struct {
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

func NewMongoVideoService(client *mongo.Client, dbName, videoCollectionName string, userCollectionName string, redisClient *redis.Client) VideoService {
	videoCollection := client.Database(dbName).Collection(videoCollectionName)
	userCollection := client.Database(dbName).Collection(userCollectionName)
	return &videoService{
		client:          client,
		videoCollection: videoCollection,
		userCollection:  userCollection,
		redis:           redisClient,
	}
}

func (service *videoService) Save(newVideo entity.Video) (entity.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := service.videoCollection.InsertOne(ctx, newVideo)
	if err != nil {
		return entity.Video{}, err
	}

	return newVideo, nil
}

func (service *videoService) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": id}
	_, err := service.videoCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	// Update the cache after successful deletion
	cacheKey := "video:" + id
	err = service.redis.Del(ctx, cacheKey).Err()
	if err != nil {
		fmt.Printf("Error deleting cached video from Redis: %v", err)
	}

	return nil
}

func (service *videoService) Update(existingVideo *entity.Video, updateFields map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	for key, value := range updateFields {
		update[key] = value
	}

	filter := bson.M{"id": existingVideo.ID}
	updateDoc := bson.M{"$set": update}

	_, err := service.videoCollection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return err
	}

	// Remove the cached individual video
	cacheKey := "video:" + existingVideo.ID
	err = service.redis.Del(ctx, cacheKey).Err()
	if err != nil {
		fmt.Printf("Error deleting cached video from Redis: %v", err)
	}

	return nil
}

func (service *videoService) FindAll() ([]entity.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Try fetch from the cache first
	cachedVideos, err := service.redis.Get(ctx, "videos").Result()
	if err == nil {
		var videos []entity.Video
		if err := json.Unmarshal([]byte(cachedVideos), &videos); err == nil {
			return videos, nil
		}
	}

	//Retrieved from database
	cursor, err := service.videoCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videos []entity.Video
	if err := cursor.All(ctx, &videos); err != nil {
		return nil, err
	}

	//Store data in the cache
	jsonVideos, _ := json.Marshal(videos)
	err = service.redis.Set(ctx, "videos", jsonVideos, 5*time.Minute).Err()
	if err != nil {
		fmt.Printf("Error caching videos in Redis: %v", err)
	}

	fmt.Printf("Found %d videos\n", len(videos))
	return videos, nil
}

func (service *videoService) FindByID(id string) (entity.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Try fetch from the cache first
	cacheKey := "video:" + id
	cachedVideo, err := service.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var video entity.Video
		if err := json.Unmarshal([]byte(cachedVideo), &video); err == nil {
			return video, nil
		}
	}

	//If not in the cache, retrieve from database
	filter := bson.M{"id": id}
	var video entity.Video
	if err := service.videoCollection.FindOne(ctx, filter).Decode(&video); err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Video{}, errors.New("Video not found")
		}
		return entity.Video{}, err
	}

	//Store retrieved video in cache
	jsonVideo, _ := json.Marshal(video)
	err = service.redis.Set(ctx, cacheKey, jsonVideo, 5*time.Minute).Err()
	if err != nil {
		fmt.Printf("Error caching video in Redis: %v", err)
	}

	return video, nil
}

func (service *videoService) VideoExists(id string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": id}
	count, err := service.videoCollection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Printf("Error checking if video exists: %v", err)
		return false
	}

	return count > 0
}

func (service *videoService) SearchAndPaginate(page string, query string, perPage int) ([]entity.Video, error) {
	// Define the MongoDB query based on the query parameter
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		return nil, err
	}

	filter := bson.M{}
	if query != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"url": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	skip := (pageNum - 1) * perPage

	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(perPage))
	cur, err := service.videoCollection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	var videos []entity.Video
	for cur.Next(context.TODO()) {
		var video entity.Video
		if err := cur.Decode(&video); err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func (service *videoService) CreateUser(email string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}

	user := bson.M{
		"email":    email,
		"password": hash,
	}

	_, err = service.userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}

	return nil
}

func (service *videoService) GetUserByEmail(email string) (User, error) {
	var user User
	filter := bson.M{"email": email}
	err := service.userCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
