package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/213-team/tg_bot/subscriptionb"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

		log.Printf("%d [%s]: %s",
			update.Message.From.ID,
			update.Message.From.UserName,
			update.Message.Text)

		switch update.Message.Command() {
		case "start":
			reply = "Type /help for help"
		case "help":
			reply = help()
		case "follow", "subscribe":
			reply = follow(update.Message, conn)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}

func help() string {
	answer := "/subscribe or /follow and channel name to subscribe channel\n"
	answer += "/unsubscribe or /unfollow and channel name to unsubscribe channel\n"
	answer += "/recommend or /explore to get recommendations\n"
	return answer
}

func follow(message *tgbotapi.Message, conn *grpc.ClientConn) string {
	messageFields := strings.Fields(message.Text)
	userID := strconv.Itoa(message.From.ID)
	if len(messageFields) != 2 {
		return help()
	}

	c := subscriptionb.NewSubscriptionServiceClient(conn)
	response, err := c.AddSubscription(context.Background(),
		&subscriptionb.AddSubscriptionReq{Subscription: &subscriptionb.Subscription{
			Channel: &subscriptionb.Channel{Id: messageFields[1]},
			User:    &subscriptionb.User{Id: userID}}})
	if err != nil {
		log.Panic("Error when calling follow: %s", err)
	}
	if response.Status.Success {
		return "Now you are following " + messageFields[1]
	}
	return "Try again later"
}
func unfollow(message *tgbotapi.Message, conn *grpc.ClientConn) string {
	messageFields := strings.Fields(message.Text)
	userID := strconv.Itoa(message.From.ID)
	if len(messageFields) != 2 {
		return help()
	}

	c := subscriptionb.NewSubscriptionServiceClient(conn)
	response, err := c.DeleteSubscription(context.Background(),
		&subscriptionb.DeleteSubscriptionReq{Subscription: &subscriptionb.Subscription{
			Channel: &subscriptionb.Channel{Id: messageFields[1]},
			User:    &subscriptionb.User{Id: userID}}})
	if err != nil {
		log.Panic("Error when calling unfollow: %s", err)
	}
	if response.Status.Success {
		return "You left channel " + messageFields[1]
	}
	return "Try again later"
}
