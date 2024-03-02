package auth

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type OauthRouter struct {
	key           string
	maxAge        int
	isProd        bool
	providerIndex *ProviderIndex
}

func SetupOAUTHRoutes(r *gin.RouterGroup) {
	auth := &OauthRouter{
		key:    "",
		maxAge: 86400 * 30,
		isProd: false,
	}

	store := sessions.NewFilesystemStore("/tmp", []byte(auth.key))
	store.MaxAge(auth.maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = auth.isProd

	gothic.Store = store

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_KEY"),
			os.Getenv("GOOGLE_SECRET"),
			os.Getenv("CALLBACK_BASE_URL")+"/google/callback",
		),
	)

	m := map[string]string{"google": "Google"}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}
	auth.providerIndex = providerIndex

	r.GET("/:provider/callback", auth.AuthCallback)
	r.GET("/:provider/logout", auth.Logout)
	r.GET("/:provider", auth.AuthHandler)
	r.GET("", auth.Ping)
}

func (auth *OauthRouter) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func (p *OauthRouter) Logout(c *gin.Context) {
	gothic.Logout(c.Writer, c.Request)
	c.Header("Location", "/")
	c.Status(http.StatusTemporaryRedirect)
}

func (p *OauthRouter) AuthHandler(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	if guser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		complete(c, guser)
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

func (auth *OauthRouter) AuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	guser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	complete(c, guser)
}

func complete(c *gin.Context, guser goth.User) {
	conn := db.DefaultClient
	user := &models.User{}
	tx := conn.Find(user, &models.User{Email: guser.Email})
	if tx.Error != nil {
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
	c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_BASE_URL"))
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}
