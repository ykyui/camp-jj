package handleMsg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ykyui/camp-jj/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//callback
func NewCamp(chatId int64, msgId int, self string, input string) (tgbotapi.Chattable, error) {
	database.NewCamp(msgId)
	return showDate(chatId, msgId)
}

func showDate(chatId int64, msgId int) (tgbotapi.Chattable, error) {
	rangeUnit, err := database.GetNewCamp(msgId)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewEditMessageTextAndMarkup(chatId, msgId, fmt.Sprintf("newCamp\ns:%s\ne:%s", rangeUnit.Start, rangeUnit.End),
		tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("start: %s", rangeUnit.Start), "setStartDate\nsetStartDate"),
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("end: %s", rangeUnit.End), "setEndDate\nsetEndDate"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("confirm", "createCamp"),
			),
		),
	)
	return msg, nil

}

//callback
func AskCampDate(chatId int64, msgId int, self string, input string) (tgbotapi.Chattable, error) {
	msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("setDate %s %d", input, msgId))
	msg.ReplyToMessageID = msgId
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, InputFieldPlaceholder: "yyyyMMdd", Selective: false}
	return msg, nil
}

//reply
func SetDate(chatId int64, msgId int, self []string, input string) (tgbotapi.Chattable, error) {
	dataSet := strings.Split(self[0], " ")
	editTargatMsgId, _ := strconv.Atoi(dataSet[2])
	rangeUnit, err := database.GetNewCamp(editTargatMsgId)
	if err != nil {
		return nil, err
	}
	switch dataSet[1] {
	case "setStartDate":
		rangeUnit.Start = input
	case "setEndDate":
		rangeUnit.End = input
	}
	database.UpdateNewCamp(editTargatMsgId, *rangeUnit)
	return showDate(chatId, editTargatMsgId)
}

// callback
func CreateCamp(chatId int64, msgId int, self string, input string) (tgbotapi.Chattable, error) {
	rangeUnit, err := database.GetNewCamp(msgId)
	if err != nil {
		return nil, err
	}
	database.CreateCamp(*rangeUnit)
	return BackToMenu(chatId, msgId, input)
}

//callback
