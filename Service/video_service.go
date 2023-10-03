package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	entity "gitlab.com/pracmaticreviews/golang-gin-poc/Entity"
)

type VideoService interface {
	Save(entity.Video) (entity.Video, error)
	Delete(string) error
	FindAll() ([]entity.Video, error)
	FindByID(string) (entity.Video, error)
	VideoExists(string) bool
	Update(*entity.Video, map[string]string) error
}

type videoService struct {
	client     *mongo.Client
	collection *mongo.Collection
	redis      *redis.Client
}

func NewMongoVideoService(client *mongo.Client, dbName, collectionName string, redisClient *redis.Client) VideoService {
	collection := client.Database(dbName).Collection(collectionName)
	return &videoService{
		client:     client,
		collection: collection,
		redis:      redisClient,
	}
}

func (service *videoService) Save(newVideo entity.Video) (entity.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := service.collection.InsertOne(ctx, newVideo)
	if err != nil {
		return entity.Video{}, err
	}

	// Cache the new video
	cacheKey := "videos"
	cachedVideos, err := service.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var videos []entity.Video
		if err := json.Unmarshal([]byte(cachedVideos), &videos); err == nil {
			// Add the new video to the cached data
			videos = append(videos, newVideo)

			// Store the updated data in the cache
			jsonVideos, _ := json.Marshal(videos)
			err = service.redis.Set(ctx, cacheKey, jsonVideos, 5*time.Minute).Err()
			if err != nil {
				fmt.Printf("Error updating cached videos in Redis: %v", err)
			}
		}
	}

	return newVideo, nil
}

func (service *videoService) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": id}
	_, err := service.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	// Update the cache after successful deletion
	cacheKey := "videos"
	cachedVideos, err := service.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var videos []entity.Video
		if err := json.Unmarshal([]byte(cachedVideos), &videos); err == nil {
			// Remove the deleted video from the cached data
			updatedVideos := make([]entity.Video, 0, len(videos))
			for _, video := range videos {
				if video.ID != id {
					updatedVideos = append(updatedVideos, video)
				}
			}

			// Store the updated data in the cache
			jsonVideos, _ := json.Marshal(updatedVideos)
			err = service.redis.Set(ctx, cacheKey, jsonVideos, 5*time.Minute).Err()
			if err != nil {
				fmt.Printf("Error updating cached videos in Redis: %v", err)
			}
		}
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

	_, err := service.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return err
	}

	// Update the cache after successful update
	cacheKey := "videos"
	cachedVideos, err := service.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var videos []entity.Video
		if err := json.Unmarshal([]byte(cachedVideos), &videos); err == nil {
			// Find the video to update in the cached data
			for i, video := range videos {
				if video.ID == existingVideo.ID {
					// Update the cached video with new values
					for key, value := range updateFields {
						switch key {
						case "title":
							videos[i].Title = value
						case "description":
							videos[i].Description = value
						case "url":
							videos[i].URL = value
						}
					}
					break
				}
			}

			// Store the updated data in the cache
			jsonVideos, _ := json.Marshal(videos)
			err = service.redis.Set(ctx, cacheKey, jsonVideos, 5*time.Minute).Err()
			if err != nil {
				fmt.Printf("Error updating cached videos in Redis: %v", err)
			}
		}
	}

	return nil
}

func (service *videoService) FindAll() ([]entity.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Check if data already in cache
	cachedVideos, err := service.redis.Get(ctx, "videos").Result()
	if err == nil {
		var videos []entity.Video
		if err := json.Unmarshal([]byte(cachedVideos), &videos); err == nil {
			return videos, nil
		}
	}

	cursor, err := service.collection.Find(ctx, bson.M{})
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
	if err := service.collection.FindOne(ctx, filter).Decode(&video); err != nil {
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
	count, err := service.collection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Printf("Error checking if video exists: %v", err)
		return false
	}

	return count > 0
}