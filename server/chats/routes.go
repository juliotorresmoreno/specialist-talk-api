package chats

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/server/events"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
	"gorm.io/gorm"
)

var log = logger.SetupLogger()

type ChatsRouter struct{}

func SetupAPIRoutes(g *gin.RouterGroup) {
	h := &ChatsRouter{}

	g.GET("", h.find)
	g.GET("/:id", h.findOne)
	g.POST("", h.create)
	g.PATCH("/:id", h.update)
	g.PUT("/:id/addUser", h.addUser)
	g.GET("/:id/users", h.getUsers)
	g.GET("/:id/messages", h.getMessages)
	g.POST("/:id/messages", h.createMessage)
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
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	PhotoURL    string     `json:"photo_url"`
	OwnerId     uint       `json:"owner_id"`
	Owner       *User      `json:"owner"`
	ChatUsers   []ChatUser `json:"chat_users,omitempty"`
	UserId      uint       `json:"user_id,omitempty" gorm:"-"`
	Code        string     `json:"code"`
	Active      bool       `json:"active"`
	CreationAt  time.Time  `json:"creation_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
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
			return db.Preload("Owner", "deleted_at is null").
				Preload("ChatUsers", "deleted_at is null").
				Preload("ChatUsers.User", "deleted_at is null")
		}).
		Where(&models.ChatUser{UserId: session.ID}).
		Find(&chatUsers).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	chats := []Chat{}
	for _, chatUser := range chatUsers {
		if chatUser.Chat.Active {
			json.NewEncoder(os.Stdout).Encode(chatUser.Chat)
			chats = append(chats, chatUser.Chat)
		}
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

	code := ""
	exists := models.Chat{}
	if payload.UserId != 0 {
		list := []uint{session.ID, payload.UserId}
		sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })

		jsonData, err := json.Marshal(list)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		hash := sha1.Sum(jsonData)
		code = fmt.Sprintf("%x", hash)

		db.DefaultClient.
			Model(&models.Chat{}).
			Preload("Owner", "deleted_at is null").
			Preload("Chats", "deleted_at is null", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User", "deleted_at is null")
			}).
			Where(&Chat{Code: code}).First(&exists)
		if exists.ID > 0 {
			if exists.OwnerId != session.ID {
				db.DefaultClient.Where(&models.Chat{ID: exists.ID}).
					Updates(&models.Chat{
						Active: true,
					})
				chatUser := &models.ChatUser{}
				db.DefaultClient.Where(&models.ChatUser{
					ChatId: exists.ID,
					UserId: session.ID,
				}).First(chatUser)
				if chatUser.ID == 0 {
					chatUser = &models.ChatUser{
						ChatId: exists.ID,
						UserId: session.ID,
					}
					err = db.DefaultClient.Create(chatUser).Error
					if err != nil {
						c.JSON(500, gin.H{"error": err.Error()})
						return
					}
				}
			}
			c.JSON(200, Chat{
				ID:          exists.ID,
				Name:        exists.Name,
				Description: exists.Description,
				PhotoURL:    exists.PhotoURL,
				Code:        exists.Code,
				Active:      exists.Active,
				OwnerId:     exists.OwnerId,
				Owner: &User{
					ID:           exists.Owner.ID,
					FirstName:    exists.Owner.FirstName,
					LastName:     exists.Owner.LastName,
					Username:     exists.Owner.Username,
					PhotoURL:     exists.Owner.PhotoURL,
					Business:     exists.Owner.Business,
					PositionName: exists.Owner.PositionName,
					CreationAt:   exists.Owner.CreationAt,
					UpdatedAt:    exists.Owner.UpdatedAt,
				},
				ChatUsers:  parseChatUsers(exists.Chats),
				CreationAt: exists.CreationAt,
				UpdatedAt:  exists.UpdatedAt,
			})
			return
		}
	}

	chat := &models.Chat{
		Name:        payload.Name,
		Description: payload.Description,
		PhotoURL:    payload.PhotoURL,
		Code:        code,
		Active:      false,
		OwnerId:     session.ID,
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

	exists = models.Chat{}
	err = db.DefaultClient.
		Preload("Owner", "deleted_at is null").
		Preload("Chats", "deleted_at is null", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User", "deleted_at is null")
		}).
		Where(&Chat{ID: chat.ID}).First(&exists).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, Chat{
		ID:          exists.ID,
		Name:        exists.Name,
		Description: exists.Description,
		PhotoURL:    exists.PhotoURL,
		Code:        exists.Code,
		Active:      exists.Active,
		OwnerId:     exists.OwnerId,
		Owner: &User{
			ID:           exists.Owner.ID,
			FirstName:    exists.Owner.FirstName,
			LastName:     exists.Owner.LastName,
			Username:     exists.Owner.Username,
			PhotoURL:     exists.Owner.PhotoURL,
			Business:     exists.Owner.Business,
			PositionName: exists.Owner.PositionName,
			CreationAt:   exists.Owner.CreationAt,
			UpdatedAt:    exists.Owner.UpdatedAt,
		},
		ChatUsers:  parseChatUsers(exists.Chats),
		CreationAt: exists.CreationAt,
		UpdatedAt:  exists.UpdatedAt,
	})
}

func parseChatUsers(chatUsers []models.ChatUser) []ChatUser {
	result := []ChatUser{}
	for _, chatUser := range chatUsers {
		result = append(result, ChatUser{
			ID:     chatUser.ID,
			ChatId: chatUser.ChatId,
			UserId: chatUser.UserId,
			User: User{
				ID:           chatUser.User.ID,
				FirstName:    chatUser.User.FirstName,
				LastName:     chatUser.User.LastName,
				Username:     chatUser.User.Username,
				PhotoURL:     chatUser.User.PhotoURL,
				Business:     chatUser.User.Business,
				PositionName: chatUser.User.PositionName,
				CreationAt:   chatUser.User.CreationAt,
				UpdatedAt:    chatUser.User.UpdatedAt,
			},
			CreationAt: chatUser.CreationAt,
			UpdatedAt:  chatUser.UpdatedAt,
		})
	}
	return result
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

	if ok, _ := h.memberOfChat(uint(id), session.ID); !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	chat := &models.Chat{
		Name:        payload.Name,
		Description: payload.Description,
		PhotoURL:    payload.PhotoURL,
	}
	err = db.DefaultClient.Model(&models.Chat{}).
		Where(&models.Chat{ID: uint(id)}).
		Updates(chat).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Updated"})
}

type AddUserPayload struct {
	UserID uint `json:"user_id" validate:"required"`
}

type AddUserErrors struct {
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

	id, _ := strconv.Atoi(c.Param("id"))
	ok, chat := h.memberOfChat(uint(id), session.ID)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	if chat.Code != "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	payload := &AddUserPayload{}
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		errorsMap := utils.ParseErrors(err.(validator.ValidationErrors))
		customErrors := AddUserErrors{
			UserID: errorsMap["UserID"],
		}
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	chatUser := &models.ChatUser{
		ChatId: uint(id),
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
	ok, chat := h.memberOfChat(uint(id), session.ID)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
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
	if ok, _ := h.memberOfChat(uint(id), session.ID); !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	users := []*ChatUser{}
	err = db.DefaultClient.Model(&models.ChatUser{}).
		Where(&models.ChatUser{ChatId: uint(id)}).
		Preload("User", "deleted_at is null").
		Find(&users).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, users)
}

type Message struct {
	ID         uint       `json:"id"`
	ChatID     uint       `json:"chat_id"`
	Chat       Chat       `json:"-"`
	UserID     uint       `json:"user_id"`
	User       User       `json:"user"`
	Content    string     `json:"content"`
	CreationAt time.Time  `json:"creation_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (h *ChatsRouter) getMessages(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if ok, _ := h.memberOfChat(uint(id), session.ID); !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	messages := []*Message{}
	err = db.DefaultClient.Model(&models.Message{}).
		Where(&models.Message{ChatId: uint(id)}).
		Preload("User", "deleted_at is null").
		Find(&messages).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, messages)
}

