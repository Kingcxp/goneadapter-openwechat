package openwechat

import (
	"fmt"

	"github.com/eatmoreapple/openwechat"
	"github.com/gonebot-dev/gonebot/message"
)

type LocationType struct{}

func (loc LocationType) AdapterName() string {
	return OpenWechat.Name
}

func (loc LocationType) TypeName() string {
	return "location"
}

func (loc LocationType) ToRawText(msg message.MessageSegment) string {
	return "[OpenWechat:location]"
}

type RealtimeLocationStartType struct{}

func (rls RealtimeLocationStartType) AdapterName() string {
	return OpenWechat.Name
}

func (rls RealtimeLocationStartType) TypeName() string {
	return "realtime_location_start"
}

func (rls RealtimeLocationStartType) ToRawText(msg message.MessageSegment) string {
	return "[OpenWechat:realtime_location_start]"
}

type RealtimeLocationStopType struct{}

func (rls RealtimeLocationStopType) AdapterName() string {
	return OpenWechat.Name
}

func (rls RealtimeLocationStopType) TypeName() string {
	return "realtime_location_stop"
}

func (rls RealtimeLocationStopType) ToRawText(msg message.MessageSegment) string {
	return "[OpenWechat:realtime_location_stop]"
}

type FriendAddType struct {
	UserName string              `json:"UserName"`
	WechatID string              `json:"wechat_id"`
	Sex      string              `json:"sex"`
	Country  string              `json:"country"`
	Province string              `json:"province"`
	City     string              `json:"city"`
	Msg      *openwechat.Message `json:"msg"`
}

func (fa FriendAddType) AdapterName() string {
	return OpenWechat.Name
}

func (fa FriendAddType) TypeName() string {
	return "friend_add"
}

func (fa FriendAddType) ToRawText(msg message.MessageSegment) string {
	result := msg.Data.(FriendAddType)
	return fmt.Sprintf("[OpenWechat:friend_add,UserName=%s,sex=%s,country=%s,province=%s,city=%s]", result.UserName, result.Sex, result.Country, result.Province, result.City)
}

type CardType struct {
	UserName string `json:"UserName"`
	WechatID string `json:"wechat_id"`
	Sex      string `json:"sex"`
	Province string `json:"province"`
	City     string `json:"city"`
}

func (card CardType) AdapterName() string {
	return OpenWechat.Name
}

func (card CardType) TypeName() string {
	return "card"
}

func (card CardType) ToRawText(msg message.MessageSegment) string {
	result := msg.Data.(CardType)
	return fmt.Sprintf("[OpenWechat:card,UserName=%s,sex=%s,province=%s,city=%s]", result.UserName, result.Sex, result.Province, result.City)
}

type RecallType struct {
	Recaller   string `json:"recaller"`
	ReplaceMsg string `json:"replace_msg"`
}

func (recall RecallType) AdapterName() string {
	return OpenWechat.Name
}

func (recall RecallType) TypeName() string {
	return "recall"
}

func (recall RecallType) ToRawText(msg message.MessageSegment) string {
	result := msg.Data.(RecallType)
	return fmt.Sprintf("[OpenWechat:recall,recaller=%s,replace_msg=%s]", result.Recaller, result.ReplaceMsg)
}

type TransferType struct{}

func (transfer TransferType) AdapterName() string {
	return OpenWechat.Name
}

func (transfer TransferType) TypeName() string {
	return "transfer"
}

func (transfer TransferType) ToRawText(msg message.MessageSegment) string {
	return "[OpenWechat:transfer]"
}

type RedPacketType struct{}

func (redPacket RedPacketType) AdapterName() string {
	return OpenWechat.Name
}

func (redPacket RedPacketType) TypeName() string {
	return "red_packet"
}

func (redPacket RedPacketType) ToRawText(msg message.MessageSegment) string {
	return "[OpenWechat:red_packet]"
}

type TickleType struct {
	Msg string `json:"msg"`
}

func (tickle TickleType) AdapterName() string {
	return OpenWechat.Name
}

func (tickle TickleType) TypeName() string {
	return "tickle"
}

func (tickle TickleType) ToRawText(msg message.MessageSegment) string {
	return fmt.Sprintf("[OpenWechat:tickle,msg=%s]", tickle.Msg)
}

type JoinGroupType struct{}

func (joinGroup JoinGroupType) AdapterName() string {
	return OpenWechat.Name
}

func (joinGroup JoinGroupType) TypeName() string {
	return "join_group"
}

func (joinGroup JoinGroupType) ToRawText(msg message.MessageSegment) string {
	return fmt.Sprintf("[OpenWechat:join_group]")
}
