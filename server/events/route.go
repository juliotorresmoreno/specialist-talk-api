package events

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
)

var log = logger.SetupLogger()

type Request struct {
	ID    uint
	Event *Event
}

type Event struct {
	Type string
	Data string
}

type Client struct {
	ID      uint
	Done    chan struct{}
	Handler chan *Event
}

type EventsRouter struct {
	Register    chan *Client
	Unregister  chan *Client
	Subscribers map[uint][]*Client
	Handler     chan *Request
}

var DefaultEventsRouter = &EventsRouter{
	Register:    make(chan *Client),
	Unregister:  make(chan *Client),
	Subscribers: make(map[uint][]*Client),
	Handler:     make(chan *Request),
}

func init() {
	go DefaultEventsRouter.Run()
}

func SetupAPIRoutes(g *gin.RouterGroup) chan *Request {
	h := DefaultEventsRouter

	g.GET("", h.Subscribe)
	g.POST("/:id", h.Publish)

	return h.Handler
}

func (h *EventsRouter) Run() {
	defer close(h.Register)
	defer close(h.Unregister)
	defer close(h.Handler)
	for {
		select {
		case client := <-h.Register:
			if _, ok := h.Subscribers[client.ID]; !ok {
				h.Subscribers[client.ID] = []*Client{}
			}
			h.Subscribers[client.ID] = append(h.Subscribers[client.ID], client)
		case client := <-h.Unregister:
			subscription, ok := h.Subscribers[client.ID]
			if !ok {
				continue
			}
			for i, c := range subscription {
				if c == client {
					h.Subscribers[client.ID] = append(subscription[:i], subscription[i+1:]...)
					break
				}
			}
		case request := <-h.Handler:
			subscription, ok := h.Subscribers[request.ID]
			if !ok {
				continue
			}
			for _, client := range subscription {
				client.Handler <- request.Event
			}
		}
	}
}

func (h *EventsRouter) Subscribe(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	ch := make(chan *Event)
	done := make(chan struct{})
	client := &Client{
		ID:      session.ID,
		Handler: ch,
		Done:    done,
	}

	defer func() {
		h.Unregister <- client
		close(ch)
		close(done)
		log.Info("Unregistering client")
	}()

	h.Register <- client
	c.Header("Content-Type", "text/event-stream")
	c.Status(200)

	c.SSEvent("connected", "Connected")
	c.Writer.Flush()

	for {
		select {
		case <-done:
			return
		case <-c.Writer.CloseNotify():
			return
		case event := <-ch:
			c.SSEvent(event.Type, event.Data)
			c.Writer.Flush()
		}
	}
}

func (h *EventsRouter) Publish(c *gin.Context) {
	_, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	event := &Event{}
	if err := c.ShouldBind(event); err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	h.Handler <- &Request{
		ID:    uint(id),
		Event: event,
	}

	c.JSON(200, gin.H{"status": "ok"})
}
