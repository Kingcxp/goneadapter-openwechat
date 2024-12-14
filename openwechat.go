package openwechat

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/gonebot-dev/gonebot/adapter"
)

var OpenWechat adapter.Adapter
var Self *openwechat.Self
var Emoji = openwechat.Emoji

func init() {
	OpenWechat.Name = "OpenWechat"
	OpenWechat.Version = "v0.1.2"
	OpenWechat.Description = "The openwechat adapter for gonebot."
	OpenWechat.Start = start
	OpenWechat.Finalize = finalize
}
