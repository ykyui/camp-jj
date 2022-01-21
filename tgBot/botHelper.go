package tgBot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ykyui/camp-jj/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ykyui/camp-jj/service"
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

func newCampKb(action string, ru *service.RangeUnit) (tgbotapi.InlineKeyboardMarkup, error) {
	if ru.Start == "" {
		ru.Start = time.Now().Format("2006-01-02")
	}

	currentDate, _ := time.Parse("2006-01-02", ru.Start)
	result := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("<", "action previousMonth"), tgbotapi.NewInlineKeyboardButtonData(currentDate.Format("2006-01"), "dump"), tgbotapi.NewInlineKeyboardButtonData(">", "action nextMonth")),
	)

	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow())
	for _, v := range []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"} {
		result.InlineKeyboard[len(result.InlineKeyboard)-1] = append(result.InlineKeyboard[len(result.InlineKeyboard)-1], tgbotapi.NewInlineKeyboardButtonData(v, "dump"))
	}

	for _, v := range monthPattern(currentDate) {
		result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow())
		for _, v := range v {
			text := v
			data := v
			if text == "dump" {
				text = " "
			} else {
				data = fmt.Sprintf("setStart %d-%02d-%02s", currentDate.Year(), currentDate.Month(), data)
			}
			result.InlineKeyboard[len(result.InlineKeyboard)-1] = append(result.InlineKeyboard[len(result.InlineKeyboard)-1], tgbotapi.NewInlineKeyboardButtonData(text, fmt.Sprintf("action %s", data)))
		}
	}

	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow())
	for i := 0; i < 7; i++ {
		result.InlineKeyboard[len(result.InlineKeyboard)-1] = append(result.InlineKeyboard[len(result.InlineKeyboard)-1], tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i), fmt.Sprintf("action setEnd %d", i)))
	}

	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("confirm", "action confirm"), tgbotapi.NewInlineKeyboardButtonData("back", "action back")))

	return result, nil
}

func campMainKb() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("join?", "direct join"), tgbotapi.NewInlineKeyboardButtonData("quit", "action quit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("food", "direct food"), tgbotapi.NewInlineKeyboardButtonData("equipment", "direct equipment")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")),
	)
}

func joinCampKb(ru *service.RangeUnit) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup()

	for _, v := range betweenDayList(ru) {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v, "action join "+v)))
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")))

	return kb
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

func campFoodKb(ru *service.RangeUnit) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup()
	for _, v := range betweenDayList(ru) {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("food for "+v, "action addfood "+v)))
	}
	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("back", "action back")))
	return kb
}

func monthPattern(d time.Time) [][]string {
	startDate := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
	lastDayOfMonth := startDate.AddDate(0, 1, -1)
	result := [][]string{{}}

	for i := 0; i < int(startDate.Weekday()); i++ {
		result[0] = append(result[0], "dump")
	}
	for i := 0; i < lastDayOfMonth.Day(); i++ {
		if len(result[len(result)-1])%7 == 0 {
			result = append(result, []string{})
		}
		result[len(result)-1] = append(result[len(result)-1], strconv.Itoa(i+1))
	}
	for i := 0; i < 6-int(lastDayOfMonth.Weekday()); i++ {
		result[len(result)-1] = append(result[len(result)-1], "dump")
	}

	return result
}

func betweenDayList(ru *service.RangeUnit) []string {
	result := make([]string, 0)
	s, _ := time.Parse("2006-01-02", ru.Start)
	e, _ := time.Parse("2006-01-02", ru.End)
	for {
		result = append(result, s.Format("2006-01-02"))
		s = s.AddDate(0, 0, 1)
		if e.Sub(s).Hours() < 0 {
			return result
		}
	}
}
