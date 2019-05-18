package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/213/tg_bot/subscriptionb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

// Configuration with bot token
type Configuration struct {
	TelegramBotToken string
}

var (
	telegramBotToken string
)

func init() {
	file, _ := os.Open("configs/secrets.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	telegramBotToken = configuration.TelegramBotToken
}

func main() {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		reply := "Type /help for help"
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s (%s)",
			update.Message.From.UserName,
			update.Message.Text,
			update.Message.Command())

		switch update.Message.Command() {
		case "start":
			reply = "Type /help for help"
		case "help":
			reply = help()
		case "follow", "subscribe":
			reply = follow(update.Message.Text, conn)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}

func help() string {
	answer := "/subscribe or /follow to subscribe channel\n"
	answer += "/unsubscribe or /unfollow to unsubscribe channel\n"
	answer += "/recommend or /explore to get recommendations\n"
	return answer
}

func follow(message string, conn *ClientConn) string {
	messageFields := strings.Fields(message)

	c := subscriptionb.NewChannelServiceClient(conn)
	response, err := c.ReadChannel(context.Background(), &subscriptionb.ReadChannelReq{Id: "333"})
	if err != nil {
		log.Fatalf("Error when calling SayHello: %s", err)
	}
	log.Printf("Response from server: %s", response.Channel.UserId)
	return "Сашка! протобуфину мне запили. а то не могу сделать " +
		strings.Join(messageFields, " ")
}
