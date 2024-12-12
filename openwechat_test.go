package openwechat_test

import (
	"testing"

	openwechat "github.com/gonebot-dev/goneadapter-openwechat"
	"github.com/gonebot-dev/gonebot"
	echo "github.com/gonebot-dev/goneplugin-echo"
	status "github.com/gonebot-dev/goneplugin-status"
)

func TestMain(m *testing.M) {
	gonebot.LoadPlugin(&echo.Echo)
	gonebot.LoadPlugin(&status.Status)

	gonebot.LoadAdapter(&openwechat.OpenWechat)

	gonebot.Run()
}
