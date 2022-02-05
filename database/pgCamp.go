package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ykyui/camp-jj/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateCamp(ru service.RangeUnit, userId int) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`insert into camp (id, start_date, end_date, create_by) 
	select coalesce(max(id),0)+1,TO_DATE($1,'YYYY-MM-DD'),TO_DATE($2,'YYYY-MM-DD'), $3 from camp`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(ru.Start, ru.End, userId); err != nil {
		return err
	}
	return tx.Commit()
}

type CampInfo struct {
	Id                   int
	Name                 string
	RangeUnit            service.RangeUnit
	FoodHeap             map[int]*Food
	EquipmentHeap        map[int]*Item
	MemberHeap           map[int]*Member
	CreateBy             int
	CreateByName         string
	MemberGroupByJoinDay map[string][]*Member
	FoodGroupByDay       map[string][]*Food
	EquipmentGroupByDay  map[string][]*Item
}

type Food struct {
	Id          int
	Name        string
	Date        string
	Ingredients map[int]Item
}

type Item struct {
	Id       int
	Name     string
	WhoBring int
	Date     string
}

type Member struct {
	Name      string
	JoinDate  string
	Food      []int
	Equipment []int
}

const (
	checkEmoji = "\xE2\x9C\x85"
	crossEmoji = "\xE2\x9D\x8C"
)

func (c *CampInfo) ToMsg() (result string) {
	result = fmt.Sprintf("create by:%s\nname:%s\n🏕️: %d\n📅: %s To %s\n", c.CreateByName, c.Name, c.Id, c.RangeUnit.Start, c.RangeUnit.End)

	result += "🧍‍♀️🧍🧍‍♂️👭 \n"
	count := 1
	for _, v := range service.BetweenDayList(&c.RangeUnit) {
		if user, ok := c.MemberGroupByJoinDay[v]; ok {
			result += fmt.Sprintf("_%s_ \n", v)
			for _, v := range user {
				result += fmt.Sprintf("%02d. %s (%d)\n", count, v.Name, len(v.Equipment)+len(v.Food))
				count++
			}
		}
	}

	result += "\n🥫 🍝 🍜 🍲 🍛  \n"
	for _, v := range service.BetweenDayList(&c.RangeUnit) {
		count := 1
		if food, ok := c.FoodGroupByDay[v]; ok {
			result += fmt.Sprintf("_%s (%d)_\n", v, len(food))
			for _, v := range food {
				result += fmt.Sprintf("%d. %s\n", count, v.Name)
				for _, v := range v.Ingredients {
					var emoji string
					var bringName string
					if user, ok := c.MemberHeap[v.WhoBring]; ok {
						emoji = checkEmoji
						bringName = fmt.Sprintf("(%s)", user.Name)
					} else {
						emoji = crossEmoji
					}
					result += fmt.Sprintf("    %s %s %s\n", emoji, v.Name, bringName)
				}
				count++
			}
		}
	}

	result += "\n🕳✂🔪🧷📌🎒 \n"
	for _, v := range service.BetweenDayList(&c.RangeUnit) {
		if equipment, ok := c.EquipmentGroupByDay[v]; ok {
			result += fmt.Sprintf("_%s (%d)_\n", v, len(equipment))
			for _, v := range equipment {
				var emoji string
				var bringName string
				if user, ok := c.MemberHeap[v.WhoBring]; ok {
					emoji = checkEmoji
					bringName = fmt.Sprintf("(%s)", user.Name)
				} else {
					emoji = crossEmoji
				}
				result += fmt.Sprintf("%s %s %s\n", emoji, v.Name, bringName)
			}
			result += "\n"
		}
	}
	return
}

func GetCampList() (result []*CampInfo, err error) {
	stmt, err := pgDb.Prepare(`select camp.id, name, TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD'), create_by, user_name
	from camp 
	left join group_user on create_by = group_user.id 
	where current_date - INTERVAL '7 day' < end_date`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if rows, err := stmt.Query(); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				id         sql.NullInt64
				name       sql.NullString
				start      sql.NullString
				end        sql.NullString
				createBy   sql.NullInt64
				createName sql.NullString
			)
			if err = rows.Scan(&id, &name, &start, &end, &createBy, &createName); err != nil {
				return nil, err
			}
			result = append(result, &CampInfo{int(id.Int64), name.String, service.RangeUnit{Start: start.String, End: end.String}, nil, nil, nil, int(createBy.Int64), createName.String, nil, nil, nil})
		}
	}
	return result, nil
}

