package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"database/sql"

	"github.com/dezhishen/onebot-plus-plugin/pkg/command"
	"github.com/dezhishen/onebot-plus-plugin/pkg/common"
	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miRemid/danmagu"
	"github.com/miRemid/danmagu/message"
)

type BiliReq struct {
	Event string `short:"e" long:"event" description:"事件类型" choice:"add" choice:"remove" default:"add"`
}

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("bilibili-live").Name("bilibili直播").Description("bilibili直播监听").Help(".bili-live -h").
		Init(func(cli cli.OnebotCli) error {
			initDB()
			intiOnebotCli(cli)
			startListen()
			return nil
		}).
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				if ok && strings.HasPrefix(v.Text, ".bili-live") {
					var line BiliReq
					res, err := command.ParseWithDescription(".bili-live", &line, strings.Split(v.Text, " "), "修改监听的配置")
					if err != nil {
						cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, fmt.Sprintf("%v", err)))
						return nil
					}
					if line.Event == "add" {
						for _, id := range res {
							_id, err := strconv.Atoi(id)
							if err != nil {
								cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, fmt.Sprintf("%v", err)))
								return nil
							}
							addListen(uint32(_id), req.GroupId)
						}
					} else if line.Event == "remove" {
						for _, id := range res {
							_id, err := strconv.Atoi(id)
							if err != nil {
								cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, fmt.Sprintf("%v", err)))
								return nil
							}
							removeListen(uint32(_id), req.GroupId)
						}
					}
				}
			}
			return nil
		}).
		//构建插件
		Build().
		//启动
		Start()
}

var clis = make(map[uint32]*danmagu.LiveClient)
var liveGroupIds = make(map[uint32][]int64)
var mux sync.Mutex

var _onebotCli cli.OnebotCli

func intiOnebotCli(onebotCli cli.OnebotCli) {
	_onebotCli = onebotCli
}
func startListen() error {
	mux.Lock()
	defer mux.Unlock()
	liveIds, err := getAllLives()
	if err != nil {
		return err
	}
	for _, liveId := range liveIds {
		groupIds, err := getGroupIdsByLiveId(liveId)
		if err != nil {
			return err
		}
		for _, groupId := range groupIds {
			if groupIds, ok := liveGroupIds[liveId]; ok {
				flag := true
				for _, e := range groupIds {
					if groupId == e {
						flag = false
					}
				}
				if flag {
					liveGroupIds[liveId] = append(groupIds, groupId)
				}
			} else {
				liveGroupIds[liveId] = []int64{groupId}
			}
			if _, ok := clis[liveId]; !ok {
				clis[liveId] = newListenCli(liveId)
				go clis[liveId].Listen()
			}
		}
	}
	return nil
}
func addListen(liveId uint32, groupId int64) {
	mux.Lock()
	defer mux.Unlock()
	if groupIds, ok := liveGroupIds[liveId]; ok {
		flag := true
		for _, e := range groupIds {
			if groupId == e {
				flag = false
			}
		}
		if flag {
			liveGroupIds[liveId] = append(groupIds, groupId)
			insertLiveGroup(liveId, groupId)
		}
	} else {
		liveGroupIds[liveId] = []int64{groupId}
	}
	if _, ok := clis[liveId]; !ok {
		clis[liveId] = newListenCli(liveId)
		go clis[liveId].Listen()
		insertLive(liveId, "")
	}
}

func removeListen(id uint32, groupId int64) {
	mux.Lock()
	defer mux.Unlock()
	if groupIds, ok := liveGroupIds[id]; ok {
		var newGIds []int64
		for _, gId := range groupIds {
			if gId != groupId {
				newGIds = append(newGIds, gId)
			}
		}
		liveGroupIds[id] = newGIds
		//删除
		delLiveGroup(id, groupId)
		if len(liveGroupIds[id]) != 0 {
			liveGroupIds[id] = groupIds
		} else {
			liveGroupIds[id] = nil
			clis[id].Close()
			clis[id] = nil
			delLive(id)

		}
	}
}

func newListenCli(id uint32) *danmagu.LiveClient {
	cli := danmagu.NewClient(id, &danmagu.ClientConfig{
		HeartBeatTime: 30,
	})
	cli.Handler(message.LIVE, func(ctx context.Context, live message.Live) {
		if groupIds, ok := liveGroupIds[id]; ok {
			for _, groupId := range groupIds {
				_onebotCli.SendGroupMsg(common.GenGroupTextMsg(groupId, fmt.Sprintf("%v开播啦", live.Roomid)))
			}
		}
	})
	cli.Handler(message.PREPARING, func(ctx context.Context, pre message.Preparing) {
		if groupIds, ok := liveGroupIds[id]; ok {
			for _, groupId := range groupIds {
				_onebotCli.SendGroupMsg(common.GenGroupTextMsg(groupId, fmt.Sprintf("%v下播啦", pre.RoomID)))
			}
		}
	})
	return cli
}

func initDB() {
	db := getDb()
	db.Exec("create table if not exists bilibili_live ( id int NOT NULL,name CHAR(32) NOT NULL)")
	db.Exec("create table if not exists bilibili_live_group(id int NOT NULL,group_id int NOT NULL)")
}

var _db *sql.DB
var db_mtx sync.RWMutex

func getDb() *sql.DB {
	db_mtx.RLock()
	if _db != nil {
		db_mtx.RUnlock()
		return _db
	}
	db_mtx.RUnlock()
	db_mtx.Lock()
	defer db_mtx.Unlock()
	if _db == nil {
		var err error
		_db, err = sql.Open("sqlite3", "./bili-live.db")
		if err != nil {
			panic(err)
		}
	}
	return _db

}
func delLive(liveId uint32) error {
	db := getDb()
	stmt, err := db.Prepare("delete from bilibili_live where id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId)
	return err
}

func delLiveGroup(liveId uint32, groupId int64) error {
	db := getDb()
	stmt, err := db.Prepare("delete from bilibili_live_group where id=? and group_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId, groupId)
	return err
}

func insertLive(liveId uint32, name string) error {
	db := getDb()
	stmt, err := db.Prepare("INSERT INTO bilibili_live(id, name) values(?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId, name)
	return err
}

func insertLiveGroup(liveId uint32, groupId int64) error {
	db := getDb()
	stmt, err := db.Prepare("INSERT INTO bilibili_live_group(id, group_id) values(?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(liveId, groupId)
	return err
}

func getAllLives() ([]uint32, error) {
	db := getDb()
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
