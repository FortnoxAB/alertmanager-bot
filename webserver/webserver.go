package webserver

import (
	"net/http"

	"github.com/fortnoxab/ginprometheus"
	"github.com/gin-gonic/gin"
	"github.com/jonaz/ginlogrus"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type webserver struct {
	Slack      *slack.Client
	Prometheus *ginprometheus.Prometheus
}

//New webserver
func New(s *slack.Client) *webserver {
	return &webserver{
		Slack: s,
	}
}

//Init a webserver with Gin
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

	channelID, timestamp, err := ws.Slack.PostMessage(msg.Channel, slack.MsgOptionUsername(msg.Username), slack.MsgOptionAttachments(msg.Attachments...))
	if err != nil {
		return errors.Wrapf(err, "Error sending to channel %s", msg.Channel)
	}
	logrus.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

func (ws *webserver) healthHandler(c *gin.Context) {

	c.String(http.StatusOK, "OK")
}

func checkErr(f func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := f(c)
		if err != nil {
			logrus.Error(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
	}
}
