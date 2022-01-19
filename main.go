package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ykyui/camp-jj/database"
	_ "github.com/ykyui/camp-jj/database"
	"github.com/ykyui/camp-jj/tgBot"
)

type initBotConfig struct {
	TgToken string
}

func main() {
	var config initBotConfig
	file := "./config/tgBotConfig.json"
	data, _ := ioutil.ReadFile(file)
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	bot, err := tgbotapi.NewBotAPI(config.TgToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	sectionHeap := make(map[int]*tgBot.MyBotSection)
	for update := range updates {
		if update.Message != nil {
			chatId := update.Message.Chat.ID
			msgId := update.Message.MessageID
			input := update.Message.Text
			if err := database.CheckValid(int(chatId)); err != nil {
				fmt.Println(err)
				continue
			}
			if update.Message.IsCommand() {
				if msg, err := bot.Send(tgbotapi.NewMessage(chatId, time.Now().String())); err != nil {
					fmt.Println(err)
				} else {
					newSection := tgBot.NewBotSection(int(chatId), msg.MessageID, bot)
					sectionHeap[newSection.Msg_id] = newSection
					go func() {
						defer delete(sectionHeap, newSection.Msg_id)
						newSection.Idle()
					}()
					switch strings.ToLower(update.Message.Command()) {
					case "campjj":
						sectionHeap[msg.MessageID].CallBackHandle("direct menu", update.Message.From)
					}
				}
			} else {
				replyMsgId := update.Message.ReplyToMessage.MessageID
				// replyMsg := update.Message.ReplyToMessage.Text
				bot.Send(tgbotapi.NewDeleteMessage(chatId, replyMsgId))
				bot.Send(tgbotapi.NewDeleteMessage(chatId, msgId))
				for _, v := range sectionHeap {
					if v.ReplyMsgId == replyMsgId {
						v.ReplyHandle(input, update.Message.From)
					}
				}
			}

		} else if update.CallbackQuery != nil {
			chatId := update.CallbackQuery.Message.Chat.ID
			msgId := update.CallbackQuery.Message.MessageID
			input := update.CallbackQuery.Data
			if section, ok := sectionHeap[msgId]; ok {
				section.CallBackHandle(input, update.CallbackQuery.From)
			} else {
				bot.Send(tgbotapi.NewDeleteMessage(chatId, msgId))
			}
		}
	}
}
