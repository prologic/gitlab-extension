package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/prologic/bitcask"
	"github.com/sirupsen/logrus"
	"server/conf"
	"server/contracts"
	"server/logging"
	"server/queue"
	"server/utils"
	"strconv"
	"strings"
)

const (
	BotTopic   = "telegram_bot"
	chatPrefix = "chat"
)

type GitlabMessage contracts.PipelinePush

// Telegram bot that forwards messages form global queue topic to telegram chats.
type Bot struct {
	*tgbotapi.BotAPI
	*conf.Config
	db        *bitcask.Bitcask
	queue     *queue.GlobalQueue
	logger    *logrus.Entry
	updatesCh tgbotapi.UpdatesChannel
}

// Creates new instance of telegram bot.
func NewBot(
	botApi *tgbotapi.BotAPI,
	db *bitcask.Bitcask,
	queue *queue.GlobalQueue,
	config *conf.Config,
	logger *logrus.Logger) *Bot {

	bot := &Bot{}
	bot.BotAPI = botApi
	bot.db = db
	bot.queue = queue
	bot.Config = config
	bot.logger = logger.WithField(logging.CorrelationIdKey, uuid.New())
	bot.updatesCh = bot.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 5})
	bot.logger.Info("Telegram bot initialized")
	return bot
}

// Start handling messages.
func (bot *Bot) Start() {
	bot.subscribeToTopic()
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
				bot.Send(chatId, "You do not have any available namespaces.")
				continue
			}

			err := bot.setChatNamespaces(chatId, availableNamespaces)
			if err != nil {
				bot.logger.Errorf("Error while updating gitlab namespaces for chat id=%d", chatId)
				bot.Send(chatId, "Sorry something went wrong.")
			}
			availableNamespaces, err = bot.getChatNamespaces(chatId)
			if err != nil {
				bot.logger.Errorf("Error while updating gitlab namespaces for chat id=%d", chatId)
				bot.Send(chatId, "Sorry something went wrong.")
			}
			namespacesString := strings.Join(availableNamespaces, ", ")
			bot.Send(chatId, fmt.Sprintf("You have been subscribed to namespaces: %s", namespacesString))
		}
	}()
}

// Send message to chat.
// chatId - identifier of telegram chat.
// text - message's text.
func (bot *Bot) Send(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	reply, err := bot.BotAPI.Send(msg)
	if err != nil {
		bot.logger.Errorf("Error while sending message to chat id=%d", chatId)
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
	err = bot.db.Lock()
	if err != nil {
		return
	}
	defer bot.db.Unlock()
	for _, ns := range namespaces {
		key := fmt.Sprintf("%s_%d_%s", chatPrefix, chatId, ns)
		if bot.db.Has(key) {
			continue
		}
		err = bot.db.Put(key, []byte("<empty>"))
		if err != nil {
			return
		}
	}
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
		bot.logger.Errorf("error while decoding JSON body: %v", err)
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
func (bot *Bot) subscribeToTopic() {
	bot.queue.AddTopic(BotTopic)
	bot.queue.Subscribe(BotTopic, func(message interface{}) {
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
			bot.logger.Errorf("Error while sending gitlab update to telegram: %v", err)
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
