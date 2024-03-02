package events

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
)

var log = logger.SetupLogger()

type EventsRouter struct {
}

type Subscription struct {
	UserId uint
	Bus    chan interface{}
}

type manager struct {
	Event       chan *models.Event
	Subscribe   chan *Subscription
	Unsubscribe chan *Subscription

	subscribers map[uint][]chan interface{}
}

func (m *manager) Run() {
	defer close(m.Event)
	defer close(m.Subscribe)
	for {
		select {
		case e := <-m.Event:
			c := m.subscribers[e.UserId]
			for _, s := range c {
				go func(s chan interface{}) {
					s <- e.Payload
				}(s)
			}
		case s := <-m.Subscribe:
			m.subscribers[s.UserId] = append(m.subscribers[s.UserId], s.Bus)
		case s := <-m.Unsubscribe:
			c := m.subscribers[s.UserId]
			for i, b := range c {
				if b == s.Bus {
					c = append(c[:i], c[i+1:]...)
					break
				}
			}
			m.subscribers[s.UserId] = c
		}
	}
}

var Manager = &manager{
	Event:       make(chan *models.Event),
	Subscribe:   make(chan *Subscription),
	Unsubscribe: make(chan *Subscription),
	subscribers: make(map[uint][]chan interface{}),
}

func SetupAPIRoutes(r *gin.RouterGroup) {
	go Manager.Run()
	events := &EventsRouter{}
	r.GET("", events.subscribe)
	r.POST("/:id", events.publish)
}

func (h *EventsRouter) publish(c *gin.Context) {
	token, err := utils.GetToken(c)
	if err != nil {
		log.Error("Error getting token", err)
		utils.Response(c, err)
		return
	}
	if token != os.Getenv("API_KEY") {
		utils.Response(c, utils.StatusUnauthorized)
		return
	}

	userId, _ := strconv.Atoi(c.Param("id"))
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Error reading body", err)
		utils.Response(c, err)
		return
	}

	evt := &models.Event{
		UserId:  uint(userId),
		Payload: string(payload),
	}
	b, _ := json.Marshal(evt)
	db.DefaultCache.Publish(context.Background(), "events", string(b))

	c.String(http.StatusNoContent, "")
}

func (h *EventsRouter) subscribe(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		log.Error("Error validating session", err)
		utils.Response(c, err)
		return
	}

	bus := make(chan interface{})
	Manager.Subscribe <- &Subscription{
		UserId: session.ID,
		Bus:    bus,
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	for {
		select {
		case <-c.Request.Context().Done():
			Manager.Unsubscribe <- &Subscription{
				UserId: session.ID,
				Bus:    bus,
			}
			return
		case e := <-bus:
			c.SSEvent("message", e)
			c.Writer.Flush()
		}
	}
}
