package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	"github.com/sirupsen/logrus"
)

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("random").Name("骰子").Description("骰子插件").Help("使用.r触发命令").
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				logrus.Infof("%v", v)
				if ok && v.Text == ".r" {
					logrus.Infof("%v", v)
					name := req.Sender.Card
					if name == "" {
						name = req.Sender.Nickname
					}
					rand.Seed(time.Now().UnixNano())
					cli.SendGroupMsg(
						&model.GroupMsg{
							GroupId: req.GroupId,
							Message: []*model.MessageSegment{
								{Type: "text", Data: &model.MessageElementText{
									Text: fmt.Sprintf("[%v]掷出[%v]", name, rand.Intn(100)),
								}},
							},
						},
					)
				}
			}
			return nil
		}).
		//构建插件
		Build().
		//启动
		Start()
}
