package tgBot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ykyui/camp-jj/database"
	"github.com/ykyui/camp-jj/service"
)

type MyBotSection struct {
	bot        *tgbotapi.BotAPI
	User       *tgbotapi.User
	update     chan bool
	Chat_id    int
	Msg_id     int
	Path       []string
	Contact    interface{}
	Action     string
	ReplyMsgId int
	Camp_id    int
	DisSub     chan<- bool
}

func NewBotSection(user *tgbotapi.User, chat_id int, msg_id int, bot *tgbotapi.BotAPI) *MyBotSection {
	section := MyBotSection{bot, user, make(chan bool, 1), chat_id, msg_id, make([]string, 0), nil, "", 0, 0, nil}
	return &section
}

func (b *MyBotSection) current() string {
	if len(b.Path) == 0 {
		b.Path = append(b.Path, "menu")
	}
	return b.Path[len(b.Path)-1]
}

func (b *MyBotSection) back() string {
	result := b.current()
	b.Path = b.Path[:len(b.Path)-1]
	b.update <- true
	return result
}

func (s *MyBotSection) Idle() {
	ticker := time.NewTicker(time.Second * 120)
	defer ticker.Stop()
	defer func() {
		s.bot.Send(tgbotapi.NewDeleteMessage(int64(s.Chat_id), s.Msg_id))
		if s.ReplyMsgId > 0 {
			s.bot.Send(tgbotapi.NewDeleteMessage(int64(s.Chat_id), s.ReplyMsgId))
		}
		if s.DisSub != nil {
			s.DisSub <- true
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
					temp := msg.(tgbotapi.EditMessageTextConfig)
					temp.ParseMode = tgbotapi.ModeMarkdown
					s.bot.Send(temp)
				}
				ticker.Reset(time.Second * 120)
			} else {
				return
			}
		}
	}
}

func (b *MyBotSection) CallBackHandle(input string) {
	action := strings.Split(input, " ")
	switch action[0] {
	case "direct":
		b.Path = append(b.Path, action[1])
		switch strings.ToLower(b.current()) {
		case "newcamp":
			b.Contact = &service.RangeUnit{Start: "", End: ""}
		case "subcamp":
			b.Camp_id, _ = strconv.Atoi(action[2])
			b.DisSub = database.SubCamp(b.Camp_id, b.update)
		case "equipment", "food":
			if campInfo, err := database.GetCampInfo(b.Camp_id); err == nil {
				if _, ok := campInfo.MemberHeap[int(b.User.ID)]; !ok {
					b.back()
					if msg, err := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), "join camp first")); err == nil {
						b.ReplyMsgId = msg.MessageID
					}

				}
			}
		}
	case "action":
		b.Action = action[1]
		if b.ReplyMsgId > 0 {
			b.bot.Send(tgbotapi.NewDeleteMessage(int64(b.Chat_id), b.ReplyMsgId))
			b.ReplyMsgId = 0
		}
		var replyMsg string
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
		case "quit":
			if campInfo, err := database.GetCampInfo(b.Camp_id); err == nil {
				if user, ok := campInfo.MemberHeap[int(b.User.ID)]; ok {
					if msg, err := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), fmt.Sprintf("%s quit %s", user.Name, user.JoinDate))); err == nil {
						b.bot.Send(tgbotapi.PinChatMessageConfig{ChatID: int64(b.Chat_id), MessageID: msg.MessageID})
					}
				}
			}
			database.Quit(b.Camp_id, b.User)
		case "previousMonth":
			date, _ := time.Parse("2006-01-02", b.Contact.(*service.RangeUnit).Start)
			b.Contact.(*service.RangeUnit).Start = date.AddDate(0, -1, 0).Format("2006-01-02")
			b.Contact.(*service.RangeUnit).End = b.Contact.(*service.RangeUnit).Start
		case "nextMonth":
			date, _ := time.Parse("2006-01-02", b.Contact.(*service.RangeUnit).Start)
			b.Contact.(*service.RangeUnit).Start = date.AddDate(0, 1, 0).Format("2006-01-02")
			b.Contact.(*service.RangeUnit).End = b.Contact.(*service.RangeUnit).Start
		case "setStart":
			b.Contact.(*service.RangeUnit).Start = action[2]
			b.Contact.(*service.RangeUnit).End = b.Contact.(*service.RangeUnit).Start
		case "setEnd":
			date, _ := time.Parse("2006-01-02", b.Contact.(*service.RangeUnit).Start)
			day, _ := strconv.Atoi(action[2])
			b.Contact.(*service.RangeUnit).End = date.AddDate(0, 0, day).Format("2006-01-02")
		case "join":
			if err := database.Join(b.Camp_id, b.User, action[2]); err != nil {
				msg, _ := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), err.Error()))
				b.ReplyMsgId = msg.MessageID
			} else {
				b.back()
				if campInfo, err := database.GetCampInfo(b.Camp_id); err == nil {
					if user, ok := campInfo.MemberHeap[int(b.User.ID)]; ok {
						if msg, err := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), fmt.Sprintf("%s join %s", user.Name, user.JoinDate))); err == nil {
							b.bot.Send(tgbotapi.PinChatMessageConfig{ChatID: int64(b.Chat_id), MessageID: msg.MessageID})
						}
					}
				}
			}
		case "addequipment":
			b.Contact = action[2]
			replyMsg = "reply this message\nequipment name\nequipment name\n.\n.\n.\n."
		case "bringequipment":
			if err := database.BringItem(b.Camp_id, int(b.User.ID), action[2], 2); err != nil {
				replyMsg = err.Error()
			}
		case "dropequipment":
			if err := database.DropItem(b.Camp_id, int(b.User.ID), action[2], 2); err != nil {
				replyMsg = err.Error()
			}
		case "addfood":
			b.Contact = action[2]
			replyMsg = "reply this message\nfood name\ningredients\ningredients\n.\n.\n.\n."
		case "bringfood":
			if err := database.BringItem(b.Camp_id, int(b.User.ID), action[2], 1); err != nil {
				replyMsg = err.Error()
			}
		case "dropfood":
			if err := database.DropItem(b.Camp_id, int(b.User.ID), action[2], 1); err != nil {
				replyMsg = err.Error()
			}
		case "dump":
			return
		}
		if replyMsg != "" {
			msg := tgbotapi.NewMessage(int64(b.Chat_id), replyMsg)
			msg.ReplyToMessageID = b.Msg_id
			//msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
			if msg, err := b.bot.Send(msg); err == nil {
				b.ReplyMsgId = msg.MessageID
			}
		}
	}

	b.update <- true

}

