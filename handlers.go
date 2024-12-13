package openwechat

import (
	"bytes"
	"encoding/base64"
	"image"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/eatmoreapple/openwechat"
	"github.com/gonebot-dev/gonebot/logging"
	"github.com/gonebot-dev/gonebot/message"
	"github.com/rs/zerolog"
)

type EmptyActionResult struct{}

func actionHandler() {
	for {
		msg := OpenWechat.ActionChannel.Pull()
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "Ignored action call.")
		*(msg.ResultChannel) <- EmptyActionResult{}
	}
}

func receiveHandler(msg *openwechat.Message) {
	formatMsg := message.NewMessage()
	formatMsg.Group = ""
	formatMsg.Self = Self.UserName
	from := "friend"
	if msg.IsSendByFriend() {
		formatMsg.IsToMe = true
		sender, _ := msg.Sender()
		receiver, _ := msg.Receiver()
		formatMsg.Sender = sender.UserName
		formatMsg.Receiver = receiver.UserName
	} else if msg.IsSendByGroup() {
		from = "group"
		groupUser, _ := msg.Sender()
		sender, _ := msg.SenderInGroup()
		group, _ := groupUser.AsGroup()
		formatMsg.Group = group.UserName
		formatMsg.Sender = sender.UserName
	}
	if msg.IsText() {
		msg.AsRead()
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s text message: %s", from, msg.Content)
		if msg.IsAt() && msg.ToUserName == Self.UserName {
			formatMsg.IsToMe = true
		}
		formatMsg.Text(msg.Content)
	} else if msg.IsPicture() || msg.IsEmoticon() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s picture message.", from)
		msg.AsRead()
		response, err := msg.GetPicture()
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "receiveHandler: Get picture error: %v", err)
			return
		}
		img, err := io.ReadAll(response.Body)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "receiveHandler: Read picture error: %v", err)
			return
		}
		base64Encoding := "base64://" + base64.StdEncoding.EncodeToString(img)
		formatMsg.Image(base64Encoding)
	} else if msg.IsLocation() {
		// Cannot get location info from location message
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s location message.", from)
		msg.AsRead()
		formatMsg.Any(LocationType{})
	} else if msg.IsRealtimeLocationStart() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s realtime location start message.", from)
		msg.AsRead()
		formatMsg.Any(RealtimeLocationStartType{})
	} else if msg.IsRealtimeLocationStop() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s realtime location stop message.", from)
		msg.AsRead()
		formatMsg.Any(RealtimeLocationStopType{})
	} else if msg.IsVoice() {
		// Just to base64
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s voice message.", from)
		msg.AsRead()
		response, _ := msg.GetVoice()
		voi, _ := io.ReadAll(response.Body)
		formatMsg.Voice(base64.StdEncoding.EncodeToString(voi))
	} else if msg.IsFriendAdd() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s friend add message.", from)
		addmsg, err := msg.FriendAddMessageContent()
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "receiveHandler: Read friend add message error: %s", err.Error())
			return
		}
		var sex string
		if addmsg.Sex == 1 {
			sex = "男"
		} else if addmsg.Sex == 0 {
			sex = "女"
		} else {
			sex = "未知"
		}
		formatMsg.Any(FriendAddType{
			NickName: addmsg.FromNickName,
			UserName: addmsg.FromUserName,
			WechatID: addmsg.Alias,
			Sex:      sex,
			Country:  addmsg.Country,
			Province: addmsg.Province,
			City:     addmsg.City,
			Msg:      msg,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s friend add message: %s", from, FriendAddType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsCard() {
		msg.AsRead()
		cardmsg, err := msg.Card()
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "receiveHandler: Read %s card message error: %s", from, err.Error())
			return
		}
		var sex string
		if cardmsg.Sex == 1 {
			sex = "男"
		} else if cardmsg.Sex == 0 {
			sex = "女"
		} else {
			sex = "未知"
		}
		formatMsg.Any(CardType{
			NickName: cardmsg.NickName,
			UserName: cardmsg.UserName,
			Sex:      sex,
			WechatID: cardmsg.Alias,
			Province: cardmsg.Province,
			City:     cardmsg.City,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s card message: %s", from, CardType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsVideo() {
		// ? To fetch the video here is really a bad idea.
		// * So we just put the msg string here.
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s video message.", from)
		msg.AsRead()
		formatMsg.Video(msg.String())
	} else if msg.IsRecalled() {
		msg.AsRead()
		revokemsg, err := msg.RevokeMsg()
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "receiveHandler: Read recalled message error: %s", err.Error())
			return
		}
		formatMsg.Any(RecallType{
			Recaller:   formatMsg.Sender,
			ReplaceMsg: revokemsg.RevokeMsg.ReplaceMsg,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s recall message: %s", from, RecallType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsSystem() {
		logging.Log(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Ignored system message.")
		return
	} else if msg.IsTransferAccounts() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s transfer accounts message.", from)
		msg.AsRead()
		formatMsg.Any(TransferType{})
	} else if msg.IsReceiveRedPacket() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s receive red packet message.", from)
		msg.AsRead()
		formatMsg.Any(RedPacketType{})
	} else if msg.IsTickled() {
		if msg.IsTickledMe() {
			formatMsg.IsToMe = true
		}
		formatMsg.Any(TickleType{
			Msg: msg.Content,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received %s tickle message: %s", from, TickleType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsJoinGroup() {
		formatMsg.Any(JoinGroupType{})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received join group message: %s", TickleType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else {
		logging.Log(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Ignored unknown message.")
		return
	}
	OpenWechat.ReceiveChannel.Push(*formatMsg, true)
}

func isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

func isBase64Img(str string) bool {
	return strings.HasPrefix(str, "base64://")
}

func sendImageToFriend(friendUserName string, img message.ImageType) {
	friends, err := Self.Friends()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToFriend: Unable to get friends: %s", err.Error())
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "sendImageToFriend: Image to friend %s.", friendUserName)
	friend := friends.SearchByUserName(1, friendUserName).First()
	if friend == nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToFriend: Friend %s not found.", friendUserName)
		return
	}
	_, err = os.Stat(img.File)
	if isURL(img.File) {
		resp, err := http.Get(img.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToFriend: Unable to get image from url: %s, Error: %s", img.File, err.Error())
			return
		}
		friend.SendImage(resp.Body)
		resp.Body.Close()
	} else if isBase64Img(img.File) {
		img.File = strings.ReplaceAll(img.File, "base64://", "")
		imgData, err := base64.StdEncoding.DecodeString(img.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToFriend: Unable to decode base64 image: %s", err.Error())
			return
		}
		image.Decode(bytes.NewReader(imgData))
		friend.SendImage(bytes.NewReader(imgData))
	} else if err == nil {
		imgData, _ := os.Open(img.File)
		friend.SendImage(imgData)
		imgData.Close()
	} else {
		logging.Log(zerolog.WarnLevel, "OpenWechat", "sendImageToFriend: Unknown image type.")
	}
}