func getCampInfo(id int) (*CampInfo, error) {
	stmt_camp, err := pgDb.Prepare(`select name, TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD'), create_by, user_name 
	from camp 
	left join group_user on create_by = group_user.id 
	where camp.id = $1`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp.Close()

	stmt_camp_user, err := pgDb.Prepare(`select user_id, TO_CHAR(join_date, 'YYYY-MM-DD'), user_name 
	from camp_member 
	left join group_user on user_id = id 
	where camp_id = $1
	order by join_date`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_user.Close()

	stmt_camp_food, err := pgDb.Prepare(`select cf.id, cf.name, TO_CHAR(cf.date, 'YYYY-MM-DD'), cfm.id, cfm.name, cub.user_id
	from camp_food cf 
	left join camp_food_ingredients cfm 
		left join camp_user_bring cub on cfm.id = cub.item_id and cub.camp_id = cfm.camp_id and cub.type = 1 
	on cf.id = cfm.food_id and cf.camp_id = cfm.camp_id
	where cf.camp_id = $1
	order by cf.date`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_food.Close()

	stmt_camp_equipment, err := pgDb.Prepare(`select ce.id, ce.name, cub.user_id, TO_CHAR(date, 'YYYY-MM-DD')
	from camp_equipment ce 
	left join camp_user_bring cub on  ce.id = cub.item_id and ce.camp_id = cub.camp_id and cub.type = 2
	where ce.camp_id = $1
	order by date`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_equipment.Close()

	var (
		start_date     sql.NullString
		end_date       sql.NullString
		name           sql.NullString
		create_by      sql.NullInt64
		create_by_name sql.NullString
	)
	if err = stmt_camp.QueryRow(id).Scan(&name, &start_date, &end_date, &create_by, &create_by_name); err != nil {
		return nil, err
	}
	campInfo := &CampInfo{id, name.String, service.RangeUnit{Start: start_date.String, End: end_date.String}, make(map[int]*Food), make(map[int]*Item), make(map[int]*Member), int(create_by.Int64), create_by_name.String, make(map[string][]*Member), make(map[string][]*Food), make(map[string][]*Item)}

	if rows, err := stmt_camp_user.Query(id); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			user_id   sql.NullInt64
			join_date sql.NullString
			user_name sql.NullString
		)
		for rows.Next() {
			if err = rows.Scan(&user_id, &join_date, &user_name); err != nil {
				return nil, err
			}
			campInfo.MemberHeap[int(user_id.Int64)] = &Member{user_name.String, join_date.String, []int{}, []int{}}
			campInfo.MemberGroupByJoinDay[join_date.String] = append(campInfo.MemberGroupByJoinDay[join_date.String], campInfo.MemberHeap[int(user_id.Int64)])
		}

	}

	if rows, err := stmt_camp_food.Query(id); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			food_id       sql.NullInt64
			food_name     sql.NullString
			food_date     sql.NullString
			food_sub_id   sql.NullInt64
			food_sub_name sql.NullString
			bring_user_id sql.NullInt64
		)
		for rows.Next() {
			if err := rows.Scan(&food_id, &food_name, &food_date, &food_sub_id, &food_sub_name, &bring_user_id); err != nil {
				return nil, err
			}
			if _, ok := campInfo.FoodHeap[int(food_id.Int64)]; !ok {
				campInfo.FoodHeap[int(food_id.Int64)] = &Food{int(food_id.Int64), food_name.String, food_date.String, make(map[int]Item)}
				campInfo.FoodGroupByDay[food_date.String] = append(campInfo.FoodGroupByDay[food_date.String], campInfo.FoodHeap[int(food_id.Int64)])
			}
			campInfo.FoodHeap[int(food_id.Int64)].Ingredients[int(food_sub_id.Int64)] = Item{int(food_sub_id.Int64), food_sub_name.String, int(bring_user_id.Int64), food_date.String}
			if bring_user_id.Valid {
				campInfo.MemberHeap[int(bring_user_id.Int64)].Food = append(campInfo.MemberHeap[int(bring_user_id.Int64)].Food, int(food_sub_id.Int64))
			}
		}
	}

	if rows, err := stmt_camp_equipment.Query(id); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			equipment_id   sql.NullInt64
			equipment_name sql.NullString
			bring_user_id  sql.NullInt64
			date           sql.NullString
		)
		for rows.Next() {
			if err = rows.Scan(&equipment_id, &equipment_name, &bring_user_id, &date); err != nil {
				return nil, err
			}
			campInfo.EquipmentHeap[int(equipment_id.Int64)] = &Item{int(equipment_id.Int64), equipment_name.String, int(bring_user_id.Int64), date.String}
			if bring_user_id.Valid {
				campInfo.MemberHeap[int(bring_user_id.Int64)].Equipment = append(campInfo.MemberHeap[int(bring_user_id.Int64)].Equipment, int(equipment_id.Int64))
			}
			campInfo.EquipmentGroupByDay[date.String] = append(campInfo.EquipmentGroupByDay[date.String], campInfo.EquipmentHeap[int(equipment_id.Int64)])
		}
	}

	// for _, v := range campInfo.MemberHeap {
	// 	campInfo.MemberGroupByJoinDay[v.JoinDate] = append(campInfo.MemberGroupByJoinDay[v.JoinDate], v)
	// }

	return campInfo, nil
}

