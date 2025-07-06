package bili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"net/url"
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
	Tags        string
	RawResponse string
}

func (v *Video) GetStream(client *BiliClient) []string {
	return client.GetVideoStream(v.BV, 1)
}

func parseVideo() {

}

func (client *BiliClient) GetVideo(bv string) (result []Video) {
	res, _ := client.Resty.R().
		Get("https://api.bilibili.com/x/web-interface/view/detail?bvid=" + bv)

	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	var array = []Video{}
	if getArray(obj, "data.View.pages") == nil {
		time.Now()
	}
	for _, i := range getArray(obj, "data.View.pages") {
		var video = Video{}
		video.Desc = getString(obj, "data.View.desc")

		video.Title = getString(i, "part")
		if video.Title == "" {
			video.Title = getString(obj, "data.View.title")
		} else {
			video.ParentTitle = getString(obj, "data.View.title")
		}

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
		var tags = ""
		video.RawResponse = res.String()
		for _, tag := range getArray(obj, "data.Tags") {
			tags = tags + getString(tag, "tag_name") + ","
		}
		video.Tags = tags
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
func (client *BiliClient) GetVideoStream(bv string, part int) []string {
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

type SearchOption struct {
	BeginTime time.Time
	EndTime   time.Time
	Keyword   string
	Page      int
}

func (client *BiliClient) SearchVideo(opinion SearchOption) []Video {
	if opinion.Page == 0 {
		opinion.Page = 1
	}
	if opinion.BeginTime.IsZero() {
		opinion.BeginTime = time.Unix(0, 0)
	}
	if opinion.EndTime.IsZero() {
		opinion.EndTime = time.Unix(0, 0)
	}

	var url0 = fmt.Sprintf("https://api.bilibili.com/x/web-interface/wbi/search/type?search_type=video&page=%d&page_size=42&pubtime_begin_s=%d&pubtime_end_s=%d&keyword=%s&__refresh__=true&ad_resource=5655&category_id=&context=&dynamic_offset=&&from_source=&highlight=0",
		opinion.Page,
		opinion.BeginTime.Unix(),
		opinion.EndTime.Unix(),
		opinion.Keyword)

	//url0 = "https://api.bilibili.com/x/web-interface/wbi/search/type?category_id=&search_type=video&ad_resource=5654&__refresh__=true&_extra=&context=&page=1&page_size=42&pubtime_begin_s=0&pubtime_end_s=0&from_source=&from_spmid=333.337&platform=pc&highlight=1&single_column=0&keyword=%E6%95%B0%E5%AD%A6%E6%9E%97%E8%80%81%E5%B8%88&qv_id=G5siyPleYmCBx7tY2Exr3K3S43G2Gi0b&source_tag=3&gaia_vtoken=&dynamic_offset=0&web_roll_page=1&web_location=1430654"
	parse, _ := url.Parse(url0)
	query, _ := client.WBI.SignQuery(parse.Query(), time.Now())
	res, err := client.Resty.R().Get("https://api.bilibili.com/x/web-interface/search/type?" + query.Encode())
	if err != nil {
		fmt.Println(err)
	}
	var results []Video
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	for _, i := range getArray(obj, "data.result") {
		var v = Video{}
		reader := bytes.NewReader([]byte(getString(i, "title")))
		root, _ := html.Parse(reader)
		find := goquery.NewDocumentFromNode(root)
		v.Title = find.Text()
		v.UName = getString(i, "author")
		v.UID = getInt64(i, "mid")
		v.Desc = getString(i, "description")
		v.Tags = getString(i, "tag")
		v.Duration = getString(i, "duration")
		v.Face = getString(i, "upic")
		v.Cover = getString(i, "pic")
		v.View = getInt(i, "play")
		v.BV = getString(i, "bvid")
		v.Favorite = getInt(i, "favorites")
		v.Like = getInt(i, "like")
		v.Danmaku = getInt(i, "danmaku")
		v.Aid = getInt64(i, "id")
		v.CreateAt = time.Unix(getInt64(i, "pubdate"), 0)
		raw, _ := json.Marshal(i)
		v.RawResponse = string(raw)
		results = append(results, v)
	}
	return results
}
