package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	url2 "net/url"
	"time"
)

var qqapi = "https://c.y.qq.com/soso/fcgi-bin/search_for_qq_cp?g_tk=5381&uin=0&format=json&inCharset=utf-8&outCharset=utf-8¬ice=0&platform=h5&needNewCode=1&zhidaqu=1&catZhida=1&t=0&flag=1&ie=utf-8&sem=1&aggr=0&perpage=20&n=20&p=1&remoteplace=txt.mqq.all&_=1520833663464"

// qqSearchList20 批量搜索qq音乐，列表 列出搜索结果 20首
func qqSearchList20(keyword string) (string, []MSG) {
	var lt []MSG
	var text string
	url := qqapi + "&w=" + url2.QueryEscape(keyword)
	client := http.Client{Timeout: time.Second * 2}
	header := make(http.Header)
	header.Set("Referer", "https://c.y.qq.com")
	header.Set("Host", "c.y.qq.com")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("search list err err:%v", err)
		return "", nil
	}
	req.Header = header
	res, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	if res != nil {
		//goland:noinspection GoDeferInLoop
		defer res.Body.Close()
	}
	r, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}
	result := gjson.ParseBytes(r)
	if result.Get("code").Int() == 0 {
		list := result.Get("data.song.list").Array()
		totalNum := result.Get("data.song.totalnum").Int()
		if totalNum > 0 {
			text = "搜索到以下歌曲：\n"
			for k, info := range list {
				lt = append(lt, MSG{
					"id":   info.Get("songid").String(),
					"sid":  info.Get("songmid").String(),
					"type": "qq",
				})
				text += fmt.Sprintf("No.%d 歌名：%s | 歌手:%s | 专辑：%s \n", k+1, info.Get("songname").String(), info.Get("singer").Array()[0].Get("name").String(), info.Get("albumname").String())
			}
			text += "请回复 选 1 这样的选+No编号进行选歌，1分钟内有效"
		}
	}
	return text, lt
}
