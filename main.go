package main

import (
	"encoding/json"
	"io/ioutil"

	_ "github.com/ykyui/camp-jj/database"
	"github.com/ykyui/camp-jj/handleMsg"
)

type initBotConfig struct {
	TgToken string
}

func main() {
	var config initBotConfig
	file := "./config/tgBotConfig.json"
	data, _ := ioutil.ReadFile(file)
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	myBot := handleMsg.NewBot(config.TgToken)
	myBot.HandleCommand("/campJJ", handleMsg.ShowMenu)
	myBot.HandleCallback("newcamp", handleMsg.NewCamp)
	myBot.HandleCallback("setStartDate", handleMsg.AskCampDate)
	myBot.HandleCallback("setEndDate", handleMsg.AskCampDate)
	myBot.HandleReply("setDate", handleMsg.SetDate)
	myBot.HandleCallback("createCamp", handleMsg.CreateCamp)
	myBot.RunBot()
}
