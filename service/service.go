package service

import "encoding/json"

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
