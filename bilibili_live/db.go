package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() {
	db := getDb()
	defer db.Close()
	db.Exec("create table if not exists bilibili_live ( id int NOT NULL,name CHAR(32) NOT NULL)")
	db.Exec("create table if not exists bilibili_live_group(id int NOT NULL,group_id int NOT NULL)")

}

// var _db *sql.DB
// var db_mtx sync.RWMutex

func getDb() *sql.DB {
	_db, err := sql.Open("sqlite3", "./bili-live.db?")
	if err != nil {
		panic(err)
	}
	return _db

}
func delLive(liveId uint32) error {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("delete from bilibili_live where id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId)
	return err
}

func delLiveGroup(liveId uint32, groupId int64) error {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("delete from bilibili_live_group where id=? and group_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId, groupId)
	return err
}

func insertLive(liveId uint32, name string) error {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO bilibili_live(id, name) values(?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId, name)
	return err
}

func insertLiveGroup(liveId uint32, groupId int64) error {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO bilibili_live_group(id, group_id) values(?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId, groupId)
	return err
}

func getAllLives() ([]uint32, error) {
	db := getDb()
	defer db.Close()
	rows, err := db.Query("select id from bilibili_live")
	if err != nil {
		return nil, err
	}
	var result []uint32
	for rows.Next() {
		var liveId uint32
		err = rows.Scan(&liveId)
		if err != nil {
			return nil, err
		}
		result = append(result, liveId)
	}
	return result, nil
}

func getGroupIdsByLiveId(liveId uint32) ([]int64, error) {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select group_id from bilibili_live_group where id = ? ")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(liveId)
	if err != nil {
		return nil, err
	}
	var result []int64
	for rows.Next() {
		var groupId int64
		err = rows.Scan(&groupId)
		if err != nil {
			return nil, err
		}
		result = append(result, groupId)
	}
	return result, nil
}
