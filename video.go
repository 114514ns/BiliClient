package bili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"os"
	"strconv"
	"strings"
	"time"
)

type Video struct {
	Aid         int64
	Title       string
	Desc        string
	UName       string
	UID         int64
	Cover       string
	BV          string
	CreateAt    time.Time
	Face        string
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
	Tags        []string
}

func (v *Video) GetStream(client *BiliClient) []string {
	return client.GetVideoStream(v.BV, 1)
}

func parseVideo() {

}

func (client BiliClient) GetVideo(bv string) (result []Video) {
	res, _ := client.Resty.R().
		Get("https://api.bilibili.com/x/web-interface/view/detail?bvid=" + bv)

	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	var array = []Video{}
	for _, i := range getArray(obj, "data.View.pages") {
		var video = Video{}
		video.Desc = getString(obj, "data.View.desc")
		video.Title = getString(i, "part")
		video.BV = bv
		video.Duration = FormatDuration(getInt(i, "duration"))
		video.Cover = getString(obj, "data.View.pic")
		video.UID = getInt64(obj, "data.View.owner.mid")
		video.Reply = getInt(obj, "data.View.stat.reply")
		video.View = getInt(obj, "data.View.stat.view")
		video.Danmaku = getInt(obj, "data.View.stat.danmaku")
		video.Favorite = getInt(obj, "data.View.stat.favorite")
		video.Coin = getInt(obj, "data.View.stat.coin")
		video.Like = getInt(obj, "data.View.stat.like")
		video.Share = getInt(obj, "data.View.stat.share")
		video.CreateAt = time.Unix(int64(getInt(obj, "data.View.ctime")), 0)
		video.Reply = getInt(obj, "data.View.stat.reply")
		video.UName = getString(obj, "data.View.owner.name")
		video.Face = getString(obj, "data.View.owner.face")
		video.Aid = getInt64(obj, "data.View.stat.aid")
		video.Cid = getInt(i, "cid")
		video.Tags = []string{}
		for _, tag := range getArray(obj, "data.Tags") {
			video.Tags = append(video.Tags, getString(tag, "tag_name"))
		}
		array = append(array, video)
	}

	return array
}
func (client *BiliClient) GetVideoByUser(mid int64, page int, byHot bool) (result []Video) {
	var order = "pubtime"
	if byHot {
		order = "click"
	}
	u := fmt.Sprintf("https://api.bilibili.com/x/space/arc/search?pn=%d&ps=42&mid=%d&order=%s&web_location=bilibili-electron", page, mid, order)

	res, _ := client.Resty.R().Get(u)
	var resObj = UserVideoListResponse{}
	json.Unmarshal(res.Body(), &resObj)
	list := make([]Video, 0)

	for _, s := range resObj.Data.List.Vlist {
		var video = Video{}
		video.BV = s.Bvid
		video.UName = s.Author
		video.UID = s.Mid
		video.Desc = s.Description
		video.Title = s.Title
		video.View = s.Play
		video.Cover = s.Pic
		video.Reply = s.Comment
		video.Danmaku = s.VideoReview
		video.CreateAt = time.Unix(int64(s.Created), 0)
		video.Duration = s.Length
		video.Aid = s.Aid
		list = append(list, video)
	}
	return list
}
func (video *Video) getComments(cursor string, client *BiliClient, sort ...ReplySort) ([]Comment, string) {
	if client.UID == 0 {
		return client.GetCommentRPC(video.Aid, cursor, CommentType.Video, sort...)
	}
	return client.GetComment(video.Aid, cursor, CommentType.Video)
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
