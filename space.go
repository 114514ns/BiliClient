package bili

import (
	"encoding/json"
	"fmt"
	"html"
	url2 "net/url"
	"strconv"
	"strings"
	"time"
)

type CommentType0 struct {
	Video   int
	Dynamic int
}

var CommentType = CommentType0{
	Video:   1,
	Dynamic: 17,
}

type Archive struct {
	Aid    int64
	UName  string
	UID    int64
	Images []string
	Type   string
	Title  string
	Text   string
	ID     string
	BV     string
}
type Comment struct {
	UName   string
	UID     int64
	Text    string
	Time    time.Time
	Like    int
	Replies int
	Avatar  string
	Reply   []Comment
	OID     int64
	ReplyID int64
	Type    int
}

type Article struct {
	UID       int64
	UName     string
	Title     string
	Text      string
	View      int
	Likes     int
	Coin      int
	Comments  int
	Forward   int
	Favourite int
	Tags      string
	DynamicID int64
	CreatedAt time.Time
}
type Location struct {
	Address  string
	Describe string
}

type FaceMap struct {
	UID   int64
	Face  string
	UName string
}

func (client *BiliClient) GetCollection(user string, page int) map[string]string {
	var url = fmt.Sprintf("https://api.bilibili.com/x/v3/fav/folder/created/list?ps=50&pn=%d&up_mid=%s", page, user)
	res, _ := client.Resty.R().Get(url)
	var list = CollectionList{}
	json.Unmarshal(res.Body(), &list)
	var m = make(map[string]string)
	for _, s := range list.Data.List {
		m[strconv.Itoa(s.ID)] = s.Title
	}
	return m
}
func (client *BiliClient) GetFollowing(user string, delay int) map[string]string {
	var m = make(map[string]string)
	page := 1
	for true {
		part := client.GetFollowingByPage(user, page)
		if len(part) == 0 {
			break
		}
		for s, s2 := range part {
			m[s] = s2
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
		page++
	}
	return m
}
func (client *BiliClient) GetFollowingByPage(user string, page int) map[string]string {
	resp, err := client.Resty.R().Get("https://line3-h5-mobile-api.biligame.com/game/center/h5/user/relationship/following_list?vmid=" + string(user) + "&ps=50&pn=" + strconv.Itoa(page))
	if err != nil {
		fmt.Println(err)
	}
	var list = FansList{}
	var m = make(map[string]string)
	json.Unmarshal(resp.Body(), &list)
	var users = list.Data.List
	for i := 0; i < len(users); i++ {
		m[users[i].Mid] = users[i].Uname
	}
	return m
}

func (client *BiliClient) GetDynamicsByUser(user int64, offset string) ([]Archive, string) {
	url := fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=%d&timezone_offset=-480&features=itemOpusStyle", user)
	u, _ := url2.Parse(url)
	signed, _ := client.WBI.SignQuery(u.Query(), time.Now())
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?" + signed.Encode())
	obj := UserDynamic{}
	json.Unmarshal(res.Body(), &obj)
	var archives = make([]Archive, 0)
	for _, item := range obj.Data.Items {
		p, _ := parseDynamc(item)
		archives = append(archives, p)
	}
	return archives, obj.Data.Offset
}
func parseDynamc(item DynamicItem) (Archive, Archive) {
	var Type = item.Type
	var orig = Archive{}
	var userName = item.Modules.ModuleAuthor.Name
	var archive = Archive{}
	archive.UName = userName
	archive.UID = item.Modules.ModuleAuthor.Mid
	archive.Aid, _ = strconv.ParseInt(item.Base.CommentID, 10, 64)
	if Type == "DYNAMIC_TYPE_FORWARD" { //转发
		archive.Type = "f"
		archive.ID = item.IDStr
		var txt = ""
		for _, node := range item.Modules.ModuleDynamic.Desc.Nodes {
			txt = txt + node.Text
			txt = txt + "\n"
		}
		//orig, _ = parseDynamc(*item.Orig, false)
		archive.Text = txt
	} else if Type == "DYNAMIC_TYPE_AV" { //发布视频
		archive.Type = "v"
		archive.ID = item.IDStr
		archive.BV = item.Modules.ModuleDynamic.Major.Archive.Bvid
		archive.Title = item.Modules.ModuleDynamic.Major.Archive.Title
	} else if Type == "DYNAMIC_TYPE_DRAW" { //图文
		archive.Type = "i"
		archive.ID = item.IDStr
		for _, pic := range item.Modules.ModuleDynamic.Major.Opus.Pics {
			archive.Images = append(archive.Images, pic.URL)
		}
		//archive.Text = item.Modules.ModuleDynamic.Major.Desc.Text
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text

	} else if Type == "DYNAMIC_TYPE_WORD" { //文字
		archive.Type = "t"
		archive.ID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
	} else if Type == "DYNAMIC_TYPE_LIVE_RCMD" {

	} else if Type == "DYNAMIC_TYPE_COMMON_SQUARE" {

	} else {
		archive.Type = Type
		archive.ID = item.IDStr
		archive.Text = item.Modules.ModuleDynamic.Major.Opus.Summary.Text
	}
	return archive, orig
}
func parseComment(reply ReplyInternalResponse) Comment {
	var comment = Comment{}
	comment.UID = reply.Mid
	comment.Text = reply.Content.Message
	comment.UName = reply.Member.Uname
	comment.Avatar = reply.Member.Avatar
	comment.ReplyID = reply.ReplyID
	comment.Like = reply.Like

	comment.Replies = len(reply.Replies)
	comment.Time = time.Unix(int64(reply.Ctime), 0)
	return comment
}
func (client *BiliClient) GetComment(oid int64, cursor string, type0 int) []Comment {
	u := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/wbi/main?oid=%d&type=%d&mode=%d&pagination_str=%s", oid, type0, 3, fmt.Sprintf("{\"offset\":\"%s\"}", cursor))
	url, _ := url2.Parse(u)
	signed, _ := client.WBI.SignQuery(url.Query(), time.Now())
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/v2/reply/wbi/main?" + signed.Encode())
	obj := CommentResponse{}
	json.Unmarshal(res.Body(), &obj)
	var list = make([]Comment, 0)
	for _, reply := range obj.Data.Replies {
		var comment = Comment{}
		comment.UID = reply.Mid
		comment.Text = reply.Content.Message
		comment.ReplyID = reply.ReplyID
		comment.UName = reply.Member.Uname
		comment.Avatar = reply.Member.Avatar
		comment.Like = reply.Like
		comment.OID = oid
		comment.Type = type0
		comment.Replies = reply.Count
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Reply = make([]Comment, 0)
		for _, response := range reply.Replies {
			comment.Reply = append(comment.Reply, parseComment(response))
			comment.OID = oid
			comment.Type = type0
		}

		list = append(list, comment)

	}
	return list
}
func (client *BiliClient) GetReply(oid int64, root int64, page int, type0 int) []Comment {
	u := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/reply?oid=%d&type=%d&root=%d&ps=10&pn=%d&web_location=333.1365", oid, type0, root, page)
	obj := CommentResponse{}
	res, _ := client.Resty.R().Get(u)
	json.Unmarshal(res.Body(), &obj)
	var list = make([]Comment, 0)
	for _, reply := range obj.Data.Replies {
		var comment = Comment{}
		comment.UID = reply.Mid
		comment.Text = reply.Content.Message
		comment.UName = reply.Member.Uname
		comment.Avatar = reply.Member.Avatar
		comment.Like = reply.Like
		comment.OID = oid
		comment.Replies = len(reply.Replies)
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Type = type0

		list = append(list, comment)
	}

	return list

}
func (comment *Comment) GetReply(page int, client *BiliClient) []Comment {
	return client.GetReply(comment.OID, comment.ReplyID, page, CommentType.Dynamic)
}
func (archive *Archive) GetComments(cursor string, client *BiliClient) []Comment {
	t := CommentType.Dynamic
	if archive.BV != "" {
		t = CommentType.Video
	}
	return client.GetComment(archive.Aid, cursor, t)
}
func (client *BiliClient) SetAnnouce(content string) {
	split := strings.Split(client.Cookie, ";")
	jct := ""
	for _, s := range split {
		if strings.Contains(s, "bili_jct=") {
			jct = strings.Replace(s, "bili_jct=", "", 1)
		}
	}
	jct = jct[1:len(jct)]
	body := fmt.Sprintf("notice=%s&csrf=%s", url2.QueryEscape(content), jct)
	var req = client.Resty.R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody(body)
	req.Post("https://api.bilibili.com/x/space/notice/set")

}