func sendImageToGroup(groupUserName string, img message.ImageType) {
	groups, err := Self.Groups()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToGroup: Unable to get groups: %s", err.Error())
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "sendImageToFriend: Image to friend %s: %s", groupUserName)
	group := groups.SearchByUserName(1, groupUserName).First()
	if group == nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToGroup: Group %s not found.", groupUserName)
		return
	}
	_, err = os.Stat(img.File)
	if isURL(img.File) {
		resp, err := http.Get(img.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToGroup: Unable to get image from url: %s, Error: %s", img.File, err.Error())
			return
		}
		group.SendImage(resp.Body)
		resp.Body.Close()
	} else if isBase64Img(img.File) {
		img.File = strings.ReplaceAll(img.File, "base64://", "")
		imgData, err := base64.StdEncoding.DecodeString(img.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToGroup: Unable to decode base64 image: %s", err.Error())
			return
		}
		group.SendImage(bytes.NewReader(imgData))
	} else if err == nil {
		imgData, _ := os.Open(img.File)
		group.SendImage(imgData)
		imgData.Close()
	} else {
		logging.Log(zerolog.WarnLevel, "OpenWechat", "sendImageToGroup: Unknown image type.")
	}
}

func sendFileToFriend(friendUserName string, f message.FileType) {
	friends, err := Self.Friends()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToFriend: Unable to get friends: %s", err.Error())
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "sendFileToFriend: File to friend %s.", friendUserName)
	friend := friends.SearchByUserName(1, friendUserName).First()
	if friend == nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToFriend: Friend %s not found.", friendUserName)
		return
	}
	_, err = os.Stat(f.File)
	if isURL(f.File) {
		resp, err := http.Get(f.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToFriend: Unable to get file from url: %s, Error: %s", f.File, err.Error())
			return
		}
		friend.SendFile(resp.Body)
		resp.Body.Close()
	} else if err == nil {
		fileData, _ := os.Open(f.File)
		friend.SendFile(fileData)
		fileData.Close()
	} else {
		logging.Logf(zerolog.WarnLevel, "OpenWechat", "sendFileToFriend: Unknown file type: %s", f.File)
	}
}

