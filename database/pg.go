package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq"
)

type pgDbConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Db       string
}

var pgDb *sql.DB

func init() {
	var config pgDbConfig
	file := "./config/pgConfig.json"
	data, _ := ioutil.ReadFile(file)
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", config.Host, config.Username, config.Password, config.Db, config.Port))
	if err != nil {
		panic(err)
	}
	pgDb = db
}

func CheckValid(chatId int) error {
	stmt, err := pgDb.Prepare("select count(*) from allow_chat where id = $1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	var count sql.NullInt64
	if err := stmt.QueryRow(chatId).Scan(&count); err != nil {
		return err
	} else if count.Int64 == 0 {
		return fmt.Errorf("uid:%d not valid", chatId)
	}

	return nil
}
