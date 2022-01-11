package common

import (
	"encoding/base64"

	"github.com/dezhishen/onebot-sdk/pkg/model"
)

func GenGroupPicMsg(groupId int64, buf []byte) *model.GroupMsg {
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

func GenGroupTextMsg(groupId int64, text string) *model.GroupMsg {
	return &model.GroupMsg{
		GroupId: groupId,
		Message: []*model.MessageSegment{
			{Type: "text", Data: &model.MessageElementText{
				Text: text,
			}},
		},
	}
}
