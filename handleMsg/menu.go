package handleMsg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ShowMenu(chatId int64, msgId int, input string) (tgbotapi.Chattable, error) {
	msg := tgbotapi.NewMessage(chatId, "menu")
	msg.ReplyToMessageID = msgId
	msg.ReplyMarkup = getMenuInlineKeyboard()
	return msg, nil
}

func BackToMenu(chatId int64, msgId int, input string) (tgbotapi.Chattable, error) {
	return tgbotapi.NewEditMessageTextAndMarkup(chatId, msgId, "menu", getMenuInlineKeyboard()), nil
}

func getMenuInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("New Camp", "newCamp"),
		),
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonData("Confirm", "confirm"),
		// ),
	)
}
