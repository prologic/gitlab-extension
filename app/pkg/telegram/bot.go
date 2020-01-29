package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/config"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/pkg/utils"
	"strconv"
	"strings"
)

const (
	chatPrefix = "chat"
)

type GitlabMessage contracts.PipelinePush

// Telegram bot that forwards messages form global queue topic to telegram chats.
type Bot struct {
	*tgbotapi.BotAPI
	*config.Config
	topic     string
	db        BotDb
	queue     broker.MessageBroker
	logger    logging.Logger
	updatesCh tgbotapi.UpdatesChannel
}

// Creates new instance of telegram bot.
func NewBot(
	topic string,
	config *config.Config,
	db BotDb,
	queue broker.MessageBroker,
	logger logging.Logger) (*Bot, error) {

	botApi, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, err
	}
	bot := &Bot{}
	bot.topic = topic
	bot.BotAPI = botApi
	bot.db = db
	bot.queue = queue
	bot.Config = config
	bot.logger = logger
	updates, err := bot.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 5})
	if err != nil {
		return nil, err
	}
	bot.updatesCh = updates
	bot.logger.Infof("Telegram bot initialized")
	return bot, nil
}

// Start handling messages.
func (bot *Bot) Start() error {
	if err := bot.subscribeToTopic(); err != nil {
		return err
	}
	go func() {
		for update := range bot.updatesCh {
			chatId := update.Message.Chat.ID
			if update.Message == nil {
				continue
			}
			if update.Message.Text == "/start" {
				bot.Send(chatId, "Please provide your gitlab private token.")
				continue
			}
			availableNamespaces := bot.getAvailableNamespaces(update.Message.Text)
			if len(availableNamespaces) == 0 {
				bot.Send(chatId, "You don't have any available groups.")
				continue
			}

			err := bot.setChatNamespaces(chatId, availableNamespaces)
			if err != nil {
				bot.logger.Errorf("ErrorResponse while updating gitlab namespaces for chat id=%d", chatId)
				bot.Send(chatId, "Sorry something went wrong.")
			}
			availableNamespaces, err = bot.getChatNamespaces(chatId)
			if err != nil {
				bot.logger.Errorf("ErrorResponse while updating gitlab namespaces for chat id=%d", chatId)
				bot.Send(chatId, "Sorry something went wrong.")
			}
			namespacesString := strings.Join(availableNamespaces, ", ")
			bot.Send(chatId, fmt.Sprintf("You have been subscribed to group: %s", namespacesString))
		}
	}()
	return nil
}

// Send message to chat.
// chatId - identifier of telegram chat.
// text - message's text.
func (bot *Bot) Send(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	reply, err := bot.BotAPI.Send(msg)
	if err != nil {
		bot.logger.Errorf("ErrorResponse while sending message to chat id=%d", chatId)
	} else {
		bot.logger.Infof("Message has been sent to chat id=%d: %v", chatId, reply.Text)
	}
}

// Get namespaces that have been bound to given chat id.
func (bot *Bot) getChatNamespaces(chatId int64) (result []string, err error) {
	prefix := fmt.Sprintf("%s_%d_", chatPrefix, chatId)
	err = bot.db.Scan(prefix, func(key string) error {
		result = append(result, strings.TrimPrefix(key, prefix))
		return nil
	})
	return
}

// Binds namespaces to given chat id.
func (bot *Bot) setChatNamespaces(chatId int64, namespaces []string) (err error) {
	err = bot.db.Transaction(func() error {
		for _, ns := range namespaces {
			key := fmt.Sprintf("%s_%d_%s", chatPrefix, chatId, ns)
			if bot.db.Contains(key) {
				continue
			}
			return bot.db.Set(key)
		}
		return nil
	})
	return
}

// Check that gitlab namespaces are accessible with provided private token
// Namespaces matched by "name" field
// Returns slice of accessible namespaces
func (bot *Bot) getAvailableNamespaces(privateToken string) (result []string) {
	url := fmt.Sprintf("%s/api/v4/namespaces", bot.GitlabUri)
	headers := map[string]string{"Private-Token": privateToken}
	response, err := utils.PerformGetRequest(bot.Client, url, headers, bot.logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		bot.logger.Errorf("error while decoding ToJson body: %v", err)
		return
	}
	for _, el := range rawJson {
		for _, ns := range bot.GitlabNamespaces {
			if el["name"].(string) == ns {
				result = append(result, ns)
			}
		}
	}
	return result
}

// Subscribes bot to specific topic in global queue.
func (bot *Bot) subscribeToTopic() (err error) {
	if err = bot.queue.AddTopic(bot.topic); err != nil {
		panic(err)
	}
	return bot.queue.Subscribe(bot.topic, func(message interface{}) {
		msg := GitlabMessage(message.(contracts.PipelinePush))
		err := bot.db.Scan(chatPrefix, func(key string) error {
			if strings.HasSuffix(key, msg.Project.Namespace) {
				parts := strings.Split(key, "_")
				if len(parts) > 2 {
					chatId, err := strconv.ParseInt(parts[1], 10, 64)
					if err != nil {
						return err
					}
					bot.Send(chatId, msg.toTelegramMessageText())
				}
			}
			return nil
		})
		if err != nil {
			bot.logger.Errorf("ErrorResponse while sending gitlab update to telegram: %v", err)
		}
	})
}

// Formats gitlab message to telegram's message text.
func (msg *GitlabMessage) toTelegramMessageText() string {
	template := "Operation: %s\r\nStatus: %s\r\nNamespace: %s\r\nProject : %s\r\nBranch: %s\r\nCommit sha: %s\r\n" +
		"Commit message: %s\r\nUser: %s\r\nCreatedAt: %s\r\nFinishedAt: %s\r\nDuration: %d"

	return fmt.Sprintf(template,
		msg.Kind,
		msg.Attributes.Status,
		msg.Project.Namespace,
		msg.Project.Name,
		msg.Attributes.Branch,
		msg.Commit.Id,
		msg.Commit.Message,
		msg.User.Name,
		msg.Attributes.CreatedAt,
		msg.Attributes.FinishedAt,
		msg.Attributes.Duration)
}
