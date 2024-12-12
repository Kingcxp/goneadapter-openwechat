package openwechat

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strconv"
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
	formatMsg.Self = strconv.FormatInt(Self.ID(), 10)
	if msg.IsSendByFriend() {
		formatMsg.IsToMe = true
		sender, _ := msg.Sender()
		receiver, _ := msg.Receiver()
		formatMsg.Sender = sender.NickName
		formatMsg.Receiver = receiver.NickName
		formatMsg.Group = ""
	} else if msg.IsSendByGroup() {
		group, _ := msg.Sender()
		sender, _ := msg.SenderInGroup()
		formatMsg.Group = openwechat.Group{User: group}.NickName
		formatMsg.Sender = sender.NickName
	}
	if msg.IsText() {
		msg.AsRead()
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received text message: %s", msg.Content)
		if msg.IsAt() && msg.ToUserName == Self.UserName {
			formatMsg.IsToMe = true
		}
		formatMsg.Text(msg.Content)
	} else if msg.IsPicture() || msg.IsEmoticon() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received picture message.")
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
		mimeType := http.DetectContentType(img)
		base64Encoding := ""
		switch mimeType {
		case "image/jpeg":
			base64Encoding += "image/jpeg;base64,"
			return
		case "image/png":
			base64Encoding += "image/png;base64,"
			return
		}
		base64Encoding += base64.StdEncoding.EncodeToString(img)
		formatMsg.Image(base64Encoding)
	} else if msg.IsLocation() {
		// Cannot get location info from location message
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received location message.")
		msg.AsRead()
		formatMsg.Any(LocationType{})
	} else if msg.IsRealtimeLocationStart() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received realtime location start message.")
		msg.AsRead()
		formatMsg.Any(RealtimeLocationStartType{})
	} else if msg.IsRealtimeLocationStop() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received realtime location stop message.")
		msg.AsRead()
		formatMsg.Any(RealtimeLocationStopType{})
	} else if msg.IsVoice() {
		// Just to base64
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received voice message.")
		msg.AsRead()
		response, _ := msg.GetVoice()
		voi, _ := io.ReadAll(response.Body)
		formatMsg.Voice(base64.StdEncoding.EncodeToString(voi))
	} else if msg.IsFriendAdd() {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received friend add message.")
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
			WechatID: addmsg.Alias,
			Sex:      sex,
			Country:  addmsg.Country,
			Province: addmsg.Province,
			City:     addmsg.City,
			Msg:      msg,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received friend add message: %s", FriendAddType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsCard() {
		msg.AsRead()
		cardmsg, err := msg.Card()
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "receiveHandler: Read card message error: %s", err.Error())
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
			Sex:      sex,
			WechatID: cardmsg.Alias,
			Province: cardmsg.Province,
			City:     cardmsg.City,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received card message: %s", CardType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsVideo() {
		// ? To fetch the video here is really a bad idea.
		// * So we just put the msg string here.
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received video message.")
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
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received recall message: %s", RecallType{}.ToRawText(formatMsg.GetSegments()[0]))
	} else if msg.IsSystem() {
		logging.Log(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Ignored system message.")
		return
	} else if msg.IsTransferAccounts() {
		logging.Log(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received transfer accounts message.")
		msg.AsRead()
		formatMsg.Any(TransferType{})
	} else if msg.IsReceiveRedPacket() {
		logging.Log(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received receive red packet message.")
		msg.AsRead()
		formatMsg.Any(RedPacketType{})
	} else if msg.IsTickled() {
		if msg.IsTickledMe() {
			formatMsg.IsToMe = true
		}
		formatMsg.Any(TickleType{
			Msg: msg.Content,
		})
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "receiveHandler: Received tickle message: %s", TickleType{}.ToRawText(formatMsg.GetSegments()[0]))
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
	return strings.Contains(str, "image/jpeg;base64,") || strings.Contains(str, "image/png;base64,")
}

func sendImageToFriend(friendNickName string, img message.ImageType) {
	friends, err := Self.Friends()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToFriend: Unable to get friends: %s", err.Error())
		return
	}
	friend := friends.SearchByNickName(1, friendNickName).First()
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
		imgData, err := base64.StdEncoding.DecodeString(img.File)
		if err != nil {
			logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToFriend: Unable to decode base64 image: %s", err.Error())
			return
		}
		friend.SendImage(bytes.NewReader(imgData))
	} else if err == nil {
		imgData, _ := os.Open(img.File)
		friend.SendImage(imgData)
		imgData.Close()
	} else {
		logging.Logf(zerolog.WarnLevel, "OpenWechat", "sendImageToFriend: Unknown image type: %s", img.File)
	}
}

func sendImageToGroup(groupNickName string, img message.ImageType) {
	groups, err := Self.Groups()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendImageToGroup: Unable to get groups: %s", err.Error())
		return
	}
	group := groups.SearchByNickName(1, groupNickName).First()
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
		logging.Logf(zerolog.WarnLevel, "OpenWechat", "sendImageToGroup: Unknown image type: %s", img.File)
	}
}

func sendFileToFriend(friendNickName string, f message.FileType) {
	friends, err := Self.Friends()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToFriend: Unable to get friends: %s", err.Error())
		return
	}
	friend := friends.SearchByNickName(1, friendNickName).First()
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

func sendFileToGroup(groupNickName string, f message.FileType) {
	groups, err := Self.Groups()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendFileToGroup: Unable to get groups: %s", err.Error())
		return
	}
	group := groups.SearchByNickName(1, groupNickName).First()
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

func sendTextToFriend(friendNickName string, text string) {
	friends, err := Self.Friends()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendTextToFriend: Unable to get friends: %s", err.Error())
		return
	}
	friend := friends.SearchByNickName(1, friendNickName).First()
	friend.SendText(text)
}

func sendTextToGroup(groupNickName string, text string) {
	groups, err := Self.Groups()
	if err != nil {
		logging.Logf(zerolog.ErrorLevel, "OpenWechat", "sendTextToGroup: Unable to get groups: %s", err.Error())
		return
	}
	group := groups.SearchByNickName(1, groupNickName).First()
	group.SendText(text)
}

func sendHandler() {
	for {
		msg := OpenWechat.SendChannel.Pull()
		text := ""
		hasText := false
		for _, segment := range msg.GetSegments() {
			if segment.Type == "image" {
				if msg.Group == "" {
					sendImageToFriend(msg.Sender, segment.Data.(message.ImageType))
				} else {
					sendImageToGroup(msg.Group, segment.Data.(message.ImageType))
				}
			} else if segment.Type == "file" {
				if msg.Group == "" {
					sendFileToFriend(msg.Sender, segment.Data.(message.FileType))
				} else {
					sendFileToGroup(msg.Group, segment.Data.(message.FileType))
				}
			} else if segment.Type == "text" {
				hasText = true
				text += segment.Data.(message.TextType).Text
			}
		}
		if hasText {
			if msg.Group == "" {
				sendTextToFriend(msg.Sender, text)
			} else {
				sendTextToGroup(msg.Group, text)
			}
		}
	}
}
