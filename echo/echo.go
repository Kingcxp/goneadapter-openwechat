package echo

import (
	"github.com/gonebot-dev/gonebot/adapter"
	"github.com/gonebot-dev/gonebot/message"
	"github.com/gonebot-dev/gonebot/plugin"
	"github.com/gonebot-dev/gonebot/plugin/rule"
)

var Echo plugin.GonePlugin

func init() {
	Echo.Name = "Echo"
	Echo.Version = "v0.1.1"
	Echo.Description = "Reply the same message of what you have sent"

	Echo.Handlers = append(Echo.Handlers, plugin.GoneHandler{
		Rules: rule.NewRules(rule.ToMe()).And(rule.Command("echo")),
		Handler: func(a *adapter.Adapter, msg message.Message) bool {
			reply := message.NewReply(msg).Join(msg)
			a.SendMessage(reply)
			return true
		},
	})
}
