package webserver

import (
	"fmt"
	"net/http"

	"github.com/fortnoxab/ginprometheus"
	"github.com/gin-gonic/gin"
	"github.com/jonaz/ginlogrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var errorsSending = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "alertmanager_bot_errors",
	Help: "Number of errors posting message to channels.",
})

func init() {
	prometheus.MustRegister(errorsSending)
}

type webserver struct {
	Slack      *slack.Client
	Prometheus *ginprometheus.Prometheus
}

// New webserver
func New(s *slack.Client) *webserver {
	return &webserver{
		Slack: s,
	}
}

// Init a webserver with Gin
func (ws *webserver) Init() *gin.Engine {

	router := gin.New()

	if ws.Prometheus != nil {
		ws.Prometheus.Use(router)
	}

	router.Use(ginlogrus.New(logrus.StandardLogger(), "/health", "/metrics"), gin.Recovery())

	router.POST("/webhook/:channel", checkErr(ws.handleWebhook))
	router.POST("/webhook", checkErr(ws.handleWebhook))

	router.GET("/health", ws.healthHandler)
	return router
}

func (ws *webserver) handleWebhook(c *gin.Context) error {
	logrus.Info(c.Request.Header)

	type WebhookMessage struct {
		// Copied from https://github.com/prometheus/alertmanager/blob/96fce3e8abd48abe17f06202d17bfaf0490f13cb/notify/impl.go#L820
		Channel     string             `json:"channel,omitempty"`
		Username    string             `json:"username,omitempty"`
		IconEmoji   string             `json:"icon_emoji,omitempty"`
		IconURL     string             `json:"icon_url,omitempty"`
		LinkNames   bool               `json:"link_names,omitempty"`
		Text        string             `json:"text,omitempty"`
		Attachments []slack.Attachment `json:"attachments"`
	}
	msg := &WebhookMessage{}
	err := c.BindJSON(msg)
	if err != nil {
		return err
	}

	if c.Param("channel") != "" {
		msg.Channel = c.Param("channel")
	}

	msgoptions := []slack.MsgOption{}

	if msg.Text != "" {
		msgoptions = append(msgoptions, slack.MsgOptionText(msg.Text, true))
	}

	if msg.Username != "" {
		msgoptions = append(msgoptions, slack.MsgOptionUsername(msg.Username))
	}

	if len(msg.Attachments) > 0 {
		msgoptions = append(msgoptions, slack.MsgOptionAttachments(msg.Attachments...))
	}

	channelID, timestamp, err := ws.Slack.PostMessage(msg.Channel, msgoptions...)
	if err != nil {
		return fmt.Errorf("error sending to channel %s: %w", msg.Channel, err)
	}
	var w gin.ResponseWriter = c.Writer
	w.WriteString("ok")
	logrus.Infof("message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

func (ws *webserver) healthHandler(c *gin.Context) {

	c.String(http.StatusOK, "OK")
}

func checkErr(f func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := f(c)
		if err != nil {
			errorsSending.Inc()
			logrus.Error(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
	}
}
