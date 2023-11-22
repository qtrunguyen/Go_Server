package controller

import (
	"net/http"
	"time"

	entity "videoAPI/Entity"
	service "videoAPI/Service"
	_ "videoAPI/docs"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type VideoController interface {
	FindAll(context *gin.Context) error
	Save(context *gin.Context) error
	Delete(context *gin.Context) error
	FindByID(context *gin.Context) error
	Update(context *gin.Context) error
	HandleVideoSearchAndPaginate(context *gin.Context) error

	//Authorization
	SignUp(context *gin.Context) error
	LogIn(context *gin.Context) error
}

type controller struct {
	service service.VideoService
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SignUpResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func New(newService service.VideoService) VideoController {
	return &controller{
		service: newService,
	}
}

// @Summary Get all videos
// @Description Get all videos in DB
// @ID find-all-videos
// @Produce  json
// @Success 200 {array} entity.Video
// @Failure 400 {object} ErrorResponse
// @Router /videos/all [get]
func (c *controller) FindAll(context *gin.Context) error {
	findVideos, err := c.service.FindAll()

	if err != nil {
		context.JSON(http.StatusNotFound, ErrorResponse{"Videos not found"})
		return err
	}

	context.JSON(http.StatusOK, findVideos)

	return nil
}

// @Summary Save a video
// @Description Save a video to the system
// @ID save-video
// @Accept  json
// @Produce  json
// @Param video body entity.Video true "Video to save"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /videos [post]
func (c *controller) Save(context *gin.Context) error {
	var video entity.Video

	if err := context.ShouldBindJSON(&video); err != nil {
		context.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return err
	}

	if c.service.VideoExists(video.ID) {
		context.JSON(http.StatusConflict, ErrorResponse{"Video ID Already Exists"})
		return nil
	}

	c.service.Save(video)
	context.JSON(http.StatusOK, SuccessResponse{"Video saved"})

	return nil
}

// @Summary Delete a video by ID
// @Description Delete a video by its ID
// @ID delete-video
// @Produce json
// @Param id path string true "Video ID to delete"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Router /videos/{id} [delete]
func (c *controller) Delete(context *gin.Context) error {
	id := context.Param("id")

	if !c.service.VideoExists(id) {
		context.JSON(http.StatusNotFound, ErrorResponse{"Video not found"})
		return nil
	}

	err := c.service.Delete(id)
	if err != nil {
		return err
	}

	context.JSON(http.StatusOK, SuccessResponse{"Video deleted"})

	return nil
}

// @Summary Find a video by ID
// @Description Find a video by its ID
// @ID find-video
// @Produce json
// @Param id path string true "Video ID to find"
// @Success 200 {object} entity.Video
// @Failure 404 {object} ErrorResponse
// @Router /videos/{id} [get]
func (c *controller) FindByID(context *gin.Context) error {
	id := context.Param("id")
	findVideo, _ := c.service.FindByID(id)

	if findVideo.ID == "" {
		context.JSON(http.StatusNotFound, ErrorResponse{"Video not found"})
		return nil
	}

	context.JSON(http.StatusOK, findVideo)

	return nil
}

// @Summary Update a video by ID
// @Description Update a video by its ID
// @ID update-video
// @Produce json
// @Param id path string true "Video ID to update"
// @Param updateFields body map[string]string true "Fields to update"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /videos/{id} [put]
func (c *controller) Update(context *gin.Context) error {
	var updateFields map[string]string
	if err := context.ShouldBindJSON(&updateFields); err != nil {
		context.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return nil
	}

	id := context.Param("id")
	existingVideo, _ := c.service.FindByID(id)

	if existingVideo.ID == "" {
		context.JSON(http.StatusNotFound, ErrorResponse{"Video not found"})
		return nil
	}

	c.service.Update(&existingVideo, updateFields)

	context.JSON(http.StatusOK, SuccessResponse{"Video updated"})

	return nil
}

// @Summary Search and paginate videos
// @Description Search and paginate videos
// @ID search-and-paginate
// @Produce json
// @Param page query string false "Page number"
// @Param q query string false "Search query"
// @Success 200 {array} entity.Video
// @Failure 500 {object} ErrorResponse
// @Router /videos [get]
func (c *controller) HandleVideoSearchAndPaginate(context *gin.Context) error {
	page := context.DefaultQuery("page", "1")
	q := context.Query("q")

	videos, err := c.service.SearchAndPaginate(page, q, 10) // Adjust perPage as needed
	if err != nil {
		context.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
		return err
	}
	context.JSON(http.StatusOK, videos)
	return nil
}

// @Summary Sign up a new user
// @Description Sign up a new user
// @ID sign-up
// @Produce json
// @Param body body object true "User data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /signup [post]
func (c *controller) SignUp(context *gin.Context) error {
	var user entity.User

	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return err
	}

	err := c.service.CreateUser(user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to create user"})
		return err
	}

	context.JSON(http.StatusOK, SuccessResponse{"User created successfully"})

	return nil
}

// @Summary Log in a user
// @Description Log in a user
// @ID log-in
// @Produce json
// @Param body body object true "User data"
// @Success 200 {object} SignUpResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /login [post]
func (c *controller) LogIn(context *gin.Context) error {
	var login_user entity.User

	if err := context.ShouldBindJSON(&login_user); err != nil {
		context.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return err
	}

	user, err := c.service.GetUserByEmail(login_user.Email)
	if err != nil {
		context.JSON(http.StatusUnauthorized, ErrorResponse{"Invalid email"})
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login_user.Password))
	if err != nil {
		context.JSON(http.StatusUnauthorized, ErrorResponse{"Invalid password"})
		return err
	}

	token, err := generateJWTToken(user.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to generate JWT token"})
		return err
	}

	context.JSON(http.StatusOK, SignUpResponse{"Login successful", token})

	return nil
}

func generateJWTToken(email string) (string, error) {
	const secretKey = "vcsbackend"
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
