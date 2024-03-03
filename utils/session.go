package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
)

var SessionFields = []string{"id", "first_name", "last_name", "username", "email", "photo_url", "phone"}

type User struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
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
		return &User{}, StatusUnauthorized
	}
	ctx := context.Background()
	cmd := db.DefaultCache.Get(ctx, "session-"+token)
	email := cmd.Val()
	if email == "" {
		return &User{}, StatusUnauthorized
	}

	conn := db.DefaultClient
	user := &models.User{}
	tx := conn.Select(SessionFields).First(user, "email = ? AND deleted_at IS NULL", email)
	if tx.Error != nil {
		return &User{}, StatusInternalServerError
	}

	db.DefaultCache.Set(ctx, "session-"+token, email, 24*time.Hour)
	session := ParseSession(token, user)

	return session.User, nil
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
