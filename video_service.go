package service

import entity "gitlab.com/pracmaticreviews/golang-gin-poc/Entity"

type VideoService interface {
	Save(entity.Video) entity.Video
	Delete(string) entity.Video
	FindAll() []entity.Video
	FindByTitle(string) entity.Video
}

type videoService struct {
	videos []entity.Video
}

func New() VideoService {
	return &videoService{}
}

func (service *videoService) FindByTitle(title string) entity.Video {
	var findVideo entity.Video

	for _, video := range service.videos {
		if video.Title == title {
			findVideo = video
			break
		}
	}

	return findVideo
}

func (service *videoService) Delete(title string) entity.Video {
	var deletedVideo entity.Video

	for i, video := range service.videos {
		if video.Title == title {
			deletedVideo = video

			service.videos[i] = service.videos[len(service.videos)-1]
			service.videos = service.videos[:len(service.videos)-1]
			break
		}
	}

	return deletedVideo
}

func (service *videoService) Save(newVideo entity.Video) entity.Video {
	service.videos = append(service.videos, newVideo)
	return newVideo
}

func (service *videoService) FindAll() []entity.Video {
	return service.videos
}
