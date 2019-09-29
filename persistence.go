package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

// RoomInfo represents the current state of a RoomInfo.
type RoomInfo struct {
	ID             int    `json:"room_id"` // room number
	RoomName       string `json:"room"`
	CurrentTitle   string `json:"title"`
	CurrentSpeaker string `json:"speaker"`
	CurrentTime    string `json:"time"`
	NextTitle      string `json:"n_title"`
	NextSpeaker    string `json:"n_speaker"`
	NextTime       string `json:"n_time"`
	AutoLoopSec    int    `json:"auto_loop_sec"`
}

// InitDB creates a new DB object using filename as parameter
func InitDB(filepath string) *bolt.DB {
	db, err := bolt.Open(filepath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Println("Error InitDB")
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}

	return db
}

// StoreItem stores multiple items
func StoreItem(db *bolt.DB, item RoomInfo) {

	roomInfoKey := strconv.Itoa(item.ID)
	roomInfoValue, err := json.Marshal(item)
	if err != nil {
		log.Println(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("room_info"))
		if err != nil {
			return err
		}
		return b.Put([]byte(roomInfoKey), []byte(roomInfoValue))
	})
}

// ReadRoomInfoTable reads all rows in the room_info table
func ReadRoomInfoTable(db *bolt.DB) (map[int]RoomInfo, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	roomInfoArray := make(map[int]RoomInfo)

	b := tx.Bucket([]byte("room_info"))
	c := b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		roomInfo := RoomInfo{}
		if err = json.Unmarshal(v, &roomInfo); err != nil {
			return nil, err
		}

		roomInfoArray[roomInfo.ID] = roomInfo
	}

	return roomInfoArray, nil

}

// ReadRoomInfo read one row, with the row ID
func ReadRoomInfo(db *bolt.DB, id int) (RoomInfo, error) {

	roomInfoKey := strconv.Itoa(id)

	tx, err := db.Begin(false)
	if err != nil {
		return RoomInfo{}, err
	}
	defer tx.Rollback()

	bkt := tx.Bucket([]byte("room_info"))
	if bkt == nil {
		return RoomInfo{}, errors.New("No room_info bucket available (database is empty)")
	}
	v := bkt.Get([]byte(roomInfoKey))

	var roomInfo RoomInfo
	if err = json.Unmarshal(v, &roomInfo); err != nil {
		return RoomInfo{}, err
	}

	return roomInfo, err
}
