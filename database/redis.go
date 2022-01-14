package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-redis/redis"
)

type redisDbConfig struct {
	Host     string
	Post     int
	Password string
	DB       int
}

var redisDb *redis.Client

func init() {
	var config redisDbConfig
	file := "./config/redisConfig.json"
	data, _ := ioutil.ReadFile(file)
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Post),
		Password: config.Password, // no password set
		DB:       config.DB,       // use default DB
	})

	if _, err := client.Ping().Result(); err != nil {
		panic(err)
	}
	redisDb = client
}

func GetMsgById(msgId int) (string, error) {
	return redisDb.Get(fmt.Sprintf("msg_%d", msgId)).Result()
}

func typeToJson(i interface{}) string {
	json, _ := json.Marshal(i)
	return string(json)
}

func jsonToType(s string, v interface{}) error {
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return err
	}
	return nil
}