func Join(campId int, user *tgbotapi.User, join_date string) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt_insert_camp_member, err := tx.Prepare(`insert into camp_member (camp_id, user_id, join_date) values ($1, $2, $3)
	ON CONFLICT (camp_id, user_id) DO UPDATE 
	SET join_date = excluded.join_date`)
	if err != nil {
		return err
	}
	defer stmt_insert_camp_member.Close()

	if _, err = stmt_insert_camp_member.Exec(campId, user.ID, join_date); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}

func Quit(campId int, user *tgbotapi.User) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt_delete_camp, err := tx.Prepare(`delete from camp_member where camp_id = $1 and user_id = $2`)
	if err != nil {
		return err
	}
	defer stmt_delete_camp.Close()

	stmt_delete_bring, err := tx.Prepare(`delete from camp_user_bring where camp_id = $1 and user_id = $2`)
	if err != nil {
		return err
	}
	defer stmt_delete_bring.Close()

	if _, err = stmt_delete_camp.Exec(campId, user.ID); err != nil {
		return err
	}

	if _, err = stmt_delete_bring.Exec(campId, user.ID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}

func AddEquipment(campId int, date string, equipmentList []string, user *tgbotapi.User) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt_insert, err := tx.Prepare(`insert into camp_equipment (camp_id, name, date, created_by) values ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}
	defer stmt_insert.Close()

	for _, v := range equipmentList {
		if _, err = stmt_insert.Exec(campId, v, date, user.ID); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}

func AddFood(campId int, date string, food_name string, ingredients []string, user *tgbotapi.User) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt_insert_food, err := tx.Prepare(`insert into camp_food (camp_id, name, date, created_by) values ($1, $2, TO_DATE($3,'YYYY-MM-DD'), $4) RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt_insert_food.Close()

	stmt_insert_ingredients, err := tx.Prepare(`insert into camp_food_ingredients (camp_id, food_id, name) values ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt_insert_ingredients.Close()
	var (
		newFoodId sql.NullInt64
	)

	if err = stmt_insert_food.QueryRow(campId, food_name, date, user.ID).Scan(&newFoodId); err != nil {
		return err
	}
	for _, v := range ingredients {
		if _, err = stmt_insert_ingredients.Exec(campId, newFoodId, v); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}

func BringItem(campId int, userId int, item_id string, _type int) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt_insert, err := tx.Prepare(`insert into camp_user_bring (user_id, camp_id, type, item_id) values ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}
	defer stmt_insert.Close()

	if _, err = stmt_insert.Exec(userId, campId, _type, item_id); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}

func DropItem(campId int, userId int, item_id string, _type int) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt_insert, err := tx.Prepare(`delete from camp_user_bring where user_id = $1 and camp_id = $2 and type = $3 and item_id = $4`)
	if err != nil {
		return err
	}
	defer stmt_insert.Close()

	if _, err = stmt_insert.Exec(userId, campId, _type, item_id); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}

func AddEditUserName(userId int, username string) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt_insert_group_user, err := tx.Prepare(`insert into group_user (id, user_name) values ($1, $2) 
	ON CONFLICT (id) DO UPDATE 
	SET id = excluded.id, 
	user_name = excluded.user_name`)
	if err != nil {
		return err
	}
	defer stmt_insert_group_user.Close()

	if _, err = stmt_insert_group_user.Exec(userId, username); err != nil {
		return err
	}

	return tx.Commit()
}

func CheckUserExist(userId int) error {
	stmt_select_user, err := pgDb.Prepare(`select count(*) from group_user where id = $1`)
	if err != nil {
		return err
	}
	defer stmt_select_user.Close()

	var count sql.NullInt64

	if err = stmt_select_user.QueryRow(userId).Scan(&count); err != nil {
		return err
	} else if count.Int64 != 1 {
		return errors.New("set name first")
	}
	return nil
}

func UpdateCampName(name string, campId int) error {
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	stmt_update_came_name, err := tx.Prepare(`update camp set name = $1 where id = $2`)
	if err != nil {
		return err
	}
	if _, err = stmt_update_came_name.Exec(name, campId); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}
