package main

import (
	"os"
	"testing"
)

// All tests in the same tests because I need state and it is easier
func TestAllPersistence(t *testing.T) {
	const dbpath = "foo.db"
	var readItems []RoomInfo
	var (
		readSingleItem RoomInfo
		err            error
	)

	// test 0 - init db
	db := InitDB(dbpath)
	defer db.Close()
	defer os.Remove(dbpath)

	// test 1 - create tables
	CreateTables(db)
	t.Log("TestAllPersistence - TEST1 - CreateTables - OK")

	// test 2 - insert data
	StoreItem(db, RoomInfo{1, "Room1", "CurrTitle", "CurrSpeaker", "CurrTime", "NextTitle", "NextSpeaker", "NextTime"})

	readItems = ReadRoomInfoTable(db)
	if 1 != len(readItems) {
		t.Errorf("Number of rows did not match the expected number 1")
	}

	// test 2.1 - insert more data
	StoreItem(db, RoomInfo{2, "Room2", "CurrTitle", "CurrSpeaker", "CurrTime", "NextTitle", "NextSpeaker", "NextTime"})
	StoreItem(db, RoomInfo{3, "Room4", "CurrTitle", "CurrSpeaker", "CurrTime", "NextTitle", "NextSpeaker", "NextTime"})
	StoreItem(db, RoomInfo{4, "Room4", "CurrTitle", "CurrSpeaker", "CurrTime", "NextTitle", "NextSpeaker", "NextTime"})

	readItems = ReadRoomInfoTable(db)
	if 4 != len(readItems) {
		t.Errorf("Number of rows did not match the expected number 1")
	}
	t.Log("TestAllPersistence - TEST2 - insert data - ", len(readItems), "(<-- 4 is expected)")

	// test 3 - retrieve single data
	readSingleItem, err = ReadRoomInfo(db, 2)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if "Room2" != readSingleItem.RoomName {
		t.Errorf("Unexpected RoomName: %s", readSingleItem.RoomName)
	}
	t.Log("TestAllPersistence - TEST3 - retrieve single data - ", readSingleItem)

	// test 4 - retrieve unexisting data
	readSingleItem, err = ReadRoomInfo(db, 99)
	if err != nil {
		// simple error is nice here.
		if err.Error() != "sql: no rows in result set" {
			t.Error("Unexpected error: ", err)
		}
	}
	t.Log("TestAllPersistence - TEST4 - retrieve unexisting data - ", err)

	// test 5 - update row
	StoreItem(db, RoomInfo{2, "Room2", "Another Title Here", "", "", "", "", ""})
	readSingleItem, err = ReadRoomInfo(db, 2)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if "Another Title Here" != readSingleItem.CurrentTitle {
		t.Errorf("Unexpected RoomName: %s", readSingleItem.RoomName)
	}
	t.Log("TestAllPersistence - TEST5 - update row - ", readSingleItem)

}
