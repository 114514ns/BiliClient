package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/copier"
	"golang.org/x/net/html"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Video struct {
	Title       string
	Desc        string
	Author      string
	UID         int64
	Cover       string
	BV          string
	PublishAt   string
	AuthorFace  string
	Cid         int
	Duration    string
	Part        int
	ParentTitle string
	View        int
	Reply       int
	Coin        int
	Share       int
	Like        int
	Danmaku     int
	Favorite    int
}

func (client BiliClient) GetVideo(bv string) (result []Video) {
	res, _ := client.Resty.R().
		Get("https://api.bilibili.com/x/web-interface/view?bvid=" + bv)

	var resObj = VideoResponse{}
	json.Unmarshal(res.Body(), &resObj)
	fmt.Println(string(res.Body()))

	var array = []Video{}

	for i, item := range resObj.Data.Pages {
		var video = Video{}
		copier.Copy(&video, resObj.Data.Stat)
		video.Author = resObj.Data.Owner.Name
		video.ParentTitle = resObj.Data.Title
		video.BV = bv
		video.Desc = resObj.Data.Desc
		video.Title = item.Title
		video.Part = i + 1
		video.Cid = item.Cid
		video.Duration = FormatDuration(item.Duration)
		video.PublishAt = time.Unix(resObj.Data.PublishAt, 0).Format(time.DateTime)
		video.Cover = resObj.Data.Cover
		video.UID = resObj.Data.Owner.Mid
		video.AuthorFace = resObj.Data.Owner.Face
		array = append(array, video)
	}

	return array
}
func (client BiliClient) GetVideoByUser(mid int64, page int, byHot bool) (result []Video) {
	var order = "pubtime"
	if byHot {
		order = "click"
	}
	htmlRes, _ := client.Resty.R().Get("https://space.bilibili.com/2/upload/video")
	reader := bytes.NewReader(htmlRes.Body())
	root, _ := html.Parse(reader)
	find := goquery.NewDocumentFromNode(root).Find("script")
	access := ""
	dmImgStr := GenerateBase64RandomString(16, 64)
	dmImgStr = dmImgStr[0 : len(dmImgStr)-2]
	cover := GenerateBase64RandomString(32, 128)
	cover = cover[0 : len(cover)-2]
	inter := `{"ds":[],"wh":[0,0,0],"of":[0,0,0]}`
	find.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "access_id") {
			access, _ = url.PathUnescape(s.Text())
		}
	})
	var i interface{}
	json.Unmarshal([]byte(access), &i)
	m, _ := i.(map[string]interface{})
	access, _ = m["name"].(string)
	u := fmt.Sprintf("https://api.bilibili.com/x/space/wbi/arc/search?pn=42&ps=%d&mid=%d&order=%s&w_webid=%s&dm_img_list=[]&dm_img_str=%s&dm_cover_img_str=%s&dm_img_inter=%s", page, mid, order, access, dmImgStr, cover, inter)
	u1, _ := url.Parse(u)
	signed, _ := client.WBI.SignQuery(u1.Query(), time.Now())
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/space/wbi/arc/search?" + signed.Encode())
	var resObj = VideoResponse{}
	json.Unmarshal(res.Body(), &resObj)
	fmt.Println(string(res.Body()))
	return nil
}

func (client BiliClient) GetVideoStream(bv string, part int) []string {
	os.Mkdir("cache", 066)
	var videolink = "https://bilibili.com/video/" + bv + "?p=" + strconv.Itoa(part)
	vRes, _ := client.Resty.R().Get(videolink)
	htmlContent := vRes.Body()
	reader := bytes.NewReader(htmlContent)
	root, _ := html.Parse(reader)
	find := goquery.NewDocumentFromNode(root).Find("script")
	var arr = make([]string, 0)
	find.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "m4s") && strings.Contains(s.Text(), "backup_url") {
			var jsonStr = strings.Replace(s.Text(), "window.__playinfo__=", "", 1)
			var v = Dash{}
			json.Unmarshal([]byte(jsonStr), &v)
			arr = append(arr, v.Data.Dash0.Audio[0].Link)
			arr = append(arr, v.Data.Dash0.Video[0].Link)
		}
	})
	return arr
}
