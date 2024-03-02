package auth

import (
	"net/http"
	"time"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
)

type AuthRouter struct {
}

func SetupAUTHRoutes(r *gin.RouterGroup) {
	auth := &AuthRouter{}

	r.GET("", auth.Ping)
	r.GET("/session", auth.Session)
	r.POST("/sign-in", auth.SignIn)
	r.POST("/sign-up", auth.SignUp)
}

type SignUpPayload struct {
	Name     string `json:"name" validate:"required,validname"`
	LastName string `json:"last_name" validate:"required,validname"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone" validate:"max=15"`
}

var signUpValidator = NewSignUpValidator()

func (auth *AuthRouter) SignUp(c *gin.Context) {
	payload := &SignUpPayload{}

	err := c.ShouldBind(payload)
	if err != nil {
		utils.Response(c, utils.StatusBadRequest)
		return
	}

	validation, err := signUpValidator.ValidateSignUp(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, validation)
		return
	}

	payload.Password, err = utils.HashPassword(payload.Password)
	if err != nil {
		utils.Response(c, utils.StatusInternalServerError)
		return
	}

	conn := db.DefaultClient

	user := &models.User{
		Name:     payload.Name,
		LastName: payload.LastName,
		Phone:    payload.Phone,
		Email:    payload.Email,
		Password: payload.Password,
	}
	tx := conn.Save(user)
	if tx.Error != nil {
		log.Info(tx.Error)

		if strings.Contains(tx.Error.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, gin.H{
				"email": payload.Email + " already exists",
			})
			return
		}
		utils.Response(c, utils.StatusInternalServerError)
		return
	}

	session, err := utils.MakeSession(user)
	if err != nil {
		utils.Response(c, err)
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    session.Token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, cookie)

	c.JSON(200, session.User)
}

type SignInPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (auth *AuthRouter) SignIn(c *gin.Context) {
	payload := &SignInPayload{}

	err := c.ShouldBind(payload)
	if err != nil {
		utils.Response(c, utils.StatusBadRequest)
		return
	}

	conn := db.DefaultClient
	user := &models.User{}

	tx := conn.Select(utils.SessionFields, "password").First(
		user, "email = ?", payload.Email,
	)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "User or password incorrect",
		})
		return
	}

	ok, err := utils.ComparePassword(payload.Password, user.Password)
	if !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "User or password incorrect",
		})
		return
	}
	user.Password = ""

	session, err := utils.MakeSession(user)
	if err != nil {
		utils.Response(c, err)
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    session.Token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, cookie)

	c.JSON(200, session.User)
}

func (auth *AuthRouter) Session(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		utils.Response(c, err)
		return
	}
	c.JSON(200, session)
}

func (auth *AuthRouter) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ok",
	})
}
