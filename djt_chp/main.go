package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	"github.com/sirupsen/logrus"
)

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("djt_chp").Name("毒鸡汤与彩虹屁").Description("毒鸡汤与彩虹屁").Help("使用.djt 与 .chp触发命令").
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				if ok && v.Text == ".djt" {
					logrus.Infof("%v", v)
					name := req.Sender.Card
					if name == "" {
						name = req.Sender.Nickname
					}
					str, _ := getDujitang()
					cli.SendGroupMsg(
						&model.GroupMsg{
							GroupId: req.GroupId,
							Message: []*model.MessageSegment{
								{Type: "text", Data: &model.MessageElementText{
									Text: fmt.Sprintf("for  %v:%v", name, str),
								}},
							},
						},
					)
				} else if ok && v.Text == ".chp" {
					logrus.Infof("%v", v)
					name := req.Sender.Card
					if name == "" {
						name = req.Sender.Nickname
					}
					str, _ := getCaihongpi()
					cli.SendGroupMsg(
						&model.GroupMsg{
							GroupId: req.GroupId,
							Message: []*model.MessageSegment{
								{Type: "text", Data: &model.MessageElementText{
									Text: fmt.Sprintf("for  %v:%v", name, str),
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

var djtUrl = "https://api.muxiaoguo.cn/api/dujitang?api_key=%v"

type resp struct {
	Data *data  `json:"data"`
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type data struct {
	Comment string `json:"comment"`
}

func getDujitang() (string, error) {
	key, err := getDujitangKey()
	if err != nil {
		return "", err
	}
	r, err := http.DefaultClient.Get(fmt.Sprintf(djtUrl, key))
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	var resp resp
	err = json.Unmarshal(robots, &resp)
	if err != nil {
		return "", err
	}
	if resp.Code != "200" {
		return "", errors.New(resp.Msg)
	}
	return resp.Data.Comment, nil
}

func getDujitangKey() (string, error) {
	str := os.Getenv("BOT_DJT_KEY")
	if str == "" {
		return "", errors.New("未配置毒鸡汤的key")
	}
	return str, nil
}

var chpUrl = "https://api.muxiaoguo.cn/api/caihongpi?api_key=%v"

func getCaihongpi() (string, error) {
	key, err := getChpKey()
	if err != nil {
		return "", err
	}
	r, err := http.DefaultClient.Get(fmt.Sprintf(chpUrl, key))
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	var resp resp
	err = json.Unmarshal(robots, &resp)
	if err != nil {
		return "", err
	}
	if resp.Code != "200" {
		return "", errors.New(resp.Msg)
	}
	return resp.Data.Comment, nil
}

func getChpKey() (string, error) {
	str := os.Getenv("BOT_CHP_KEY")
	if str == "" {
		return "", errors.New("未配置毒鸡汤的key")
	}
	return str, nil
}
