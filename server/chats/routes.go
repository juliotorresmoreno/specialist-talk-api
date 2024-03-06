package chats

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
	"gorm.io/gorm"
)

type ChatsRouter struct{}

func SetupAPIRoutes(g *gin.RouterGroup) {
	h := &ChatsRouter{}

	g.GET("", h.find)
	g.GET("/:id", h.findOne)
	g.POST("", h.create)
	g.PATCH("/:id", h.update)
	g.PUT("/addUser", h.addUser)
	g.GET("/:id/users", h.getUsers)
}

type User struct {
	ID           uint       `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Username     string     `json:"username"`
	PhotoURL     string     `json:"photo_url"`
	Business     string     `json:"business"`
	PositionName string     `json:"position_name"`
	Chats        []Chat     `json:"-" gorm:"foreignKey:OwnerId"`
	ChatUsers    []ChatUser `json:"-" gorm:"foreignKey:UserId"`
	CreationAt   time.Time  `json:"creation_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type Chat struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	OwnerId    uint       `json:"owner_id"`
	Owner      *User      `json:"owner"`
	Chats      []ChatUser `json:"-"`
	CreationAt time.Time  `json:"creation_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type ChatUser struct {
	ID         uint       `json:"id"`
	ChatId     uint       `json:"chat_id"`
	Chat       Chat       `json:"-"`
	UserId     uint       `json:"user_id"`
	User       User       `json:"user"`
	CreationAt time.Time  `json:"creation_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (h *ChatsRouter) find(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	chatUsers := []*ChatUser{}
	err = db.DefaultClient.Model(&models.ChatUser{}).
		Preload("Chat", "deleted_at is null", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Owner", "deleted_at is null")
		}).
		Where(&models.ChatUser{UserId: session.ID}).
		Find(&chatUsers).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	chats := []Chat{}
	for _, chatUser := range chatUsers {
		chats = append(chats, chatUser.Chat)
	}

	c.JSON(200, chats)
}

type CreateErrors struct {
	Name string `json:"name,omitempty"`
}

func (h *ChatsRouter) create(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	payload := &Chat{}
	err = c.ShouldBind(payload)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		errorsMap := utils.ParseErrors(err.(validator.ValidationErrors))
		customErrors := CreateErrors{
			Name: errorsMap["Name"],
		}
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	chat := &models.Chat{
		Name:    payload.Name,
		OwnerId: session.ID,
	}
	err = db.DefaultClient.Create(chat).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	chatUser := &models.ChatUser{
		ChatId: chat.ID,
		UserId: session.ID,
	}
	err = db.DefaultClient.Create(chatUser).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Created"})
}

func (h *ChatsRouter) update(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	payload := &Chat{}
	err = c.ShouldBind(payload)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		errorsMap := utils.ParseErrors(err.(validator.ValidationErrors))
		customErrors := CreateErrors{
			Name: errorsMap["Name"],
		}
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	chat := &models.Chat{
		Name: payload.Name,
	}
	err = db.DefaultClient.Model(&models.Chat{}).
		Where(&models.Chat{ID: uint(id), OwnerId: session.ID}).
		Updates(chat).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Updated"})
}

type AddUserPayload struct {
	ChatID uint `json:"chat_id" validate:"required"`
	UserID uint `json:"user_id" validate:"required"`
}

type AddUserErrors struct {
	ChatID string `json:"chat_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

func (h *ChatsRouter) addUser(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	chat := &Chat{}
	err = db.DefaultClient.Model(&models.Chat{}).
		Where(&models.Chat{ID: chat.ID}).
		First(chat).Error
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if chat.OwnerId != session.ID {
		chatUser := &ChatUser{}
		err = db.DefaultClient.Model(&models.ChatUser{}).
			Where(&models.ChatUser{ChatId: chat.ID, UserId: session.ID}).
			First(chatUser).Error
		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
	}

	payload := &AddUserPayload{}
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		errorsMap := utils.ParseErrors(err.(validator.ValidationErrors))
		customErrors := AddUserErrors{
			ChatID: errorsMap["ChatID"],
			UserID: errorsMap["UserID"],
		}
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	chatUser := &models.ChatUser{
		ChatId: payload.ChatID,
		UserId: payload.UserID,
	}
	err = db.DefaultClient.Create(chatUser).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, gin.H{"message": "Created"})
}

func (h *ChatsRouter) findOne(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	chat := &Chat{}
	err = db.DefaultClient.Model(&models.Chat{}).
		Preload("Owner", "deleted_at is null").
		Where(&models.Chat{ID: uint(id)}).
		First(chat).Error
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if chat.OwnerId != session.ID {
		chatUser := &ChatUser{}
		err = db.DefaultClient.Model(&models.ChatUser{}).
			Where(&models.ChatUser{ChatId: chat.ID, UserId: session.ID}).
			First(chatUser).Error
		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
	}

	c.JSON(200, chat)
}

func (h *ChatsRouter) getUsers(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	chat := &Chat{}
	err = db.DefaultClient.Model(&models.Chat{}).
		Where(&models.Chat{ID: uint(id)}).
		First(chat).Error
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if chat.OwnerId != session.ID {
		chatUser := &ChatUser{}
		err = db.DefaultClient.Model(&models.ChatUser{}).
			Where(&models.ChatUser{ChatId: chat.ID, UserId: session.ID}).
			First(chatUser).Error
		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
	}

	users := []*ChatUser{}
	err = db.DefaultClient.Model(&models.ChatUser{}).
		Where(&models.ChatUser{ChatId: chat.ID}).
		Preload("User", "deleted_at is null").
		Find(&users).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, users)
}
