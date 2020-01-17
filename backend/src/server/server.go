package main

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	externalCache "github.com/patrickmn/go-cache"
	"github.com/prologic/bitcask"
	"io"
	"io/ioutil"
	"os"
	cache2 "server/cache"
	"server/telegram"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/olahol/melody.v1"

	"server/conf"
	"server/handlers"
	"server/logging"
	"server/queue"
)

const (
	timestampFormat       = "02-01-2006 15:04:05.999 -0700"
	defaultConfigFilePath = "config.yaml"
	configFileFlagUsage   = "Configuration file path"
)

func main() {
	configFile := flag.String("config", defaultConfigFilePath, configFileFlagUsage)
	flag.Parse()

	// set logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: timestampFormat,
		PrettyPrint:     false,
	})

	// set config
	config := conf.Get(*configFile, logger)

	// sel logging targets
	setLogger(config, logger)

	// set router
	router := setRouter(logger)

	// set websocket handler
	wsHandler := melody.New()

	// set embedded db
	db, err := bitcask.Open("db")
	defer func() {
		err := db.Close()
		logger.Errorf("Error while closing db: %v", err)
	}()

	// set global queue
	globalQueue := queue.NewGlobalQueue(logger)

	// set cache
	cache := cache2.NewCache(externalCache.New(60*time.Minute, -1), globalQueue, logger)

	// set telegram bot
	setTelegramBot(config, logger, db, globalQueue)

	//set html handler
	router.Use(handlers.Serve("/", handlers.LocalFile("./www", true)))

	// set proxy handler
	proxyHandler := handlers.NewProxyHandler(config, cache, logger)
	router.GET("/projects", proxyHandler.Handle)

	// set socket handler
	socketHandler := handlers.NewSocketHandler(wsHandler, globalQueue, logger)
	router.GET("/ws", socketHandler.Handle)

	// set webhook handler
	topics := []string{handlers.SocketTopic, cache2.UpdateCacheTopic}
	if config.BotEnabled {
		topics = append(topics, telegram.BotTopic)
	}
	webhookHandler := handlers.NewWebhookHandler(globalQueue, topics, logger)
	router.POST("/webhook", webhookHandler.Handle)

	err = router.Run(fmt.Sprintf(":%d", config.Port))
	if err != nil {
		logger.Fatalf("Unable to start server: %v", err)
	}
}

func setTelegramBot(config *conf.Config, logger *logrus.Logger, db *bitcask.Bitcask, globalQueue *queue.GlobalQueue) {
	if config.BotEnabled {
		botApi, err := tgbotapi.NewBotAPI(config.BotToken)
		if err != nil {
			logger.Fatalf("Unable to authorize to telegram bot API: %v", err)
		}
		bot := telegram.NewBot(botApi, db, globalQueue, config, logger)
		bot.Start()
	}
}

func setRouter(logger *logrus.Logger) *gin.Engine {
	router := gin.New()
	router.Use(logging.Logger(logger))
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
		AllowOrigins:     []string{"http://localhost*", "http://devservice.tech*"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
	}))
	return router
}

func setLogger(config *conf.Config, logger *logrus.Logger) {
	writers := make([]io.Writer, 0)
	if config.HasConsoleLogging() {
		writers = append(writers, os.Stdout)
	}
	if config.HasFileLogging() && config.RollingFileSettings != nil {
		fileLogger := config.RollingFileSettings.CreateLumberjack(logger)
		writers = append(writers, fileLogger)
	}
	if len(writers) == 0 {
		logger.SetOutput(ioutil.Discard)
	} else {
		logger.SetOutput(io.MultiWriter(writers...))
	}
}
