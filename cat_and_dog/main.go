package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
)

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("cat_and_dog").Name("猫猫图和狗狗图").Description("猫猫图和狗狗图").Help("来点猫猫图/来点狗狗图").
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				if !ok {
					return nil
				}
				if v.Text == ".cat" || v.Text == ".thecat" || v.Text == "来点猫猫图" {
					b, e := getCatPic()
					if e != nil {
						cli.SendGroupMsg(
							&model.GroupMsg{
								GroupId: req.GroupId,
								Message: []*model.MessageSegment{
									{Type: "text", Data: &model.MessageElementText{
										Text: fmt.Sprintf("获取猫猫图发生异常,%v", e),
									}},
								},
							},
						)
						return nil
					}
					cli.SendGroupMsg(genPicMsg(req.GroupId, b))
				} else if v.Text == ".dog" || v.Text == ".thedog" || v.Text == "来点狗狗图" {
					b, e := getDogPic()
					if e != nil {
						cli.SendGroupMsg(
							&model.GroupMsg{
								GroupId: req.GroupId,
								Message: []*model.MessageSegment{
									{Type: "text", Data: &model.MessageElementText{
										Text: fmt.Sprintf("获取狗狗图发生异常,%v", e),
									}},
								},
							},
						)
						return nil
					}
					cli.SendGroupMsg(genPicMsg(req.GroupId, b))
				}

			}
			return nil
		}).
		//构建插件
		Build().
		//启动
		Start()
}

type Data struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

func genPicMsg(groupId int64, buf []byte) *model.GroupMsg {
	return &model.GroupMsg{
		GroupId: groupId,
		Message: []*model.MessageSegment{
			{Type: "image", Data: &model.MessageElementImage{
				ImageType: "",
				File:      "base64://" + base64.StdEncoding.EncodeToString(buf),
			}},
		},
	}
}

var catUrl = "https://api.thecatapi.com/v1/images/search"

func getCatPic() ([]byte, error) {

	r, err := http.DefaultClient.Get(catUrl)
	if err != nil {
		return nil, err
	}
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	resp := string(robots)
	if resp == "" {
		return nil, errors.New("请稍后重试")
	}

	var datas []Data
	err = json.Unmarshal(robots, &datas)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, errors.New("请稍后重试")
	}
	imageResp, err := http.DefaultClient.Get(datas[0].Url)
	if err != nil {
		return nil, err
	}
	imageBytes, err := ioutil.ReadAll(imageResp.Body)
	if err != nil {
		return nil, err
	}
	imageResp.Body.Close()
	return imageBytes, nil
}

var dogUrl = "https://api.thedogapi.com/v1/images/search"

func getDogPic() ([]byte, error) {
	r, err := http.DefaultClient.Get(dogUrl)
	if err != nil {
		return nil, err
	}
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	resp := string(robots)
	if resp == "" {
		return nil, errors.New("请稍后重试")
	}

	var datas []Data
	err = json.Unmarshal(robots, &datas)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, errors.New("请稍后重试")
	}
	imageResp, err := http.DefaultClient.Get(datas[0].Url)
	if err != nil {
		return nil, err
	}
	imageBytes, err := ioutil.ReadAll(imageResp.Body)
	if err != nil {
		return nil, err
	}
	imageResp.Body.Close()
	return imageBytes, nil
}
