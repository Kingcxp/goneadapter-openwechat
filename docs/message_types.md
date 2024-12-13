# 消息类型

[消息段类型](#locationtype)
- [位置信息](#locationtype)
- [位置共享开启信息](#realtimelocationstarttype)
- [位置共享结束信息](#realtimelocationstoptype)
- [好友添加信息](#friendaddtype)
- [名片信息](#cardtype)
- [撤回消息](#recalltype)
- [转账消息](#transfertype)
- [红包消息](#redpackettype)
- [戳一戳消息](#tickletype)
- [入群消息](#joingrouptype)


**注意：OpenWechat 在收到消息时涉及了这些消息段类型，但你不应当在回复消息时使用它们，OpenWechat 并不支持这些消息类型的发送**

**这些消息段会以对应的 `TypeName()` 中指定的类型出现在消息段中，你可以通过 `msg.GetSegments()[0].Type` 来判断该消息是否为消息段消息**


### LocationType
位置信息，OpenWechat 无法获取具体内容
```go
type LocationType struct {}
```

### RealtimeLocationStartType
位置共享开启信息，OpenWechat 无法获取具体内容
```go
type RealtimeLocationStartType struct {}
```

### RealtimeLocationStopType
位置共享结束信息，OpenWechat 无法获取具体内容
```go
type RealtimeLocationStopType struct {}
```

### FriendAddType
好友添加信息，提供如下字段：
```go
type FriendAddType struct {
	UserName string              `json:"UserName"`
	WechatID string              `json:"wechat_id"`
	Sex      string              `json:"sex"`
	Country  string              `json:"country"`
	Province string              `json:"province"`
	City     string              `json:"city"`
	Msg      *openwechat.Message `json:"msg"`
}
```

### CardType
名片信息，提供如下字段：
```go
type CardType struct {
	UserName string `json:"UserName"`
	WechatID string `json:"wechat_id"`
	Sex      string `json:"sex"`
	Province string `json:"province"`
	City     string `json:"city"`
}
```

### RecallType
撤回消息，提供消息撤回者名称和用来替换撤回消息的内容
```go
type RecallType struct {
	Recaller   string `json:"recaller"`
	ReplaceMsg string `json:"replace_msg"`
}
```

### TransferType
转账消息，无法获知具体内容
```go
type TransferType struct{}
```

### RedPacketType
红包消息，无法获知具体内容
```go
type RedPacketType struct{}
```

### TickleType
戳一戳消息，能知道戳一戳消息的具体文本内容
```go
type TickleType struct {
	Msg string `json:"msg"`
}
```

### JoinGroupType
入群消息，无法获知具体内容
```go
type JoinGroupType struct {}
```
