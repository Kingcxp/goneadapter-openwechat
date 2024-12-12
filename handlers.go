package openwechat

import (
	"encoding/base64"
	"io"
	"net/http"
	"strconv"

	"github.com/eatmoreapple/openwechat"
	"github.com/gonebot-dev/gonebot/logging"
	"github.com/gonebot-dev/gonebot/message"
	"github.com/rs/zerolog"
)

func receiveHandler(msg *openwechat.Message) {
	formatMsg := message.NewMessage()
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
