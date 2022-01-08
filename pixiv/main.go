package main

import (
	"fmt"
	"sync"

	"github.com/dezhishen/onebot-plus-plugin/pkg/common"
	"github.com/dezhishen/onebot-plus-plugin/pkg/pixiv"
	"github.com/dezhishen/onebot-plus/pkg/cli"
	"github.com/dezhishen/onebot-plus/pkg/plugin"
	"github.com/dezhishen/onebot-sdk/pkg/model"
)

var mutex sync.Mutex

var allImages []string

func main() {
	plugin.NewOnebotEventPluginBuilder().
		//设置插件内容
		Id("pixiv").Name("pixiv").Description("随机推荐Pixiv图片").Help("gkd").
		//Onebot回调事件
		MessageGroup(func(req *model.EventMessageGroup, cli cli.OnebotCli) error {
			if len(req.Message) > 0 && req.Message[0].Type == "text" {
				v, ok := req.Message[0].Data.(*model.MessageElementText)
				if !ok {
					return nil
				}
				if v.Text == "gkd" || v.Text == "来点色图" || v.Text == "色图呢" {
					b, e := getPic()
					if e != nil {
						cli.SendGroupMsg(
							&model.GroupMsg{
								GroupId: req.GroupId,
								Message: []*model.MessageSegment{
									{Type: "text", Data: &model.MessageElementText{
										Text: fmt.Sprintf("获取图片失败,错误,%v", e),
									}},
								},
							},
						)
						return nil
					}
					cli.SendGroupMsg(common.GenGroupPicMsg(req.GroupId, b))
				}
			}
			return nil
		}).
		//构建插件
		Build().
		//启动
		Start()
}

func getPic() ([]byte, error) {
	return pixiv.DownloadImage(randomAImage())
}
func randomAImage() string {
	mutex.Lock()
	defer mutex.Unlock()
	if len(allImages) == 0 {
		resetImages()
	}
	result := allImages[0]
	allImages = allImages[1:]
	return result
}

func resetImages() {
	mutex.Lock()
}
