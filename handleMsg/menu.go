package handleMsg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ykyui/camp-jj/database"
	// . "github.com/ykyui/camp-jj/service"
)

func ShowMenu(chatId int64, msgId int, input string) (tgbotapi.Chattable, error) {
	msg := tgbotapi.NewMessage(chatId, "menu")
	msg.ReplyToMessageID = msgId
	if kb, err := getMenuInlineKeyboard(); err != nil {
		return nil, err
	} else {
		msg.ReplyMarkup = kb
	}
	return msg, nil
}

func BackToMenu(chatId int64, msgId int, input string) (tgbotapi.Chattable, error) {
	if kb, err := getMenuInlineKeyboard(); err != nil {
		return nil, err
	} else {
		return tgbotapi.NewEditMessageTextAndMarkup(chatId, msgId, "menu", *kb), nil
	}
}

func getMenuInlineKeyboard() (*tgbotapi.InlineKeyboardMarkup, error) {
	campList, err := database.GetCampList()
	if err != nil {
		return nil, err
	}
	kbg := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("New Camp", "newCamp"),
		),
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonData("Confirm", "confirm"),
		// ),
	)
	for _, v := range campList {

		bt := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s to %s", v.RangeUnit.Start, v.RangeUnit.End), fmt.Sprintf("getCampInfo\n%d", v.Id))
		kbg.InlineKeyboard = append(kbg.InlineKeyboard, []tgbotapi.InlineKeyboardButton{bt})

	}

	return &kbg, nil
}
