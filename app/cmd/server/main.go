package main

import (
	"flag"
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/caching"
	"github.com/ricdeau/gitlab-extension/app/pkg/config"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/handlers"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/pkg/telegram"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/olahol/melody.v1"
)

const (
	timestampFormat       = "02-01-2006 15:04:05.999 -0700"
	defaultConfigFilePath = "config.yaml"
	configFileFlagUsage   = "Configuration file path"
)

// topic names
const (
	UpdateCacheTopic = "cache"
	SocketTopic      = "ws"
	BotTopic         = "telegram_bot"
)

func main() {
	configFile := flag.String("config", defaultConfigFilePath, configFileFlagUsage)
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: timestampFormat,
		PrettyPrint:     false,
	})
	logger.SetOutput(os.Stdout)
	conf := config.Get(*configFile, logger)

	router := gin.New()
	msgBroker := broker.New()
	cache := caching.New(1 * time.Hour)

	setRouter(router, conf, logger)
	setCache(cache, msgBroker, logger)
	setTelegramBot(conf, logger, msgBroker)

	//set html handler
	router.Use(static.Serve("/", static.LocalFile("./www", true)))
	router.GET("/projects", handlers.NewProxy(conf, cache, logger).Handler())
	router.GET("/ws", handlers.NewSocket(SocketTopic, melody.New(), msgBroker, logger).Handler())
	router.POST("/webhook", handlers.NewWebhook(msgBroker, SocketTopic, UpdateCacheTopic, BotTopic).Handler())

	err := router.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logger.Fatalf("Unable to start server: %v", err)
	}
}

func setTelegramBot(conf *config.Config, logger *logrus.Logger, broker broker.MessageBroker) {
	db, err := telegram.NewBotDb()
	if err != nil {
		logger.Errorf("Unable to create bot db: %v", err)
		return
	}
	bot, err := telegram.NewBot(BotTopic, conf, db, broker, logger)
	if err != nil {
		logger.Errorf("Unable to authorize to telegram bot API: %v", err)
		return
	}
	err = bot.Start()
	if err != nil {
		logger.Errorf("Unable to start telegram bot: %v", err)
		return
	}
}

func setRouter(router *gin.Engine, conf *config.Config, logger *logrus.Logger) {
	router.Use(logging.Middleware(logger))
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowOrigins:     conf.Origins,
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
	}))
}

func setCache(cache caching.ProjectsCache, broker broker.MessageBroker, logger *logrus.Logger) {
	if err := broker.AddTopic(UpdateCacheTopic); err != nil {

	}
	err := broker.Subscribe(UpdateCacheTopic, func(message interface{}) {
		push, ok := message.(contracts.PipelinePush)
		if !ok {
			logger.Errorf("Invalid message type: %T", message)
		}
		err := cache.UpdatePipeline(push)
		if err != nil {
			logger.Errorf("ErrorResponse while updating cache: %v", err)
		}
	})
	if err != nil {
		logger.Fatalf("Set cache error: %v", err)
	}
}
