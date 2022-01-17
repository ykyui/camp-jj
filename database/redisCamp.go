package database

import (
	"fmt"
	"time"

	. "github.com/ykyui/camp-jj/service"
)

type RangeUnit struct {
	Start string
	End   string
}

func NewCamp(msgId int) RangeUnit {
	rangeUnit := RangeUnit{}
	redisDb.Set(fmt.Sprintf("newCamp_%d", msgId), TypeToJson(rangeUnit), time.Second*120)
	return rangeUnit
}

func GetNewCamp(msgId int) (*RangeUnit, error) {
	var rangeUnit RangeUnit
	s, err := redisDb.Get(fmt.Sprintf("newCamp_%d", msgId)).Result()
	if err != nil {
		return nil, err
	}
	if err = JsonToType(s, &rangeUnit); err != nil {
		return nil, err
	}
	return &rangeUnit, nil
}

func UpdateNewCamp(msgId int, rangeUnit RangeUnit) {
	redisDb.Set(fmt.Sprintf("newCamp_%d", msgId), TypeToJson(rangeUnit), time.Second*120)
}
