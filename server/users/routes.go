package users

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
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
	Email        string     `json:"email" validate:"omitempty,email"`
	Username     string     `json:"username"`
	PhotoURL     string     `json:"photo_url"`
	Phone        string     `json:"phone" validate:"omitempty,min=7,max=15"`
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
		Where("concat(first_name, ' ', last_name) LIKE ? or email LIKE ?", strings.ReplaceAll(q, " ", "%"), q).
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
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	user := &models.User{
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		Business:     payload.Business,
		PositionName: payload.PositionName,
		Url:          payload.Url,
		Description:  payload.Description,
		Phone:        payload.Phone,
		PhotoURL:     payload.PhotoURL,
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
