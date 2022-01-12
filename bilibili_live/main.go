package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dezhishen/onebot-plus-plugin/pkg/command"
	"github.com/dezhishen/onebot-plus-plugin/pkg/common"
	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	"github.com/miRemid/danmagu"
	"github.com/miRemid/danmagu/message"
	"github.com/sirupsen/logrus"
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
						for i, id := range res {
							if i == 0 {
								continue
							}
							_id, err := strconv.Atoi(id)
							if err != nil {
								cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, fmt.Sprintf("%v", err)))
								return nil
							}
							addListen(uint32(_id), req.GroupId)
						}
						cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, "成功订阅"))
					} else if line.Event == "remove" {
						for i, id := range res {
							if i == 0 {
								continue
							}
							_id, err := strconv.Atoi(id)
							if err != nil {
								cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, fmt.Sprintf("%v", err)))
								return nil
							}
							removeListen(uint32(_id), req.GroupId)
						}
						cli.SendGroupMsg(common.GenGroupTextMsg(req.GroupId, "成功取消"))
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
				break
			}
		}
		if flag {
			liveGroupIds[liveId] = append(groupIds, groupId)
			err := insertLiveGroup(liveId, groupId)
			if err != nil {
				logrus.Errorf("insertLiveGroup err %v", err)
			}
		}
	} else {
		liveGroupIds[liveId] = []int64{groupId}
		err := insertLiveGroup(liveId, groupId)
		if err != nil {
			logrus.Errorf("insertLiveGroup err %v", err)
		}
	}
	if _, ok := clis[liveId]; !ok {
		clis[liveId] = newListenCli(liveId)
		go clis[liveId].Listen()
		err := insertLive(liveId, "")
		if err != nil {
			logrus.Errorf("insertLive err %v", err)
		}
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
				msg := common.GenGroupTextMsg(groupId, fmt.Sprintf("开播啦\nhttps://live.bilibili.com/%v", live.Roomid))
				_onebotCli.SendGroupMsg(msg)
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
	cli.Handler(message.DANMU_MSG, func(ctx context.Context, msg message.Danmaku) {
	})
	cli.Handler(message.RQZ, func(ctx context.Context, rqz uint32) {
	})
	cli.Handler(message.DEFAULT, func(c1 context.Context, c2 *message.Context) {
	})
	cli.Handler(message.SEND_GIFT, func(ctx context.Context, gift message.SendGift) {
	})
	cli.Handler(message.INTERACT_WORD, func(ctx context.Context, word message.InteractWord) {
	})
	cli.Handler(message.ONLINE_RANK_V2, func(ctx context.Context, rankV2 message.OnlineRankV2) {
	})
	cli.Handler(message.ONLINE_RANK_TOP3, func(ctx context.Context, rankTOP3 message.OnlineRankTOP3) {
	})
	cli.Handler(message.COMBO_SEND, func(ctx context.Context, combo message.ComboSend) {
	})
	cli.Handler(message.WIDGET_BANNER, func(ctx context.Context, banner message.WidgetBanner) {
	})
	cli.Handler(message.ENTRY_EFFECT, func(ctx context.Context, entry message.EntryEffect) {
	})
	cli.Handler(message.ONLINE_RANK_COUNT, func(ctx context.Context, count message.OnlineRankCount) {
	})
	cli.Handler(message.ROOM_RANK, func(ctx context.Context, rank message.RoomRank) {
	})
	return cli
}
