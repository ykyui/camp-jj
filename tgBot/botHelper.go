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

	result.InlineKeyboard = append(result.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("confirm", "action confirm"), tgbotapi.NewInlineKeyboardButtonData("🔙", "action back")))

	return result, nil
}

func campMainKb() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🤝", "direct join"), tgbotapi.NewInlineKeyboardButtonData("👋", "action quit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🍻🍷🍺", "direct food"), tgbotapi.NewInlineKeyboardButtonData("🎒", "direct equipment")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙", "action back")),
	)
}

func joinCampKb(ru *service.RangeUnit) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup()

	for _, v := range service.BetweenDayList(ru) {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v, "action join "+v)))
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙", "action back")))

	return kb
}

func campEquipmentKb(ru *service.RangeUnit, equipment map[string][]*database.Item, userId int) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup()
	for _, v := range service.BetweenDayList(ru) {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s ➕ 🎒 ", v), "action addequipment "+v)))
		if e, ok := equipment[v]; ok {
			for i, v := range e {
				if i%2 == 0 {
					kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow())
				}
				index := len(kb.InlineKeyboard) - 1
				if v.WhoBring == userId {
					kb.InlineKeyboard[index] = append(kb.InlineKeyboard[index], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("❌ %s", v.Name), fmt.Sprintf("action dropequipment %d", v.Id)))
				} else if v.WhoBring == 0 {
					kb.InlineKeyboard[index] = append(kb.InlineKeyboard[index], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✅ %s", v.Name), fmt.Sprintf("action bringequipment %d", v.Id)))
				}
			}
		}
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙", "action back")))
	return kb
}

func campFoodKb(ru *service.RangeUnit, food map[string][]*database.Food, userId int) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup()
	for _, v := range service.BetweenDayList(ru) {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s ➕ 🍳 ", v), "action addfood "+v)))
		if f, ok := food[v]; ok {
			count := 0
			for _, v := range f {
				for _, v := range v.Ingredients {
					if count%2 == 0 {
						kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow())
					}
					count++
					index := len(kb.InlineKeyboard) - 1
					if v.WhoBring == userId {
						kb.InlineKeyboard[index] = append(kb.InlineKeyboard[index], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("❌ %s", v.Name), fmt.Sprintf("action dropfood %d", v.Id)))
					} else if v.WhoBring == 0 {
						kb.InlineKeyboard[index] = append(kb.InlineKeyboard[index], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✅ %s", v.Name), fmt.Sprintf("action bringfood %d", v.Id)))
					}
				}
			}
		}
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙", "action back")))
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
