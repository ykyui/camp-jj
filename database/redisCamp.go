package database

import (
	"fmt"
	"time"

	"github.com/ykyui/camp-jj/service"
)

func NewCamp(msgId int) service.RangeUnit {
	rangeUnit := service.RangeUnit{}
	redisDb.Set(fmt.Sprintf("newCamp_%d", msgId), service.TypeToJson(rangeUnit), time.Second*120)
	return rangeUnit
}

func GetNewCamp(msgId int) (*service.RangeUnit, error) {
	var rangeUnit service.RangeUnit
	s, err := redisDb.Get(fmt.Sprintf("newCamp_%d", msgId)).Result()
	if err != nil {
		return nil, err
	}
	if err = service.JsonToType(s, &rangeUnit); err != nil {
		return nil, err
	}
	return &rangeUnit, nil
}

func UpdateNewCamp(msgId int, rangeUnit service.RangeUnit) {
	redisDb.Set(fmt.Sprintf("newCamp_%d", msgId), service.TypeToJson(rangeUnit), time.Second*120)
}
