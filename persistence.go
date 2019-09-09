package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Represents the current state of a RoomInfo.
type RoomInfo struct {
	ID             int // room number
	RoomName       string
	CurrentTitle   string
	CurrentSpeaker string
	CurrentTime    string
	NextTitle      string
	NextSpeaker    string
	NextTime       string
}

// InitDB creates a new DB object using filename as parameter
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

// CreateTables create all tables needed.
func CreateTables(db *sql.DB) {
	// create table if not exists
	sqlTable := `
	CREATE TABLE IF NOT EXISTS room_info(
		id INTEGER PRIMARY KEY,
		room_name TEXT NOT NULL,
		current_title TEXT NOT NULL,
		current_speaker TEXT NOT NULL,
		current_time TEXT NOT NULL,
		next_title TEXT NOT NULL,
		next_speaker TEXT NOT NULL,
		next_time TEXT NOT NULL
	);
	`

	_, err := db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}
}

// StoreItem stores multiple items
func StoreItem(db *sql.DB, item RoomInfo) {
	sqlAdditem := `
	INSERT OR REPLACE INTO room_info(
		id,
		room_name,
		current_title,
		current_speaker,
		current_time,
		next_title,
		next_speaker,
		next_time
	) values(?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sqlAdditem)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.ID, item.RoomName,
		item.CurrentTitle, item.CurrentSpeaker, item.CurrentTime,
		item.NextTitle, item.NextSpeaker, item.NextTime)
	if err != nil {
		panic(err)

	}
}

// ReadRoomInfoTable reads all rows in the room_info table
func ReadRoomInfoTable(db *sql.DB) []RoomInfo {
	sqlReadall := `
	SELECT * FROM room_info
	ORDER BY id DESC
	`

	rows, err := db.Query(sqlReadall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []RoomInfo
	for rows.Next() {
		item := RoomInfo{}
		err2 := rows.Scan(&item.ID, &item.RoomName,
			&item.CurrentTitle, &item.CurrentSpeaker, &item.CurrentTime,
			&item.NextTitle, &item.NextSpeaker, &item.NextTime)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}

// ReadRoomInfo read one row, with the row ID
func ReadRoomInfo(db *sql.DB, id int) (RoomInfo, error) {
	sqlRead := `
	SELECT * FROM room_info
	WHERE id = ?
	`

	stmt, err := db.Prepare(sqlRead)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var item RoomInfo
	err = stmt.QueryRow(id).Scan(&item.ID, &item.RoomName,
		&item.CurrentTitle, &item.CurrentSpeaker, &item.CurrentTime,
		&item.NextTitle, &item.NextSpeaker, &item.NextTime)

	return item, err
}
