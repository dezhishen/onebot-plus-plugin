package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dezhishen/onebot-plus-plugin/pkg/command"
)

func TestCommandHelp(t *testing.T) {
	var req BiliReq
	_, err := command.ParseWithDescription(".bili-live", &req, strings.Split(".bili-live -e h", " "), "修改监听的配置")
	println(fmt.Sprintf("%v", err))
}
