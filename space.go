package bili

import (
	"fmt"
	json "github.com/bytedance/sonic"
	"github.com/jhump/protoreflect/dynamic"
	"html"
	url2 "net/url"
	"sort"
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
	Dynamic: 11,
}

type Dynamic struct {
	Top         bool
	UName       string
	UID         int64
	Face        string
	Images      []string
	Type        string
	Title       string
	Text        string
	ID          int64
	BV          string
	Comments    int
	Like        int
	Forward     int
	CommentID   int64
	CommentType int
	CreateAt    time.Time
	ForwardFrom int64
	RawResponse string
	Forwarded   bool
}
type Comment struct {
	ID      int64
	UName   string
	Face    string
	UID     int64
	Text    string
	Time    time.Time
	Like    int
	Replies int
	Reply   []Comment
	ReplyID int64
	Type    int
	OID     int64
	Root    int64
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

type User struct {
	UName       string
	UID         int64
	Bio         string
	Face        string
	Fans        uint
	Level       int8
	VerifyType  int8
	Archives    uint
	Like        uint
	Verify      string
	RawResponse string
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

func (client *BiliClient) GetDynamicsByUser(user int64, offset0 ...string) ([]Dynamic, string) {
	offset := ""
	if len(offset0) != 0 {
		offset = offset0[0]
	}
	url := fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?host_mid=%d&offset=%s&web_location=electron", user, offset)
	u, _ := url2.Parse(url)
	signed, _ := client.WBI.SignQuery(u.Query(), time.Now())
	res, err := client.Resty.R().Get("https://api.bilibili.com/x/polymer/web-dynamic/desktop/v1/feed/space?" + signed.Encode())
	if err != nil {
		fmt.Println(err)
		return []Dynamic{}, "-1"
	}
	var obj interface{}
	json.Unmarshal(res.Body(), &obj)
	if getString(obj, "message") != "0" {
		fmt.Println(res.String())

		return client.GetDynamicsByUser(user, offset0...)
	}
	var archives = parseDynamic(obj, (offset == "" || offset == "-480"))
	return archives, getString(obj, "data.offset")
}
func (dyn *Dynamic) GetComment() {

}

func parseDynamic0(m interface{}) []Dynamic {
	var result []Dynamic
	var dynamic = Dynamic{}
	dynamic.ID = toInt64(getString(m, "id_str"))
	if dynamic.ID == 0 {
		dynamic.ID = getInt64(m, "id_str")
	}
	if dynamic.ID == 0 {
		time.Now()
	}
	dynamic.Type = getString(m, "type")
	_, ok := m.(map[string]interface{})["orig"]
	if ok {
		dynamic.ForwardFrom = toInt64(getString(m, "orig.id_str"))
	}

	for _, i2 := range m.(map[string]interface{})["modules"].([]interface{}) {
		switch getString(i2, "module_type") {

		case "MODULE_TYPE_AUTHOR":
			dynamic.CreateAt = time.Unix(getInt64(i2, "module_author.pub_ts"), 0)
			dynamic.UID = getInt64(i2, "module_author.user.mid")
			dynamic.UName = getString(i2, "module_author.user.name")
			dynamic.Face = getString(i2, "module_author.user.face")
			dynamic.Top = getBool(i2, "module_author.is_top")
		case "MODULE_TYPE_DESC":
			dynamic.Text = getString(i2, "module_desc.text")

		case "MODULE_TYPE_DYNAMIC":
			var images []string
			typo := getString(i2, "module_dynamic.type")
			if typo == "MDL_DYN_TYPE_DRAW" {
				for _, o := range i2.(map[string]interface{})["module_dynamic"].(map[string]interface{})["dyn_draw"].(map[string]interface{})["items"].([]interface{}) {
					images = append(images, getString(o, "src"))
				}
				dynamic.Images = images
			}
			if typo == "MDL_DYN_TYPE_ARCHIVE" {
				dynamic.BV = getString(i2, "module_dynamic.dyn_archive.bvid")
				dynamic.Images = []string{getString(i2, "module_dynamic.dyn_archive.cover")}
				dynamic.Title = getString(i2, "module_dynamic.dyn_archive.title")
			}
			if typo == "MDL_DYN_TYPE_FORWARD" {
				dynamic.ForwardFrom = toInt64(getString(i2, "module_dynamic.dyn_forward.item.id_str"))
				//result = append(result, parseDynamic0(i2.(map[string]interface{})["module_dynamic"].(map[string]interface{})["dyn_forward"].(map[string]interface{})["item"])...)
			}
		case "MODULE_TYPE_STAT":
			dynamic.Comments = getInt(i2, "module_stat.comment.count")
			dynamic.Forward = getInt(i2, "module_stat.forward.count")
			dynamic.Like = getInt(i2, "module_stat.like.count")
			dynamic.CommentID = toInt64(getString(i2, "module_stat.comment.comment_id"))
			dynamic.CommentType = getInt(i2, "module_stat.comment.comment_type")

		}

	}
	result = append(result, dynamic)
	sort.SliceStable(result, func(i, j int) bool {
		return true
	})
	return result
}

func parseDynamic(item interface{}, isFirst bool) []Dynamic {
	var result []Dynamic
	for _, m := range item.(map[string]interface{})["data"].(map[string]interface{})["items"].([]interface{}) {

		parse := parseDynamic0(m)

		dyn := parse[0]
		dyn.RawResponse, _ = json.ConfigFastest.MarshalToString(&m)
		dyn.Forwarded = false
		result = append(result, dyn)
		if len(parse) == 2 {
			parse[1].Forwarded = true
			result = append(result, parse[1])
		}
	}
	return result
}
func parseComment(reply ReplyInternalResponse) Comment {
	var comment = Comment{}
	comment.UID = reply.Mid
	comment.Text = reply.Content.Message
	comment.UName = reply.Member.Uname
	comment.Face = reply.Member.Avatar
	comment.ReplyID = reply.ReplyID
	comment.Like = reply.Like

	comment.Replies = len(reply.Replies)
	comment.Time = time.Unix(int64(reply.Ctime), 0)
	return comment
}
func parseRPCComment(obj interface{}) Comment {
	var comment = Comment{}
	comment.ID = toInt64(getString(obj, "id"))
	comment.UID = toInt64(getString(obj, "mid"))
	comment.OID = toInt64(getString(obj, "oid"))
	comment.Time = time.Unix(toInt64(getString(obj, "ctime")), 0)
	comment.Text = getString(obj, "content.message")
	comment.UName = getString(obj, "member.name")
	comment.Face = getString(obj, "member.face")
	reply, ok := obj.(map[string]interface{})["replies"].([]interface{})
	comment.Reply = make([]Comment, 0)
	if getString(obj, "like") != "" {
		comment.Like = int(toInt64(getString(obj, "like")))
	}
	if getString(obj, "count") != "" {
		comment.Replies = int(toInt64(getString(obj, "count")))
	}
	if ok {
		if comment.Replies == len(reply) {
			for _, i := range reply {
				comment.Reply = append(comment.Reply, parseRPCComment(i))
			}
		}

	}
	comment.ReplyID = toInt64(getString(obj, "id"))
	return comment

}
func (client *BiliClient) GetComment(oid int64, cursor string, type0 int, mode int) ([]Comment, string) {
	u := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/wbi/main?oid=%d&type=%d&mode=%d&pagination_str=%s", oid, type0, mode, fmt.Sprintf("{\"offset\":\"%s\"}", cursor))
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
		comment.Face = reply.Member.Avatar
		comment.Like = reply.Like
		comment.ID = reply.ReplyID
		comment.Type = type0
		comment.Replies = reply.Count
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Reply = make([]Comment, 0)
		if comment.Replies == len(reply.Replies) {
			for _, response := range reply.Replies {
				var sub = parseComment(response)
				comment.Reply = append(comment.Reply, sub)
				list = append(list, comment)
			}
		} else {
			var offset = ""
			for {
				subs, o := client.GetReply(oid, comment.ID, offset, type0)
				offset = o
				if len(subs) > 0 {
					list = append(list, subs...)
					comment.Reply = append(comment.Reply, subs...)
				}
				if len(subs) == 0 || offset == "" {
					break
				}
			}
			time.Now()
		}

		list = append(list, comment)

	}
	var off = obj.Data.Cursor.PaginationReply.NextOffset
	if off == "" || len(list) == 0 {
		time.Now()
	}
	return list, off
}