func sendFileToGroup(groupUserName string, f message.FileType) {
	groups, err := Self.Groups()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToGroup: Unable to get groups: %s", err.Error())
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "sendFileToGroup: File to friend %s.", groupUserName)
	group := groups.SearchByUserName(1, groupUserName).First()
	if group == nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToGroup: Group %s not found.", groupUserName)
		return
	}
	_, err = os.Stat(f.File)
	if isURL(f.File) {
		resp, err := http.Get(f.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToGroup: Unable to get file from url: %s, Error: %s", f.File, err.Error())
			return
		}
		group.SendFile(resp.Body)
		resp.Body.Close()
	} else if err == nil {
		fileData, _ := os.Open(f.File)
		group.SendFile(fileData)
		fileData.Close()
	} else {
		logging.Logf(zerolog.WarnLevel, "OpenWechat", "sendFileToGroup: Unknown file type: %s", f.File)
	}
}

func sendTextToFriend(friendUserName string, text string) {
	friends, err := Self.Friends()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendTextToFriend: Unable to get friends: %s", err.Error())
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "sendTextToFriend: Text to friend %s: %s", friendUserName, text)
	friend := friends.SearchByUserName(1, friendUserName).First()
	if friend == nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendTextToFriend: Friend %s not found.", friendUserName)
		return
	}
	friend.SendText(text)
}

func sendTextToGroup(groupUserName string, text string) {
	groups, err := Self.Groups()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendTextToGroup: Unable to get groups: %s", err.Error())
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "sendTextToGroup: Text to group %s: %s", groupUserName, text)
	group := groups.SearchByUserName(1, groupUserName).First()
	if group == nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendTextToGroup: Group %s not found.", groupUserName)
		return
	}
	group.SendText(text)
}

func sendImage(receiver, group string, img message.ImageType) {
	if group == "" {
		sendImageToFriend(receiver, img)
	} else {
		sendImageToGroup(group, img)
	}
}

func sendFile(receiver, group string, f message.FileType) {
	if group == "" {
		sendFileToFriend(receiver, f)
	} else {
		sendFileToGroup(group, f)
	}
}

func sendText(receiver, group string, text string) {
	if group == "" {
		sendTextToFriend(receiver, text)
	} else {
		sendTextToGroup(group, text)
	}
}

func sendHandler() {
	for {
		msg := OpenWechat.SendChannel.Pull()
		text := ""
		hasText := false
		for _, segment := range msg.GetSegments() {
			if segment.Type == "image" {
				if hasText {
					sendText(msg.Receiver, msg.Group, text)
					text = ""
					hasText = false
				}
				sendImage(msg.Receiver, msg.Group, segment.Data.(message.ImageType))
			} else if segment.Type == "file" {
				if hasText {
					sendText(msg.Receiver, msg.Group, text)
					text = ""
					hasText = false
				}
				sendFile(msg.Receiver, msg.Group, segment.Data.(message.FileType))
			} else if segment.Type == "text" {
				hasText = true
				text += segment.Data.(message.TextType).Text
			}
		}
		if hasText {
			sendText(msg.Receiver, msg.Group, text)
		}
	}
}