func (b *MyBotSection) ReplyHandle(input string) (result tgbotapi.Chattable) {
	switch b.Action {
	case "addequipment":
		if err := database.AddEquipment(b.Camp_id, b.Contact.(string), strings.Split(input, "\n"), b.User); err != nil {
			msg, _ := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), err.Error()))
			b.ReplyMsgId = msg.MessageID
		}
	case "addfood":
		temp := strings.Split(input, "\n")
		if len(temp) < 1 {
		} else if err := database.AddFood(b.Camp_id, b.Contact.(string), temp[0], temp[1:], b.User); err != nil {
			msg, _ := b.bot.Send(tgbotapi.NewMessage(int64(b.Chat_id), err.Error()))
			b.ReplyMsgId = msg.MessageID
		}
	}
	b.Action = ""
	b.update <- true
	return
}

func (b *MyBotSection) updateMsg() (tgbotapi.Chattable, error) {
	userName := b.User.UserName
	switch strings.ToLower(b.current()) {
	case "menu":
		kb, err := menuKb()
		if err != nil {
			return nil, err
		}
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, fmt.Sprintf("%s\nmenu", userName), kb), nil
	case "newcamp":
		temp := b.Contact.(*service.RangeUnit)
		kb, _ := newCampKb(b.Action, temp)
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, fmt.Sprintf("%s\nnewCamp\ns: %s\ne: %s", userName, temp.Start, temp.End), kb), nil
	case "subcamp":
		camp, err := database.GetCampInfo(b.Camp_id)
		if err != nil {
			return nil, err
		}
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, userName+"\n"+camp.ToMsg(), campMainKb()), nil
	case "join":
		camp, err := database.GetCampInfo(b.Camp_id)
		if err != nil {
			return nil, err
		}
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, userName+"\n"+camp.ToMsg(), joinCampKb(&camp.RangeUnit)), nil
	case "food":
		camp, err := database.GetCampInfo(b.Camp_id)
		if err != nil {
			return nil, err
		}
		ru := camp.RangeUnit
		ru.Start = camp.MemberHeap[int(b.User.ID)].JoinDate
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, userName+"\n"+camp.ToMsg(), campFoodKb(&ru, camp.FoodGroupByDay, int(b.User.ID))), nil
	case "equipment":
		camp, err := database.GetCampInfo(b.Camp_id)
		if err != nil {
			return nil, err
		}
		ru := camp.RangeUnit
		ru.Start = camp.MemberHeap[int(b.User.ID)].JoinDate
		return tgbotapi.NewEditMessageTextAndMarkup(int64(b.Chat_id), b.Msg_id, userName+"\n"+camp.ToMsg(), campEquipmentKb(&ru, camp.EquipmentGroupByDay, int(b.User.ID))), nil
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