type ReplySort string

const (
	REPLY_SORT_TIME = "MAIN_LIST_TIME"
	REPLY_SORT_HOT  = "MAIN_LIST_HOT"
)

func UniqueByField[T any, K comparable](items []T, keyFunc func(T) K) []T {
	seen := make(map[K]struct{})
	var result []T

	for _, item := range items {
		key := keyFunc(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
func (client *BiliClient) GetCommentRPC(oid int64, cursor string, type0 int, sort ...ReplySort) ([]Comment, string) {

	var jsonStr = `
		{
		    "oid": "%d",
		    "type": "%d",
		    "extra": "{}",
		    "adExtra": "",
		    "filterTagName": "全部",
		    "mode": "%s",
		    "pagination": {
		        "offset": "%s"
		    }
		}
		`
	var mode = "MAIN_LIST_HOT"
	if len(sort) > 0 {
		mode = string(sort[0])
	}
	var processed = fmt.Sprintf(jsonStr, oid, type0, mode, cursor)
	msg := dynamic.NewMessage(protoMap[ProtoType.Reply_MainListReq])
	msg.UnmarshalJSON([]byte(processed))
	payload, _ := msg.Marshal()

	res, err := client.Resty.R().SetBody(payload).Post("https://app.bilibili.com/bilibili.main.community.reply.v1.Reply/MainList")
	if err != nil {
		fmt.Println(err)
	}
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)

	var array []Comment
	var gotCursor = getString(obj, "paginationReply.nextOffset")

	if obj["replies"] == nil {
		return array, ""
	}
	for _, i := range obj["replies"].([]interface{}) {

		var o = parseRPCComment(i)
		o.Type = type0
		o.OID = oid
		if len(o.Reply) != o.Replies {
			var offset = ""
			for {
				subs, of := client.GetReply(oid, o.ID, offset, type0)
				offset = of
				if len(subs) > 0 {
					array = append(array, subs...)
					o.Reply = append(o.Reply, subs...)
				}
				if len(subs) == 0 || offset == "" {
					break
				}
			}
		} else if len(o.Reply) != 0 {
			for _, comment := range o.Reply {
				array = append(array, comment)
			}
		}

		array = append(array, o)
	}

	if getBool(obj, "cursor.isEnd") {
		gotCursor = ""
	}
	return array, gotCursor

	//client.Resty.R().SetBody()
}
func (client *BiliClient) GetReplyRPC(oid int64, root int64, type0 int, sort ...ReplySort) ([]Comment, string) {
	var jsonStr = `
		{
		    "oid": %d,
		    "type": %d,
		    "root": %d,
		    "mode": "3",
			"extra":"\"\n{\"spmid\":\"united.player-video-detail.0.0\",\"from_spmid\":\"dt.dt.video.0\"}\""
		}
		`
	var processed = fmt.Sprintf(jsonStr, oid, type0, root)
	msg := dynamic.NewMessage(protoMap[ProtoType.Reply_DetailListReq])
	err := msg.UnmarshalJSON([]byte(processed))
	if err != nil {
		return nil, ""
	}
	payload, _ := msg.Marshal()

	res, err := client.Resty.R().SetBody(payload).Post("https://app.bilibili.com/bilibili.main.community.reply.v1.Reply/DetailList")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res.String())
	return nil, ""
}
func (client *BiliClient) GetReply(oid int64, root int64, cursor string, type0 int) ([]Comment, string) {
	u := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/detail?oid=%d&type=%d&root=%d&web_location=333.1365&pagination_str=%s", oid, type0, root, fmt.Sprintf("{\"offset\":\"%s\"}", cursor))
	obj := CommentResponse{}
	res, _ := client.Resty.R().Get(u)
	json.Unmarshal(res.Body(), &obj)
	var list = make([]Comment, 0)
	for _, reply := range obj.Data.Root.Replies {
		var comment = Comment{}
		comment.UID = reply.Mid
		comment.Text = reply.Content.Message
		comment.UName = reply.Member.Uname
		comment.Face = reply.Member.Avatar
		comment.Like = reply.Like
		comment.ID = reply.ReplyID
		comment.Replies = len(reply.Replies)
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Type = type0
		comment.ReplyID = root
		comment.OID = oid
		comment.Root = root
		list = append(list, comment)
	}

	var off = obj.Data.Cursor.PaginationReply.NextOffset
	if getBool(obj, "data.cursor.isEnd") {
		off = ""
	}
	return list, off

}
func (comment *Comment) GetReply(offset string, client *BiliClient) ([]Comment, string) {
	return client.GetReply(comment.OID, comment.ReplyID, offset, comment.Type)
}
func (archive *Dynamic) GetComments(cursor string, client *BiliClient, sort ...ReplySort) ([]Comment, string) {
	t := CommentType.Dynamic
	if archive.BV != "" {
		t = CommentType.Video
	}
	var array []Comment
	var offset = ""
	if client.UID == 0 {
		array, offset = client.GetCommentRPC(archive.CommentID, cursor, archive.CommentType, sort...)
	} else {
		array, offset = client.GetComment(archive.CommentID, cursor, t, 3)
	}
	var sum = 0
	for _, comment := range array {
		sum = sum + 1
		sum = sum + comment.Replies
	}
	if sum == archive.Comments {
		return array, ""
	} else {
		return array, offset
	}

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
func (client *BiliClient) GetUser(id int64) (User, int) {
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/web-interface/card?mid=" + toString(id))
	var obj interface{}
	json.Unmarshal(res.Body(), &obj)
	var user = User{}
	user.UID = id
	if getInt(obj, "code") == -404 {
		return user, 3
	}
	if getInt(obj, "code") == -352 {
		return user, 2
	}
	user.UName = getString(obj, "data.card.name")

	user.Level = int8(getInt64(obj, "data.card.level_info.current_level"))
	user.Face = getString(obj, "data.card.face")
	user.Bio = getString(obj, "data.card.sign")
	user.Archives = uint(getInt(obj, "data.archive_count"))
	user.Like = uint(getInt(obj, "data.like_num"))
	user.Fans = uint(getInt(obj, "data.follower"))
	user.Verify = strings.Replace(getString(obj, "data.card.official_verify.desc"), "、", ",", 1145)
	user.VerifyType = int8(getInt(obj, "data.card.official_verify.type"))
	user.RawResponse = res.String()
	return user, 1
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

	if !strings.Contains(res.String()[0:70], "_ts_rpc_return_") || strings.Contains(res.String()[0:70], "服务调用超时") {
		fmt.Println(res.String())
		time.Sleep(time.Second * 3)
		return client.BatchGetFace(id)

	}
	for s2, i := range m["_ts_rpc_return_"].(map[string]interface{})["data"].(map[string]interface{}) {
		result = append(result, FaceMap{UID: toInt64(s2), Face: "https:" + getString(i, "face"), UName: getString(i, "uname")})
	}
	if len(result) != 50 {
		fmt.Println(res.String())
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

	//start := time.Now()
	res, _ := client.Resty.R().Get(fmt.Sprintf("https://api.bilibili.com/x/article/view?id=%s", strconv.FormatInt(cv, 10)))
	//fmt.Println(time.Since(start))
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
