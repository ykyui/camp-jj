package handleMsg

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ykyui/camp-jj/database"
)

type MyBot struct {
	bot *tgbotapi.BotAPI

	msg      map[string]MyBotMsgHandle
	command  map[string]MyBotMsgHandle
	callBack map[string]MyBotCallbackHandle
	reply    map[string]MyBotReplyHandle
}

func NewBot(token string) *MyBot {
	_bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	_bot.Debug = true

	log.Printf("Authorized on account %s", _bot.Self.UserName)

	return &MyBot{_bot, make(map[string]MyBotMsgHandle), make(map[string]MyBotMsgHandle), make(map[string]MyBotCallbackHandle), make(map[string]MyBotReplyHandle)}
}

func (b *MyBot) RunBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			chatId := update.Message.Chat.ID
			msgId := update.Message.MessageID
			input := update.Message.Text
			if err := database.CheckValid(int(chatId)); err != nil {
				fmt.Println(err)
				continue
			}
			if update.Message.IsCommand() {
				if f, ok := b.command[strings.ToLower(input)]; ok {
					msg, err := f(chatId, msgId, input)
					if err != nil {
						fmt.Println(err)
						msg = tgbotapi.NewMessage(chatId, err.Error())
					}
					sent, _ := b.bot.Send(msg)
					go func(chatId int64, msgId int) {
						time.Sleep(time.Second * 125)
						if _, err := database.GetMsgById(msgId); err != nil {
							b.bot.Send(tgbotapi.NewDeleteMessage(chatId, msgId))
						}
					}(chatId, sent.MessageID)
				} else {
					fmt.Println(input)
				}
			} else {
				replyMsgId := update.Message.ReplyToMessage.MessageID
				replyMsg := strings.Split(update.Message.ReplyToMessage.Text, "\n")
				action := strings.Split(replyMsg[0], " ")[0]
				b.bot.Send(tgbotapi.NewDeleteMessage(chatId, replyMsgId))
				b.bot.Send(tgbotapi.NewDeleteMessage(chatId, msgId))
				if f, ok := b.reply[strings.ToLower(action)]; ok {
					msg, err := f(chatId, msgId, replyMsg, input)
					if err != nil {
						fmt.Println(err)
						continue
					}
					b.bot.Send(msg)
				} else {
					fmt.Println(input)
				}
			}

		} else if update.CallbackQuery != nil {
			chatId := update.CallbackQuery.Message.Chat.ID
			msgId := update.CallbackQuery.Message.MessageID
			self := strings.Split(update.CallbackQuery.Message.Text, "\n")
			input := strings.Split(update.CallbackQuery.Data, "\n")
			if f, ok := b.callBack[strings.ToLower(input[0])]; ok {
				msg, err := f(chatId, msgId, self[0], strings.Join(input[1:len(input)], "\n"))
				if err != nil {
					fmt.Println(err)
					msg = tgbotapi.NewEditMessageText(chatId, msgId, err.Error())
				}
				sent, _ := b.bot.Send(msg)
				if _, ok := msg.(tgbotapi.MessageConfig); ok {
					go func(chatId int64, msgId int) {
						time.Sleep(time.Second * 60)
						b.bot.Send(tgbotapi.NewDeleteMessage(chatId, msgId))
					}(chatId, sent.MessageID)
				}
			} else {
				fmt.Println(input)
			}
		}
	}
}

type MyBotMsgHandle func(chatId int64, msgId int, input string) (tgbotapi.Chattable, error)

func (b *MyBot) HandleCommand(action string, handle MyBotMsgHandle) {
	b.command[strings.ToLower(action)] = handle
}

func (b *MyBot) HandleNormalMsg(action string, handle MyBotMsgHandle) {
	b.msg[strings.ToLower(action)] = handle
}

type MyBotCallbackHandle func(chatId int64, msgId int, self string, input string) (tgbotapi.Chattable, error)

func (b *MyBot) HandleCallback(action string, handle MyBotCallbackHandle) {
	b.callBack[strings.ToLower(action)] = handle
}

type MyBotReplyHandle func(chatId int64, msgId int, self []string, input string) (tgbotapi.Chattable, error)

func (b *MyBot) HandleReply(action string, handle MyBotReplyHandle) {
	b.reply[strings.ToLower(action)] = handle
}