func (client *BiliClient) GetFansMedal(id string) []Medal {
	u := "https://api.live.bilibili.com/xlive/web-ucenter/user/MedalWall?target_id=" + id
	res, _ := client.Resty.R().Get(u)
	obj := FansWallListResponse{}
	json.Unmarshal(res.Body(), &obj)
	list := make([]Medal, 0)
	for _, s := range obj.Data.List {
		medal := Medal{}
		medal.Level = int8(s.MedalInfo.Level)
		medal.Name = s.MedalInfo.MedalName
		medal.GuardLevel = int8(s.MedalInfo.GuardLevel)
		medal.Color = s.MedalInfo.UinfoMedal.V2MedalColorStart
		medal.LiverUID = s.MedalInfo.TargetId
		medal.LiverName = s.MedalInfo.MedalName
		list = append(list, medal)

	}
	return list
}
func (client *BiliClient) GetStats(uid int64) map[string]int {
	var m = make(map[string]int)
	u := fmt.Sprintf("https://api.bilibili.com/x/space/upstat?mid=%d", uid)
	res, _ := client.Resty.R().Get(u)
	nextNumber(res.String(), strings.Index(res.String(), "view"))
	m["view"] = int(nextNumber(res.String(), strings.Index(res.String(), "view")))
	m["likes"] = int(nextNumber(res.String(), strings.Index(res.String(), "likes")))
	u = fmt.Sprintf("Https://api.bilibili.com/x/relation/stat?vmid=%d", uid)
	res, _ = client.Resty.R().Get(u)
	m["follower"] = int(nextNumber(res.String(), strings.Index(res.String(), "follower")))
	m["following"] = int(nextNumber(res.String(), strings.Index(res.String(), "following")))
	return m

}
func (client *BiliClient) getUser(id int64) {
	u := appendWts("https://api.bilibili.com/x/space/wbi/acc/info?mid="+toString(id), client)
	res, _ := client.Resty.R().Get(u)
	fmt.Println(res.String())
}

