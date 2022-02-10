package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/scjtqs2/bot_adapter/coolq"
	"github.com/scjtqs2/bot_adapter/event"
	"github.com/scjtqs2/bot_adapter/pb/entity"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/scjtqs2/bot_music/music"
)

// QqMusicStatus 缓存
const QqMusicStatus = "QQ_MUSIC_STATUS_"

func parseMsg(data string) {
	msg := gjson.Parse(data)
	switch msg.Get("post_type").String() {
	case "message": // 消息事件
		switch msg.Get("message_type").String() {
		case event.MESSAGE_TYPE_PRIVATE:
			var req event.MessagePrivate
			_ = json.Unmarshal([]byte(msg.Raw), &req)
			parsePrivateMsg(req)
		case event.MESSAGE_TYPE_GROUP:
			var req event.MessageGroup
			_ = json.Unmarshal([]byte(msg.Raw), &req)
			parseGroup(req)
		}
	case "notice": // 通知事件
		switch msg.Get("notice_type").String() {
		case event.NOTICE_TYPE_FRIEND_ADD:
			var req event.NoticeFriendAdd
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_FRIEND_RECALL:
			var req event.NoticeFriendRecall
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_GROUP_BAN:
			var req event.NoticeGroupBan
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_GROUP_DECREASE:
			var req event.NoticeGroupDecrease
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_GROUP_INCREASE:
			var req event.NoticeGroupIncrease
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_GROUP_ADMIN:
			var req event.NoticeGroupAdmin
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_GROUP_RECALL:
			var req event.NoticeGroupRecall
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_GROUP_UPLOAD:
			var req event.NoticeGroupUpload
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_POKE:
			var req event.NoticePoke
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_HONOR:
			var req event.NoticeHonor
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.NOTICE_TYPE_LUCKY_KING:
			var req event.NoticeLuckyKing
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.CUSTOM_NOTICE_TYPE_GROUP_CARD:
		case event.CUSTOM_NOTICE_TYPE_OFFLINE_FILE:
		}
	case "request": // 请求事件
		switch msg.Get("request_type").String() {
		case event.REQUEST_TYPE_FRIEND:
			var req event.RequestFriend
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.REQUEST_TYPE_GROUP:
			var req event.RequestGroup
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		}
	case "meta_event": // 元事件
		switch msg.Get("meta_event_type").String() {
		case event.META_EVENT_LIFECYCLE:
			var req event.MetaEventLifecycle
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		case event.META_EVENT_HEARTBEAT:
			var req event.MetaEventHeartbeat
			_ = json.Unmarshal([]byte(msg.Raw), &req)
		}
	}
}

func checkKeywords(message interface{}, atqq int64, fromqq int64, cachekey string, isGroup bool) {
	var (
		str string
		ok  bool
	)
	// 判断是 string还是 array
	if str, ok = message.(string); !ok {
		b, _ := json.Marshal(message)
		str = gjson.ParseBytes(b).Array()[0].Get("data.text").String()
	}
	c := cache.Get(cachekey)
	if c != nil && !c.Expired() && c.Value() != nil {
		// 缓存有效 . 处理 搜索结果选择
		if !strings.HasPrefix(str, "选") {
			if isGroup {
				_, _ = botAdapterClient.SendGroupMsg(context.TODO(), &entity.SendGroupMsgReq{
					GroupId: fromqq,
					Message: []byte("输入的选择有误，请重新选择编号"),
				})
				return
			}
			_, _ = botAdapterClient.SendPrivateMsg(context.TODO(), &entity.SendPrivateMsgReq{
				UserId:  fromqq,
				Message: []byte("输入的选择有误，请重新选择编号"),
			})
			return
		}
		num, err := strconv.ParseInt(strings.TrimSpace(strings.ReplaceAll(str, "选", "")), 10, 10)
		if err != nil {
			log.Errorf("回答选择错误，err:=%v", err)
			if isGroup {
				_, _ = botAdapterClient.SendGroupMsg(context.TODO(), &entity.SendGroupMsgReq{
					GroupId: fromqq,
					Message: []byte("输入的选择有误，请重新选择编号"),
				})
				return
			}
			_, _ = botAdapterClient.SendPrivateMsg(context.TODO(), &entity.SendPrivateMsgReq{
				UserId:  fromqq,
				Message: []byte("输入的选择有误，请重新选择编号"),
			})
			return
		}
		cache.Delete(cachekey)
		res := gjson.ParseBytes(c.Value().([]byte)).Array()[num-1]
		text := coolq.EnMusicCode(res.Get("type").String(), res.Get("id").String())
		if isGroup {
			_, _ = botAdapterClient.SendGroupMsg(context.TODO(), &entity.SendGroupMsgReq{
				GroupId: fromqq,
				Message: []byte(text),
			})
			return
		}
		_, _ = botAdapterClient.SendPrivateMsg(context.TODO(), &entity.SendPrivateMsgReq{
			UserId:  fromqq,
			Message: []byte(text),
		})
		return
	}
	var retstr string
	var list []music.MSG
	// 处理 点歌步奏，罗列搜索结果
	if strings.HasPrefix(str, "点歌") {
		keywords := strings.Split(str, "点歌")
		if len(keywords) < 2 {
			return
		}
		retstr, list = music.QQSearchList20(strings.TrimSpace(keywords[1]))
	}
	if strings.HasPrefix(str, "qq点歌") {
		keywords := strings.Split(str, "qq点歌")
		if len(keywords) < 2 {
			return
		}
		retstr, list = music.QQSearchList20(strings.TrimSpace(keywords[1]))
	}
	if strings.HasPrefix(str, "网易点歌") {
		keywords := strings.Split(str, "网易点歌")
		if len(keywords) < 2 {
			return
		}
		retstr, list = music.WangyisearchList20(strings.TrimSpace(keywords[1]))
	}
	if retstr == "" {
		return
	}
	b, _ := json.Marshal(list)
	cache.Set(cachekey, b, time.Minute)
	if isGroup {
		text := fmt.Sprintf("%s %s", coolq.EnAtCode(fmt.Sprintf("%d", atqq)), retstr)
		_, _ = botAdapterClient.SendGroupMsg(context.TODO(), &entity.SendGroupMsgReq{
			GroupId: fromqq,
			Message: []byte(text),
		})
		return
	}
	_, _ = botAdapterClient.SendPrivateMsg(context.TODO(), &entity.SendPrivateMsgReq{
		UserId:  fromqq,
		Message: []byte(retstr),
	})
}

func parsePrivateMsg(req event.MessagePrivate) {
	key := fmt.Sprintf("%s%d-%d-%d", QqMusicStatus, 0, req.UserID, req.SelfID)
	checkKeywords(req.Message, 0, req.Sender.UserID, key, false)
}

func parseGroup(req event.MessageGroup) {
	key := fmt.Sprintf("%s%d-%d-%d", QqMusicStatus, req.GroupID, req.UserID, req.SelfID)
	checkKeywords(req.Message, req.Sender.UserID, req.GroupID, key, true)
}
