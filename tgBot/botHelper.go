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
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("food", "direct food"), tgbotapi.NewInlineKeyboardButtonData("equipment", "direct equipment")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")),
	)
}

func campEquipmentKb() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("add", "action addEquipment") /* tgbotapi.NewInlineKeyboardButtonData("delete", "direct deleteEquipment") */),
		// tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("bring", "direct bringEquipment")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")),
	)
}

func bringEquipmentKb(userId int, equipment map[int]database.Item) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup()
	for id, v := range equipment {
		if len(kb.InlineKeyboard) == 0 || len(kb.InlineKeyboard[len(kb.InlineKeyboard)-1])%2 == 0 {
			kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow())
		}
		if v.WhoBring == 0 {
			kb.InlineKeyboard[len(kb.InlineKeyboard)-1] = append(kb.InlineKeyboard[len(kb.InlineKeyboard)-1], tgbotapi.NewInlineKeyboardButtonData("bring "+v.Name, fmt.Sprintf("action bringEquipment %d", id)))
		} else if v.WhoBring == userId {
			kb.InlineKeyboard[len(kb.InlineKeyboard)-1] = append(kb.InlineKeyboard[len(kb.InlineKeyboard)-1], tgbotapi.NewInlineKeyboardButtonData("drop "+v.Name, fmt.Sprintf("action bringEquipment %d", id)))
		}
	}
	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")))
	return kb
}

func campFoodKb() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("add", "action addFood") /* tgbotapi.NewInlineKeyboardButtonData("delete", "direct deleteFood") */),
		// tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("bring", "direct bringFood")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")),
	)
}