func (client *BiliClient) BatchGetFace(id []int64) []FaceMap {

	var s = "https://api.live.bilibili.com/xlive/fuxi-interface/UserService/getUserInfo?_ts_rpc_args_=[["
	for i, i2 := range id {
		s = s + strconv.FormatInt(i2, 10)
		if i != len(id)-1 {
			s = s + ","
		}
	}
	s = s + `],true,""]`
	res, _ := client.Resty.R().Get(s)
	var m map[string]interface{}
	json.Unmarshal(res.Body(), &m)
	var result []FaceMap
	for s2, i := range m["_ts_rpc_return_"].(map[string]interface{})["data"].(map[string]interface{}) {
		result = append(result, FaceMap{UID: toInt64(s2), Face: "https:" + getString(i, "face"), UName: getString(i, "uname")})
	}
	return result
}
func (client *BiliClient) GetFace(id int64) string {
	var s = "https://api.live.bilibili.com/xlive/fuxi-interface/UserService/getUserInfo?_ts_rpc_args_=[[" + strconv.FormatInt(id, 10)
	s = s + `],true,""]`
	res, _ := client.Resty.R().Get(s)
	type Response struct {
		TsRpcReturn struct {
			Data map[string]struct {
				UID   string `json:"uid"`
				UName string `json:"uname"`
				Face  string `json:"face"`
			} `json:"data"`
		} `json:"_ts_rpc_return_"`
	}

	var r = Response{}
	json.Unmarshal(res.Body(), &r)

	return "https:" + r.TsRpcReturn.Data[strconv.FormatInt(id, 10)].Face
}
func (client *BiliClient) GetArticle(cv int64, callback func(string), rawResponse func(string2 string)) Article {

	start := time.Now()
	res, _ := client.Resty.R().Get(fmt.Sprintf("https://api.bilibili.com/x/article/view?id=%s", strconv.FormatInt(cv, 10)))
	fmt.Println(time.Since(start))
	article := Article{}
	var obj interface{}
	if rawResponse != nil {
		rawResponse(res.String())
	}
	json.Unmarshal(res.Body(), &obj)
	if obj == nil {
		if callback != nil {
			callback("risk")
			return Article{}
		}
	}
	code := getInt(obj, "code")
	if code == -404 {
		if callback != nil {
			callback("no")
		}
		return article
	} else if code == 0 {
		article.UID = getInt64(obj, "data.author.mid")
		article.UName = getString(obj, "data.author.name")
		article.CreatedAt = time.Unix(int64(getInt64(obj, "data.publish_time")), 0)
		article.View = getInt(obj, "data.stats.view")
		article.Coin = getInt(obj, "data.stats.coin")
		article.Forward = getInt(obj, "data.stats.share")
		article.Comments = getInt(obj, "data.stats.reply")
		article.Likes = getInt(obj, "data.stats.like")
		article.Title = getString(obj, "data.title")
		article.Favourite = getInt(obj, "data.stats.favorite")
		article.Text = html.UnescapeString(getString(obj, "data.content"))
		article.DynamicID, _ = strconv.ParseInt(getString(obj, "data.dyn_id_str"), 10, 64)
		tags, ok := obj.(map[string]interface{})["data"].(map[string]interface{})["tags"].([]interface{})
		if ok {
			t := ""
			for _, tag := range tags {
				t = t + getString(tag, "name") + ","
			}
			article.Tags = t[:len(t)-1]
		}
	} else {
		if callback != nil {
			callback("risk")
		}
	}

	return article

}

func (client *BiliClient) GetLocation(ip ...string) Location {
	u := "https://api.bilibili.com/x/web-interface/zone"
	if len(ip) != 0 {
		u = "https://api.live.bilibili.com/client/v1/Ip/getInfoNew?ip=" + ip[0]
	}
	res, err := client.Resty.R().Get(u)
	var obj interface{}
	json.Unmarshal(res.Body(), &obj)

	if err != nil {
		fmt.Println(err)
	}
	result := Location{}
	result.Address = getString(obj, "data.addr")
	result.Describe = getString(obj, "data.country") + getString(obj, "data.province") + getString(obj, "data.city") + getString(obj, "data.isp")

	return result
}
