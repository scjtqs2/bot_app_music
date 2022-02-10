package music

import (
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var wangyiapi = "http://s.music.163.com/search/get/?src=lofter&type=1&limit=20&offset=0&callback"

// WangyisearchList20 批量查询 网易云 列表
func WangyisearchList20(keyword string) (string, []MSG) {
	var lt []MSG
	var text string
	url := wangyiapi + "&s=" + url2.QueryEscape(keyword)
	client := http.Client{Timeout: time.Second * 2}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("search list err err:%v", err)
		return "", nil
	}
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("WangyisearchList20 list http err:%v", err)
		return "", nil
	}
	if res != nil {
		//goland:noinspection GoDeferInLoop
		defer res.Body.Close()
	}
	r, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("WangyisearchList20 list read resp err:%v", err)
		return "", nil
	}
	result := gjson.ParseBytes(r)
	if result.Get("code").Int() == 200 {
		list := result.Get("result.songs").Array()
		totalNum := result.Get("result.songCount").Int()
		if totalNum > 0 {
			text = "搜索到以下歌曲 \n"
			for k, info := range list {
				name := strings.TrimSpace(info.Get("name").String())
				id := info.Get("id").String()
				picurl := info.Get("album.picUrl").String() + "?param=90y90"
				pageurl := info.Get("page").String()
				pre := ""
				author := ""
				for _, v := range info.Get("artists").Array() {
					author += pre + v.Get("name").String()
					pre = "、"
				}
				lt = append(lt, MSG{
					"id":      id,
					"pageurl": pageurl,
					"picurl":  picurl,
					"name":    name,
					"type":    "163",
				})
				text += fmt.Sprintf("No.%d 歌名：%s | 歌手:%s | 专辑：%s \n", k+1, name, author, info.Get("album.name").String())
			}
			text += "请回复 选 1 这样的选+No编号进行选歌，1分钟内有效"
		}
	}
	return text, lt
}
