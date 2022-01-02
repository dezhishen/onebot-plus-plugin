package main

import (
	"encoding/base64"
	"fmt"

	"github.com/dezhishen/onebot-plus-plugin/today_fortune/fortune"
	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	"github.com/sirupsen/logrus"
)

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("today_fortune").Name("今日运势").Description("今日运势").Help("使用.tf，或者戳一戳触发命令").
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				if ok && v.Text == ".tf" {
					name := req.Sender.Card
					if name == "" {
						name = req.Sender.Nickname
					}
					cli.SendGroupMsg(genMsg(name, req.GroupId))
				}
			}
			return nil
		}).
		NoticeGroupNotifyPoke(func(req *model.EventNoticeGroupNotifyPoke, cli cli.OnebotCli) error {
			if req.TargetId != req.SelfId {
				return nil
			}
			info, err := cli.GetGroupMemberInfo(req.GroupId, req.UserId, false)
			if err != nil {
				return nil
			}
			name := info.Data.Card
			if name == "" {
				name = info.Data.Nickname
			}
			cli.SendGroupMsg(genMsg(name, req.GroupId))
			return nil
		}).
		//构建插件
		Build().
		//启动
		Start()
}

func genMsg(name string, groupId int64) *model.GroupMsg {
	buf, err := fortune.GetPic()
	if err != nil {
		logrus.Errorf("运签发生异常,%v", err)
	}
	file := "base64://" + base64.StdEncoding.EncodeToString(buf)
	logrus.Info(file)
	return &model.GroupMsg{
		GroupId: groupId,
		Message: []*model.MessageSegment{
			{
				Type: "text",
				Data: &model.MessageElementText{
					Text: fmt.Sprintf("for:%v", name),
				},
			},
			{
				Type: "image",
				Data: &model.MessageElementImage{
					ImageType: "",
					File:      file,
				}},
		},
	}
}
