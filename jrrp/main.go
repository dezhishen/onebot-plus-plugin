package main

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	"github.com/fogleman/gg"
	"github.com/sirupsen/logrus"
)

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("jrrp").Name("今日人品").Description("今日人品").Help("使用.jrrp，或者戳一戳触发命令").
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				if ok && v.Text == ".jrrp" {
					name := req.Sender.Card
					if name == "" {
						name = req.Sender.Nickname
					}
					cli.SendGroupMsg(genMsg(name, req.Sender.UserId, req.GroupId))
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
			cli.SendGroupMsg(genMsg(name, req.UserId, req.GroupId))
			return nil
		}).
		//构建插件
		Build().
		//启动
		Start()
}

func genMsg(name string, userId, groupId int64) *model.GroupMsg {
	buf, err := getImage(userId)
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

func getImage(id int64) ([]byte, error) {
	timeNow := time.Now().Local()
	path := getFileName(id, timeNow)
	b, err := getFile(id, path)
	if err != nil {
		return nil, err
	}
	if b != nil {
		return b, err
	}
	err = randomFile(timeNow, "jrrp", id, true, path)
	if err != nil {
		return nil, err
	}
	return getFile(id, path)
}

func getFileName(id int64, t time.Time) string {
	return fmt.Sprintf("./jrrp/%v-%v.png", id, timeToStr(t))
}

func getFile(id int64, path string) ([]byte, error) {
	// url = strings.Replace(url, "large", "original", -1)
	exists, _ := pathExists(path)
	if exists {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		return content, err
	}
	return nil, nil
}

func randomFile(t time.Time, pid string, uid int64, genIfNil bool, path string) error {
	r, err := getScore(t, pid, uid, true)
	if err != nil {
		return err
	}
	var score = r / 20
	full := "★"
	empty := "☆"
	var text string
	for i := 1; i <= score; i++ {
		text += full
	}
	for i := 1; i <= 5-score; i++ {
		text += empty
	}
	err = CreatImage(text, path)
	if err != nil {
		return err
	}
	return err
}

func getScore(t time.Time, pid string, uid int64, genIfNil bool) (int, error) {
	rand.Seed(time.Now().UnixNano())
	score := rand.Intn(100) + 1
	return score, nil
}

func timeToStr(t time.Time) string {
	return fmt.Sprintf("%v-%v-%v", t.Year(), int(t.Month()), t.Day())
}

func init() {
	exists, _ := pathExists("./jrrp")
	if !exists {
		os.Mkdir("./jrrp", 0777)
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreatImage(text string, path string) error {
	//图片的宽度
	var srcWidth float64 = 100
	//图片的高度
	var srcHeight float64 = 100
	dc := gg.NewContext(int(srcWidth), int(srcHeight))
	//设置背景色
	dc.SetColor(color.White)
	dc.Clear()
	dc.SetRGB255(255, 0, 0)
	if err := dc.LoadFontFace("./fortune/sakura.ttf", 25); err != nil {
		return err
	}
	sWidth, sHeight := dc.MeasureString(text)
	dc.DrawString(text, (srcWidth-sWidth)/2, (srcHeight-sHeight)/2)
	err := dc.SavePNG(path)
	if err != nil {
		return err
	}
	return nil
}
