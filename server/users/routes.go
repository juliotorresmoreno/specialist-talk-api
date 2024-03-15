package users

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
	"github.com/juliotorresmoreno/specialist-talk-api/miniostorage"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
	"github.com/minio/minio-go/v7"
)

var log = logger.SetupLogger()

type UsersRouter struct {
}

func SetupAPIRoutes(r *gin.RouterGroup) {
	users := &UsersRouter{}

	r.GET("", users.find)
	r.GET("/:username", users.findOne)
	r.GET("/me", users.findMe)
	r.PATCH("/me", users.updateMe)
}

type User struct {
	ID           uint       `json:"id"`
	Verified     bool       `json:"verified"`
	FirstName    string     `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName     string     `json:"last_name" validate:"omitempty,min=2,max=100"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email" validate:"omitempty,email"`
	Username     string     `json:"username"`
	PhotoURL     string     `json:"photo_url"`
	Photo        string     `json:"photo,omitempty" gorm:"-"`
	Phone        string     `json:"phone"`
	Business     string     `json:"business"`
	PositionName string     `json:"position_name"`
	Url          string     `json:"url" validate:"omitempty,url"`
	Description  string     `json:"description" validate:"max=1000"`
	CreationAt   time.Time  `json:"creation_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (h *UsersRouter) findOne(c *gin.Context) {
	_, err := utils.ValidateSession(c)
	if err != nil {
		utils.Response(c, err)
		return
	}

	username := c.Param("username")
	user := &User{}
	err = db.DefaultClient.Model(&models.User{}).
		Where(User{Username: username}).
		First(user).Error
	if err != nil {
		log.Error("Error getting user", err)
		utils.Response(c, err)
		return
	}

	c.JSON(200, user)
}

func (h *UsersRouter) find(c *gin.Context) {
	_, err := utils.ValidateSession(c)
	if err != nil {
		utils.Response(c, err)
		return
	}

	q := "%" + c.Query("q") + "%"

	conn := db.DefaultClient
	users := &[]*User{}
	err = conn.Model(&models.User{}).
		Where("full_name LIKE ? or email LIKE ?", strings.ReplaceAll(q, " ", "%"), q).
		Where("deleted_at IS NULL").
		Find(users).Error
	if err != nil {
		log.Error("Error getting users", err)
		utils.Response(c, err)
		return
	}

	c.JSON(200, users)
}

func (h *UsersRouter) findMe(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		log.Error("Error validating session", err)
		utils.Response(c, err)
		return
	}

	conn := db.DefaultClient
	user := &User{}
	tx := conn.Model(&models.User{}).Where("id = ?", session.ID).First(user)
	if tx.Error != nil {
		log.Error("Error getting users", tx.Error)
		utils.Response(c, tx.Error)
		return
	}

	c.JSON(200, user)
}

type UpdateValidationErrors struct {
	Name         string `json:"name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Business     string `json:"business,omitempty"`
	PositionName string `json:"position_name,omitempty"`
	Url          string `json:"url,omitempty"`
	Description  string `json:"description,omitempty"`
}

func (h *UsersRouter) updateMe(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		log.Error("Error validating session", err)
		utils.Response(c, err)
		return
	}

	payload := &User{}
	if err := c.ShouldBind(payload); err != nil {
		log.Error("Error binding JSON", err)
		utils.Response(c, err)
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		errorsMap := utils.ParseErrors(err.(validator.ValidationErrors))

		customErrors := UpdateValidationErrors{
			Name:         errorsMap["Name"],
			LastName:     errorsMap["LastName"],
			Business:     errorsMap["Business"],
			PositionName: errorsMap["PositionName"],
			Url:          errorsMap["Url"],
			Description:  errorsMap["Description"],
		}
		log.Error("Error validating payload", err)
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	photoURL := ""
	if payload.Photo != "" {
		photoURL, err = h.uploadPhoto(payload.Photo)
		if err != nil {
			log.Error("Error uploading photo", err)
			utils.Response(c, err)
			return
		}
	}

	user := &models.User{
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		FullName:     fullName(session, payload),
		Business:     payload.Business,
		PositionName: payload.PositionName,
		Url:          payload.Url,
		Description:  payload.Description,
		Phone:        payload.Phone,
		PhotoURL:     photoURL,
		UpdatedAt:    time.Now(),
	}

	conn := db.DefaultClient
	tx := conn.Model(&models.User{}).Where("id = ?", session.ID).Updates(user)
	if tx.Error != nil {
		log.Error("Error updating user", tx.Error)
		utils.Response(c, tx.Error)
		return
	}

	c.JSON(200, gin.H{"message": "Profile updated successfully"})
}

func (h *UsersRouter) uploadPhoto(photo string) (string, error) {
	bucket := os.Getenv("MINIO_BUCKET")
	miniostorage, err := miniostorage.NewClient()
	if err != nil {
		return "", err
	}
	attachment, err := utils.ParseBase64File(photo)
	if err != nil {
		return "", err
	}
	decoded, err := base64.StdEncoding.DecodeString(attachment)
	if err != nil {
		return "", err
	}

	converted, err := utils.ConvertToJPEG(bytes.NewBufferString(string(decoded)))
	if err != nil {
		return "", err
	}

	objectName := utils.GenerateRandomFileName("photo_", ".jpeg")
	_, err = miniostorage.PutObject(
		context.Background(), bucket, objectName, converted,
		int64(converted.Len()), minio.PutObjectOptions{},
	)
	if err != nil {
		return "", err
	}

	photoURL := os.Getenv("ASSETS_PATH") + "/" + objectName

	return photoURL, nil
}

func fullName(old *utils.User, payload *User) string {
	if payload.FirstName != "" && payload.LastName != "" {
		return strings.ToLower(payload.FirstName + " " + payload.LastName)
	}
	if payload.FirstName != "" {
		return strings.ToLower(payload.FirstName + " " + old.LastName)
	}
	if payload.LastName != "" {
		return strings.ToLower(old.FirstName + " " + payload.LastName)
	}
	return ""
}
