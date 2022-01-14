package database

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
