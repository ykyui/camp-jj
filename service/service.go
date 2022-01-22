package service

import (
	"encoding/json"
	"time"
)

func TypeToJson(i interface{}) string {
	json, _ := json.Marshal(i)
	return string(json)
}

func JsonToType(s string, v interface{}) error {
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return err
	}
	return nil
}

type RangeUnit struct {
	Start string
	End   string
}

func BetweenDayList(ru *RangeUnit) []string {
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
