package tgBot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ykyui/camp-jj/database"
	"github.com/ykyui/camp-jj/service"
)

type MyBotSection struct {
	bot        *tgbotapi.BotAPI
	update     chan bool
	Chat_id    int
	Msg_id     int
	Path       []string
	Contact    interface{}
	Action     string
	ReplyMsgId int
}

func NewBotSection(chat_id int, msg_id int, bot *tgbotapi.BotAPI) *MyBotSection {
	section := MyBotSection{bot, make(chan bool, 1), chat_id, msg_id, make([]string, 0), nil, "", 0}
	return &section
}

func (b *MyBotSection) current() string {
	return b.Path[len(b.Path)-1]
}

func (b *MyBotSection) back() string {
	result := b.current()
	b.Path = b.Path[:len(b.Path)-1]
	b.update <- true
	return result
}

func (b *MyBotSection) ReplyHandle(input string) (result tgbotapi.Chattable) {
	switch b.Action {
	case "setStart":
		b.Contact.(*service.RangeUnit).Start = input
	case "setEnd":
		b.Contact.(*service.RangeUnit).End = input
	}
	b.Action = ""
	b.update <- true
	return
}

func (b *MyBotSection) CallBackHandle(input string) {
	action := strings.Split(input, " ")
	switch action[0] {
	case "direct":
		b.Path = append(b.Path, action[1])
		switch strings.ToLower(b.current()) {
		case "menu":
			break
		case "newcamp":
			b.Contact = &service.RangeUnit{Start: "yyyymmdd", End: "yyyymmdd"}
		}
	case "action":
		b.Action = action[1]
		if b.ReplyMsgId > 0 {
			b.bot.Send(tgbotapi.NewDeleteMessage(int64(b.Chat_id), b.ReplyMsgId))
			b.ReplyMsgId = 0
		}
		switch b.Action {
		case "back":
			b.back()
		case "confirm":
			if err := b.confirmAction(); err != nil {
				msg, _ := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), err.Error()))
				b.ReplyMsgId = msg.MessageID
			} else {
				b.back()
			}
		default:
			msg := tgbotapi.NewMessage(int64(b.Chat_id), "please replay this message")
			msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
			if msg, err := b.bot.Send(msg); err == nil {
				b.ReplyMsgId = msg.MessageID
			}
		}
	}

	b.update <- true
}

func (s *MyBotSection) Idle() {
	ticker := time.NewTicker(time.Second * 120)
	defer ticker.Stop()
	defer func() {
		s.bot.Send(tgbotapi.NewDeleteMessage(int64(s.Chat_id), s.Msg_id))
		if s.ReplyMsgId > 0 {
			s.bot.Send(tgbotapi.NewDeleteMessage(int64(s.Chat_id), s.ReplyMsgId))
		}
	}()
	for {
		select {
		case <-ticker.C:
			return
		case r := <-s.update:
			if r {
				if msg, err := s.updateMsg(); err != nil {
					s.bot.Send(tgbotapi.NewEditMessageText(int64(s.Chat_id), s.Msg_id, err.Error()))
				} else {
					s.bot.Send(msg)
				}
				ticker.Reset(time.Second * 120)
			} else {
				ticker.Stop()
			}
		}
	}
}

func (b *MyBotSection) updateMsg() (tgbotapi.Chattable, error) {
	switch strings.ToLower(b.current()) {
	case "menu":
		kb, err := menuKb()
		if err != nil {
			return nil, err
		}
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, "menu", *kb), nil
	case "newcamp":
		kb, _ := newCampKb()
		temp := b.Contact.(*service.RangeUnit)
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, fmt.Sprintf("newCamp\ns: %s\ne: %s", temp.Start, temp.End), *kb), nil
	}
	return nil, nil
}

func (b *MyBotSection) confirmAction() error {
	switch strings.ToLower(b.current()) {
	case "newcamp":
		temp := b.Contact.(*service.RangeUnit)
		return database.CreateCamp(*temp)
	}
	return nil
}