type CreateMessagePayload struct {
	Content string `json:"content" validate:"required"`
}

type CreateMessageErrors struct {
	Content string `json:"content,omitempty"`
}

func (h *ChatsRouter) createMessage(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		log.Error("Unauthorized: ", err)
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	payload := &CreateMessagePayload{}
	err = c.ShouldBind(payload)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		errorsMap := utils.ParseErrors(err.(validator.ValidationErrors))
		customErrors := CreateMessageErrors{
			Content: errorsMap["Content"],
		}
		c.JSON(http.StatusBadRequest, customErrors)
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	ok, chat := h.memberOfChat(uint(id), session.ID)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	if !chat.Active {
		err := db.DefaultClient.Model(&models.Chat{}).
			Where(&models.Chat{ID: uint(id)}).
			Updates(&models.Chat{Active: true}).Error
		if err != nil {
			log.Error("Error updating chat: ", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
	}

	message := &models.Message{
		ChatId:  uint(id),
		UserId:  session.ID,
		Content: payload.Content,
	}
	err = db.DefaultClient.Create(message).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	h.sendToChat(uint(id), Message{
		ID:     message.ID,
		ChatID: message.ChatId,
		UserID: message.UserId,
		User: User{
			ID:        session.ID,
			FirstName: session.FirstName,
			LastName:  session.LastName,
			Username:  session.Username,
			PhotoURL:  session.PhotoURL,
		},
		Content: message.Content,
	})

	c.JSON(200, gin.H{"message": "Created"})
}

func (h *ChatsRouter) sendToChat(chatID uint, message Message) error {
	chatUsers := []*ChatUser{}
	err := db.DefaultClient.Model(&models.ChatUser{}).
		Where(&models.ChatUser{ChatId: chatID}).
		Find(&chatUsers).Error
	if err != nil {
		return err
	}
	data, _ := json.Marshal(message)
	for _, chatUser := range chatUsers {
		events.DefaultEventsRouter.Handler <- &events.Request{
			ID: chatUser.UserId,
			Event: &events.Event{
				Type: "message",
				Data: string(data),
			},
		}
	}
	return nil
}

func (h *ChatsRouter) memberOfChat(chatID, userID uint) (bool, *Chat) {
	chat := &Chat{}
	err := db.DefaultClient.Model(&models.Chat{}).
		Where(&models.Chat{ID: uint(chatID)}).
		First(chat).Error
	if err != nil {
		return false, nil
	}

	if chat.OwnerId != userID {
		chatUser := &ChatUser{}
		err = db.DefaultClient.Model(&models.ChatUser{}).
			Where(&models.ChatUser{ChatId: chat.ID, UserId: userID}).
			First(chatUser).Error
		if err != nil {
			return false, nil
		}
	}

	return true, chat
}
