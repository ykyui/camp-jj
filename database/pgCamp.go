package database

import (
	"database/sql"
	"errors"
	"fmt"

	"regexp"

	"github.com/ykyui/camp-jj/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateCamp(ru service.RangeUnit) error {
	dateReg, _ := regexp.Compile(`^(19[0-9]{2}|2[0-9]{3})(0[1-9]|1[012])([123]0|[012][1-9]|31)$`)
	if !dateReg.Match([]byte(ru.Start)) || !dateReg.Match([]byte(ru.End)) {
		return errors.New("dateFormatError")
	}
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`insert into camp (id, start_date, end_date) 
	select coalesce(max(id),0)+1,TO_DATE($1,'YYYYMMDD'),TO_DATE($2,'YYYYMMDD') from camp`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(ru.Start, ru.End); err != nil {
		return err
	}
	return tx.Commit()
}

type CampInfo struct {
	Id            int
	RangeUnit     service.RangeUnit
	FoodHeap      map[int]Food
	EquipmentHeap map[int]Item
	MemberHeap    map[int]*Member
}

type Food struct {
	Id       int
	Name     string
	Material map[int]Item
}

type Item struct {
	Name     string
	WhoBring int
}

type Member struct {
	Name      string
	JoinDate  string
	Food      []int
	Equipment []int
}

func (c *CampInfo) ToMsg() (result string) {
	result = fmt.Sprintf("campId: %d\ndate: %s To %s\n", c.Id, c.RangeUnit.Start, c.RangeUnit.End)

	result += "member\n"
	for _, v := range c.MemberHeap {
		result += fmt.Sprintf("name: %s %s\n", v.Name, v.JoinDate)
	}

	result += "food\n"
	for _, v := range c.FoodHeap {
		result += fmt.Sprintf("name: %s\n", v.Name)
		for _, v := range v.Material {
			var bringName string
			if bring, ok := c.MemberHeap[v.WhoBring]; ok {
				bringName = bring.Name
			}
			result += fmt.Sprintf("material: %s %s\n", v.Name, bringName)
		}
	}

	result += "equipment \n"
	for _, v := range c.EquipmentHeap {
		var bringName string
		if bring, ok := c.MemberHeap[v.WhoBring]; ok {
			bringName = bring.Name
		}
		result += fmt.Sprintf(`
		name: %s %s
		`, v.Name, bringName)
	}
	return
}

func GetCampList() (result []*CampInfo, err error) {
	stmt, err := pgDb.Prepare(`select id, TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD') from camp where current_date - INTERVAL '7 day' < end_date`)
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
				id    sql.NullInt64
				start sql.NullString
				end   sql.NullString
			)
			if err = rows.Scan(&id, &start, &end); err != nil {
				return nil, err
			}
			result = append(result, &CampInfo{int(id.Int64), service.RangeUnit{Start: start.String, End: end.String}, nil, nil, nil})
		}
	}
	return result, nil
}

