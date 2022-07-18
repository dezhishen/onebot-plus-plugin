package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/dezhishen/onebot-plus-plugin/pkg/pixiv"
	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
	"github.com/sirupsen/logrus"
)

var mutex sync.Mutex

var allImages []*pixiv.PixivImage

func sendError(cli cli.OnebotCli, groupId int64, errMsg string) {
	cli.SendGroupMsg(
		&model.GroupMsg{
			GroupId: groupId,
			Message: []*model.MessageSegment{
				{Type: "text", Data: &model.MessageElementText{
					Text: errMsg,
				}},
			},
		},
	)
}

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("pixiv").Name("pixiv").Description("随机推荐Pixiv图片").Help("gkd").
		//Onebot回调事件
		MessageGroup(
			func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
				defer func() {
					if r := recover(); r != nil {
						logrus.Info("recover", r)
					}
				}()
				if len(req.Message) > 0 && req.Message[0].Type == "text" {
					v, ok := req.Message[0].Data.(*model.MessageElementText)
					if !ok {
						return nil
					}
					if v.Text == "gkd" || v.Text == "来点色图" || v.Text == "色图呢" {
						image, err := getPic()
						if err != nil {
							sendError(cli, req.GroupId, fmt.Sprintf("获取图片失败,错误,%v", err))
							return nil
						}
						buf, err := downloadImage(image)
						if err != nil {
							cli.SendGroupMsg(
								&model.GroupMsg{
									GroupId: req.GroupId,
									Message: []*model.MessageSegment{
										{Type: "text", Data: &model.MessageElementText{
											Text: fmt.Sprintf("获取图片失败,错误,%v", err),
										}},
									},
								},
							)
							return nil
						}
						if image.R18 {
							r, _ := cli.GetLoginInfo()
							resp, err := cli.SendGroupForwardMessageByRawMsg(req.GroupId, r.Data.UserId, r.Data.Nickname, []*model.MessageSegment{
								{
									Type: "image",
									Data: &model.MessageElementImage{
										ImageType: "",
										File:      "base64://" + base64.StdEncoding.EncodeToString(buf),
									},
								},
							})
							if err != nil {
								sendError(cli, req.GroupId, fmt.Sprintf("发送消息,错误,%v", err))
								return nil
							}
							if resp.Retcode != 0 {
								sendError(cli, req.GroupId, fmt.Sprintf("发送消息,错误,%v", "消息可能被风控"))
								return nil
							}
							time.Sleep(15 * time.Second)
							cli.DelMsg(resp.Data.MessageId)
						} else {
							_, err := cli.SendGroupMsg(&model.GroupMsg{
								GroupId: req.GroupId,
								Message: []*model.MessageSegment{
									{Type: "image", Data: &model.MessageElementImage{
										ImageType: "",
										File:      "base64://" + base64.StdEncoding.EncodeToString(buf),
									}},
								},
							})
							if err != nil {
								sendError(cli, req.GroupId, fmt.Sprintf("发送消息,错误,%v", err))
								return nil
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

// func getRawRandPic() ([]byte, error) {
// 	imageResp, err := http.DefaultClient.Get("https://pximg.rainchan.win/rawimg")
// 	if err != nil {
// 		return nil, err
// 	}
// 	imageBytes, err := ioutil.ReadAll(imageResp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	imageResp.Body.Close()
// 	return imageBytes, nil
// }

func downloadImage(p *pixiv.PixivImage) ([]byte, error) {
	imageResp, err := http.DefaultClient.Get(p.Urls.GetUrl())
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

func getPic() (*pixiv.PixivImage, error) {
	image, e := randomAImage()
	if e != nil {
		return nil, e
	}
	return image, nil
}

func randomAImage() (*pixiv.PixivImage, error) {
	mutex.Lock()
	defer mutex.Unlock()
	for i := 0; i < 10; i++ {
		if len(allImages) == 0 {
			e := resetImages()
			if e != nil {
				return nil, e
			}
		}
		result := allImages[0]
		allImages = allImages[1:]
		if result.Urls.GetUrl() != "" {
			return result, nil
		}

	}
	return nil, errors.New("cannot find image")
}

func resetImages() error {
	r, e := pixiv.RandomImgsWithRetry()
	if e != nil {
		return e
	}
	allImages = r
	return nil
}
