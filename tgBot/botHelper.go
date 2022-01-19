package tgBot

import (
	"fmt"

	"github.com/ykyui/camp-jj/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func menuKb() (tgbotapi.InlineKeyboardMarkup, error) {
	result := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("new camp", "direct newCamp")),
	)
	camp, err := database.GetCampList()
	if err != nil {
		return result, err
	}
	for _, v := range camp {
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s To %s", v.RangeUnit.Start, v.RangeUnit.End), fmt.Sprintf("direct subcamp %d", v.Id))))
	}
	return result, nil
}

func newCampKb() (tgbotapi.InlineKeyboardMarkup, error) {
	result := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("setStart", "action setStart"), tgbotapi.NewInlineKeyboardButtonData("setEnd", "action setEnd")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("confirm", "action confirm"), tgbotapi.NewInlineKeyboardButtonData("back", "action back")),
	)
	return result, nil
}

func campMainKb() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("join?", "action join"), tgbotapi.NewInlineKeyboardButtonData("quit", "action quit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")),
	)
}
