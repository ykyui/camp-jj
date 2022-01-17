package database

import "database/sql"

func CreateCamp(ru RangeUnit) {
	tx, err := pgDb.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`insert into camp (id, start_date, end_date) 
	select coalesce(max(id),0)+1,TO_DATE($1,'YYYYMMDD'),TO_DATE($2,'YYYYMMDD') from camp`)
	if err != nil {
		return
	}
	defer stmt.Close()
	if _, err = stmt.Exec(ru.Start, ru.End); err != nil {
		return
	}
	tx.Commit()
}

type CampInfo struct {
	Id        int
	RangeUnit RangeUnit
}

func GetCampList() (result map[int]*CampInfo, err error) {
	stmt, err := pgDb.Prepare(`select id, TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD') from camp where current_date - INTERVAL '7 day' < end_date`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result = make(map[int]*CampInfo)
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
			result[int(id.Int64)] = &CampInfo{int(id.Int64), RangeUnit{start.String, end.String}}
		}
	}
	return result, nil
}
