package main

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prologic/bitcask"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/caching"
	"github.com/ricdeau/gitlab-extension/app/pkg/config"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/handlers"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/pkg/telegram"
	"io"
	"io/ioutil"
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

const (
	UpdateCacheTopic = "cache"
)

func main() {
	configFile := flag.String("config", defaultConfigFilePath, configFileFlagUsage)
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: timestampFormat,
		PrettyPrint:     false,
	})
	conf := config.Get(*configFile, logger)

	router := gin.New()
	msgBroker := broker.New()
	cache := caching.New(1 * time.Hour)

	setLogger(conf, logger)
	setRouter(router, logger)
	setCache(cache, msgBroker, logger)

	// set websocket handler
	wsHandler := melody.New()

	// set embedded db
	db, err := bitcask.Open("db")
	defer func() {
		err := db.Close()
		logger.Errorf("Error while closing db: %v", err)
	}()

	// set telegram bot
	setTelegramBot(conf, logger, db, msgBroker)

	//set html handler
	router.Use(static.Serve("/", static.LocalFile("./www", true)))

	// set proxy handler
	proxyHandler := handlers.NewProxyHandler(conf, cache, logger)
	router.GET("/projects", proxyHandler.Handle)

	// set socket handler
	socketHandler := handlers.NewSocketHandler(wsHandler, msgBroker, logger)
	router.GET("/ws", socketHandler.Handle)

	// set webhook handler
	topics := []string{handlers.SocketTopic, UpdateCacheTopic}
	if conf.BotEnabled {
		topics = append(topics, telegram.BotTopic)
	}
	webhookHandler := handlers.NewWebhook(msgBroker, topics...)
	router.POST("/webhook", webhookHandler.CreateHandler())

	err = router.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logger.Fatalf("Unable to start boot: %v", err)
	}
}

func setTelegramBot(conf *config.Config, logger *logrus.Logger, db *bitcask.Bitcask, broker broker.MessageBroker) {
	if conf.BotEnabled {
		botApi, err := tgbotapi.NewBotAPI(conf.BotToken)
		if err != nil {
			logger.Fatalf("Unable to authorize to telegram bot API: %v", err)
		}
		bot := telegram.NewBot(botApi, db, broker, conf, logger)
		bot.Start()
	}
}

func setRouter(router *gin.Engine, logger *logrus.Logger) {
	router.Use(logging.Middleware(logger))
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowOrigins:     []string{"http://localhost*", "http://devservice.tech*"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
	}))
}

func setLogger(config *config.Config, logger *logrus.Logger) {
	writers := make([]io.Writer, 0)
	if config.HasConsoleLogging() {
		writers = append(writers, os.Stdout)
	}
	if config.HasFileLogging() && config.RollingFileSettings != nil {
		fileLogger := config.RollingFileSettings.CreateRollingWriter(logger)
		writers = append(writers, fileLogger)
	}
	if len(writers) == 0 {
		logger.SetOutput(ioutil.Discard)
	} else {
		logger.SetOutput(io.MultiWriter(writers...))
	}
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
			logger.Errorf("Error while updating cache: %v", err)
		}
	})
	if err != nil {
		logger.Fatalf("Set cache error: %v", err)
	}
}
