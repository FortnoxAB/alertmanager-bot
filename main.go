package main

import (
	"context"
	"strings"
	"time"

	"github.com/fortnoxab/alertmanager-bot/config"
	"github.com/fortnoxab/alertmanager-bot/webserver"
	"github.com/fortnoxab/fnxlogrus"
	"github.com/fortnoxab/ginprometheus"
	"github.com/jonaz/gograce"
	"github.com/koding/multiconfig"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

func main() {
	config := &config.Config{}
	multiconfig.MustLoad(&config)

	fnxlogrus.Init(config.Log, logrus.StandardLogger())

	s := slack.New(config.Token)

	_, err := s.AuthTest()
	if err != nil {
		logrus.Error(err)
		return
	}

	ws := webserver.New(s)
	ws.Prometheus = ginprometheus.New("http")

	srv, shutdown := gograce.NewServerWithTimeout(5 * time.Second)
	srv.Handler = ws.Init()
	srv.Addr = ":" + config.Port

	ctx, cancel := context.WithCancel(context.Background())
	go rtmWorker(ctx, s)

	logrus.Error(srv.ListenAndServe())
	cancel()
	<-shutdown

}

func rtmWorker(ctx context.Context, s *slack.Client) {

	rtm := s.NewRTM()
	go rtm.ManageConnection()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("stopping rtmWorker")
			rtm.Disconnect()
			return
		case msg := <-rtm.IncomingEvents:
			processEvent(rtm, msg)

		}
	}

}

var slackInfo *slack.Info

func mentionedMe(msg string) bool {
	if slackInfo == nil || slackInfo.User == nil {
		return false
	}

	uid := slackInfo.User.ID
	if strings.Contains(msg, "<@"+uid+">") {
		logrus.Info("return true")
		return true
	}
	return false

}

func processEvent(rtm *slack.RTM, event slack.RTMEvent) {

	switch msg := event.Data.(type) {
	case *slack.HelloEvent:
		// Ignore hello

	case *slack.ConnectedEvent:
		logrus.Info("Connected info:", msg.Info)
		slackInfo = msg.Info

	case *slack.MessageEvent:
		logrus.Infof("Message: %s in/from: %s type: %s", msg.Text, msg.Channel, msg.SubType)
		// if direct message to the bot. Reply with help stuff

		helpMention := mentionedMe(msg.Text) && strings.Contains(msg.Text, "help")
		helpPrivate := msg.Text == "help" && strings.HasPrefix(msg.Channel, "D")
		if helpMention || helpPrivate {
			rtm.SendMessage(rtm.NewOutgoingMessage("By adding me to a channel you can receive alertmanager alerts.", msg.Channel))
		}

	case *slack.LatencyReport:
		logrus.Infof("Current latency: %v", msg.Value)

	case *slack.GroupJoinedEvent:
		logrus.Infof("Bot joined: %s", msg.Channel.Name)

	case *slack.GroupLeftEvent:
		logrus.Infof("Bot left: %s", msg.Channel)

	case *slack.RTMError:
		logrus.Errorf("Error: %s\n", msg.Error())

	case *slack.DisconnectedEvent:
		logrus.Error("Disconnected", msg)

	case *slack.InvalidAuthEvent:
		logrus.Error("Invalid credentials")
		return

	default:
	}
}