func getCampInfo(id int) (*CampInfo, error) {
	stmt_camp_date, err := pgDb.Prepare(`select TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD') from camp where id = $1`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_date.Close()

	stmt_camp_user, err := pgDb.Prepare(`select user_id, TO_CHAR(join_date, 'YYYY-MM-DD'), user_name from camp_member left join group_user on user_id = id where camp_id = $1`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_user.Close()

	stmt_camp_food, err := pgDb.Prepare(`select cf.id, cf.name, cfm.id, cfm.name, cub.user_id
	from camp_food cf 
	left join camp_food_meterial cfm 
		left join camp_user_bring cub on cfm.id = cub.item_id and cub.camp_id = cfm.camp_id and cub.type = 1 
	on cf.id = cfm.food_id and cf.camp_id = cfm.camp_id
	where cf.camp_id = $1`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_food.Close()

	stmt_camp_equipment, err := pgDb.Prepare(`select ce.id, ce.name, cub.user_id 
	from camp_equipment ce 
	left join camp_user_bring cub on  ce.id = cub.item_id and ce.camp_id = cub.camp_id and cub.type = 2
	where ce.camp_id = $1`)
	if err != nil {
		return nil, err
	}
	defer stmt_camp_equipment.Close()

	var (
		start_date sql.NullString
		end_date   sql.NullString
	)
	if err = stmt_camp_date.QueryRow(id).Scan(&start_date, &end_date); err != nil {
		return nil, err
	}
	campInfo := &CampInfo{id, service.RangeUnit{Start: start_date.String, End: end_date.String}, make(map[int]Food), make(map[int]Item), make(map[int]*Member)}

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
		}

	}

	if rows, err := stmt_camp_food.Query(id); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			food_id       sql.NullInt64
			food_name     sql.NullString
			food_sub_id   sql.NullInt64
			food_sub_name sql.NullString
			bring_user_id sql.NullInt64
		)
		for rows.Next() {
			if err := rows.Scan(&food_id, &food_name, &food_sub_id, &food_sub_name, &bring_user_id); err != nil {
				return nil, err
			}
			if _, ok := campInfo.FoodHeap[int(food_id.Int64)]; !ok {
				campInfo.FoodHeap[int(food_id.Int64)] = Food{int(food_id.Int64), food_name.String, make(map[int]Item)}
			}
			campInfo.FoodHeap[int(food_id.Int64)].Material[int(food_sub_id.Int64)] = Item{food_sub_name.String, int(bring_user_id.Int64)}
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
		)
		for rows.Next() {
			if err = rows.Scan(&equipment_id, &equipment_name, &bring_user_id); err != nil {
				return nil, err
			}
			campInfo.EquipmentHeap[int(equipment_id.Int64)] = Item{equipment_name.String, int(bring_user_id.Int64)}
			if bring_user_id.Valid {
				campInfo.MemberHeap[int(bring_user_id.Int64)].Equipment = append(campInfo.MemberHeap[int(bring_user_id.Int64)].Food, int(equipment_id.Int64))
			}
		}
	}

	return campInfo, nil
}

func Join(campId int, user *tgbotapi.User, join_date string) error {
	dateReg, _ := regexp.Compile(`^(19[0-9]{2}|2[0-9]{3})(0[1-9]|1[012])([123]0|[012][1-9]|31)$`)
	if !dateReg.Match([]byte(join_date)) {
		return errors.New("dateFormatError")
	}
	tx, err := pgDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt_check_input, err := tx.Prepare(`select count(*) from camp where id = $1 and TO_DATE($2,'YYYYMMDD') between start_date and end_date`)
	if err != nil {
		return err
	}
	defer stmt_check_input.Close()
	stmt_insert_group_user, err := tx.Prepare(`insert into group_user (id, user_name) values ($1, $2) 
	ON CONFLICT (id) DO UPDATE 
	SET id = excluded.id, 
	user_name = excluded.user_name`)
	if err != nil {
		return err
	}
	defer stmt_insert_group_user.Close()

	stmt_insert_camp_member, err := tx.Prepare(`insert into camp_member (camp_id, user_id, join_date) values ($1, $2, $3)
	ON CONFLICT (camp_id, user_id) DO UPDATE 
	SET join_date = excluded.join_date`)
	if err != nil {
		return err
	}
	defer stmt_insert_camp_member.Close()

	var count sql.NullInt64
	if err = stmt_check_input.QueryRow(campId, join_date).Scan(&count); err != nil {
		return err
	} else if count.Int64 != 1 {
		return errors.New("not within")
	}

	if _, err = stmt_insert_group_user.Exec(user.ID, user.UserName); err != nil {
		return err
	}
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

	stmt_delete, err := tx.Prepare(`delete from camp_member where camp_id = $1 and user_id = $2`)
	if err != nil {
		return err
	}
	defer stmt_delete.Close()

	if _, err = stmt_delete.Exec(campId, user.ID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	updateCampInfo(campId)
	return nil
}
