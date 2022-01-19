package database

import (
	"fmt"
	"time"

	"github.com/ykyui/camp-jj/service"
)

func updateCampInfo(id int) (*CampInfo, error) {
	defer redisDb.Publish(fmt.Sprintf("camp_%d", id), nil)
	redisKey := fmt.Sprintf("campInfo_%d", id)
	campInfo, err := getCampInfo(id)
	if err != nil {
		return nil, err
	}
	return campInfo, redisDb.Set(redisKey, service.TypeToJson(campInfo), time.Second*120).Err()
}

func GetCampInfo(id int) (campInfo *CampInfo, err error) {
	redisKey := fmt.Sprintf("campInfo_%d", id)
	camp, err := redisDb.Get(redisKey).Result()
	if err == nil {
		err = service.JsonToType(camp, &campInfo)
		if err == nil {
			redisDb.Expire(redisKey, time.Second*120)
			return
		}
	}
	return updateCampInfo(id)
}
