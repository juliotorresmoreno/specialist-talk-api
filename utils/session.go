package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
)

type User struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	Phone     string `json:"phone"`
}

type Session struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

func ValidateSession(c *gin.Context) (*User, error) {
	token, err := GetToken(c)
	if err != nil {
		return nil, StatusUnauthorized
	}
	cmd := db.DefaultCache.Get(context.Background(), "session-"+token)
	email := cmd.Val()
	if email == "" {
		return nil, StatusUnauthorized
	}

	conn := db.DefaultClient
	user := &User{}
	err = conn.Model(&models.User{}).First(user, "email = ? AND deleted_at IS NULL", email).Error
	if err != nil {
		return nil, StatusInternalServerError
	}

	db.DefaultCache.Set(context.Background(), "session-"+token, email, 24*time.Hour)

	return user, nil
}

func ParseSession(token string, user *models.User) *Session {
	return &Session{
		Token: token,
		User: &User{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			PhotoURL:  user.PhotoURL,
			Phone:     user.Phone,
		},
	}
}

func MakeSession(user *models.User) (*Session, error) {
	token, err := GenerateRandomString(128)
	if err != nil {
		return &Session{}, StatusInternalServerError
	}

	cmd := db.DefaultCache.Set(
		context.Background(),
		"session-"+token,
		user.Email, 24*time.Hour,
	)
	if cmd.Err() != nil {
		return &Session{}, StatusInternalServerError
	}

	return ParseSession(token, user), nil
}
