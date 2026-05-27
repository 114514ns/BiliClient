package bili

import (
	"fmt"
	"html"
	url2 "net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jhump/protoreflect/dynamic"
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
	Emotes      string

	View     int
	Danmakus int
}
type Reply struct {
	ID      int64
	UName   string
	Face    string
	UID     int64
	Text    string
	Time    time.Time
	Likes   int
	Replies int
	Reply   []Reply
	ReplyID int64
	Type    int
	OID     int64
	Root    int64
	Images  []string
	Emotes  map[string]string
	Member  bool
	Level   int8
	CardBg  *struct {
		ID     int
		Num    int
		Color  string
		Image  string
		Prefix string
		Name   string
	}
	Medal *struct {
		Name  string
		Level int8
	}
	Pendant *struct {
		ID    int
		Name  string
		Image string
	}
	Mention interface{}
}

var emptyCardBg = struct {
	ID     int
	Num    int
	Color  string
	Image  string
	Prefix string
	Name   string
}{}
var emptyMedal = struct {
	Name  string
	Level int8
}{}
var emptyPendant = struct {
	ID    int
	Name  string
	Image string
}{}

type Article struct {
	ID        int
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
	Images    string
}
type Location struct {
	Address  string
	Describe string
}

type FaceMap struct {
	UID   int64
	Face  string
	UName string
	Bio   string
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
	Announce    string
}

type ActionItem struct {
	UName    string
	UID      int64
	TargetID int64
	OID      int64
}

// 获取用户收藏夹列表
func (client *BiliClient) GetFavList(user string, page int) map[string]string {
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

// 获取合集
func (client *BiliClient) getUserCollection0(uid int64, page int) []Collection {
	var url = fmt.Sprintf("https://api.bilibili.com/x/polymer/web-space/seasons_series_list?mid=%d&page_size=20&page_num=%d", uid, page)

	var u0, _ = url2.Parse(url)

	query, _ := client.WBI.SignQuery(u0.Query(), time.Now())

	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/polymer/web-space/seasons_series_list?" + query.Encode())

	var obj map[string]interface{}

	json.Unmarshal(res.Body(), &obj)

	var m = make([]Collection, 0)

	for _, item := range getArray(obj, "data.items_lists.seasons_list") {
		var c Collection
		c.Title = getString(item, "meta.name")
		c.Type = "season"
		c.ID = getInt64(item, "meta.season_id")
		c.Count = getInt(item, "meta.totsl")
		c.UID = uid

		m = append(m, c)
	}
	for _, item := range getArray(obj, "data.items_lists.series_list") {
		var c Collection
		c.Title = getString(item, "meta.name")
		c.Type = "series"
		c.ID = getInt64(item, "meta.series_id")
		c.UID = uid
		c.Count = getInt(item, "meta.totsl")
		m = append(m, c)
	}

	return m

}

func (client *BiliClient) GetUserCollection(uid int64) []Collection {
	var dst []Collection
	var page = 1
	for {
		var a = len(dst)
		dst = append(dst, client.getUserCollection0(uid, page)...)
		if len(dst)-a != 20 {
			break
		} else {
			page++
			time.Sleep(300 * time.Millisecond)
		}
	}
	return dst
}

// 获取合集内容
func (client *BiliClient) GetCollectionItems(id int64, typo string, page int, uid int64, ret ...bool) ([]Video, string, string) {
	var prefix = "https://api.bilibili.com/x/series/archives"

	var p = "series_id"

	if typo == "season" {
		//prefix = "https://api.bilibili.com/x/polymer/web-space/seasons_archives_list"
		prefix = "https://api.bilibili.com/x/space/seasons/archives/list"
		p = "season_id"
	}

	var desc = ""

	var u = fmt.Sprintf("%s?mid=%d&%s=%d&ps=%d&pn=%d&sort_reverse=false&sort=desc", prefix, uid, p, id, 20, page)
	if typo == "season" {
		u = strings.Replace(u, "&ps", "&page_size", 1)
		u = strings.Replace(u, "pn", "page_num", 1)
	}
	var u0, _ = url2.Parse(u)
	query, _ := client.WBI.SignQuery(u0.Query(), time.Now())
	res, _ := client.Resty.R().SetHeader("Cookie", fmt.Sprintf("buvid3=%sinfoc", uuid.New())).Get(prefix + "?" + query.Encode())
	var obj map[string]interface{}

	json.Unmarshal(res.Body(), &obj)

	var m = make([]Video, 0)
	desc = getString(obj, "data.meta.description")
	for _, item := range getArray(obj, "data.archives") {
		var v Video
		v.BV = getString(item, "bvid")
		v.Title = getString(item, "title")
		v.Duration = FormatDuration(getInt(item, "duration"))
		v.CreateAt = time.Unix(int64(getInt(item, "pubdate")), 0).Add(time.Hour * 8)
		v.Aid = getInt64(item, "aid")
		v.UID = uid
		m = append(m, v)
	}
	var title = getString(obj, "data.meta.name")
	if title == "" {

		r0, _ := client.Resty.R().Get(fmt.Sprintf("https://api.bilibili.com/x/series/series?series_id=%d", id))
		var o map[string]interface{}
		json.Unmarshal(r0.Body(), &o)
		title = getString(o, "data.meta.name")

	}

	if typo != "" || len(ret) != 0 {
		return m, title, desc
	} else {
		return client.GetCollectionItems(id, "season", page, uid, true)
	}

}

// 获取用户关注
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

// 分页获取用户关注
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

// 获取用户动态
func (client *BiliClient) GetDynamicsByUser(user int64, includeForward bool, offset0 ...string) ([]Dynamic, string) {
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

		return client.GetDynamicsByUser(user, includeForward, offset0...)
	}
	var archives = parseDynamic(obj, (offset == "" || offset == "-480"), includeForward)
	return archives, getString(obj, "data.offset")
}

// 根据oid获取动态
func (client *BiliClient) GetDynamicDetail(oid int64) []Dynamic {
	parse, _ := url2.Parse("https://api.bilibili.com/x/polymer/web-dynamic/desktop/v1/detail?build=11705&platform=web&device=linux&mobi_app=pc_electron&id=" + toString(oid))
	newQuery, _ := client.WBI.SignQuery(parse.Query(), time.Now())
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/polymer/web-dynamic/desktop/v1/detail?" + newQuery.Encode())
	var obj map[string]interface{}
	fmt.Println(res.String())
	json.Unmarshal(res.Body(), &obj)
	archives := parseDynamic0(obj["data"].(map[string]interface{})["item"])
	return archives
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
			var emotes = make(map[string]string)
			for _, i := range getArray(i2, "module_desc.rich_text_nodes") {
				if getString(i, "type") == "RICH_TEXT_NODE_TYPE_EMOJI" {
					emotes[getString(i, "text")] = getString(i, "emoji.icon_url")
				}
			}
			marshal, _ := json.Marshal(emotes)
			dynamic.Emotes = string(marshal)

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
				dynamic.View = toInt(getString(i2, "module_dynamic.dyn_archive.stat.play"))
				dynamic.Danmakus = toInt(getString(i2, "module_dynamic.dyn_archive.stat.danmaku"))
			}
			if typo == "MDL_DYN_TYPE_FORWARD" {
				dynamic.ForwardFrom = toInt64(getString(i2, "module_dynamic.dyn_forward.item.id_str"))
				result = append(result, parseDynamic0(i2.(map[string]interface{})["module_dynamic"].(map[string]interface{})["dyn_forward"].(map[string]interface{})["item"])...)
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

func parseDynamic(item interface{}, isFirst bool, includeForward bool) []Dynamic {
	var result []Dynamic
	for _, m := range item.(map[string]interface{})["data"].(map[string]interface{})["items"].([]interface{}) {

		parse := parseDynamic0(m)

		dyn := parse[0]
		dyn.RawResponse, _ = json.ConfigFastest.MarshalToString(&m)
		dyn.Forwarded = false
		result = append(result, dyn)
		if len(parse) == 2 {
			if includeForward {
				parse[1].Forwarded = true
				result = append(result, parse[1])
			}

		}
	}
	return result
}
func parseComment(reply ReplyInternalResponse) Reply {
	var comment = Reply{}
	comment.UID = reply.Mid
	comment.Text = reply.Content.Message
	comment.UName = reply.Member.Uname
	comment.Face = reply.Member.Avatar
	comment.ReplyID = reply.ReplyID
	comment.Likes = reply.Like

	comment.Replies = len(reply.Replies)
	comment.Time = time.Unix(int64(reply.Ctime), 0)
	return comment
}
func parseRPCComment(obj interface{}) Reply {
	var comment = Reply{}
	comment.ID = toInt64(getString(obj, "id"))
	comment.UID = toInt64(getString(obj, "mid"))
	comment.OID = toInt64(getString(obj, "oid"))
	comment.Time = time.Unix(toInt64(getString(obj, "ctime")), 0)
	comment.Text = getString(obj, "content.message")
	comment.UName = getString(obj, "member.name")
	comment.Face = getString(obj, "member.face")
	reply, ok := obj.(map[string]interface{})["replies"].([]interface{})
	comment.Reply = make([]Reply, 0)
	if getString(obj, "like") != "" {
		comment.Likes = int(toInt64(getString(obj, "like")))
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
	comment.ReplyID = toInt64(getString(obj, "root"))
	for _, i := range getArray(obj, "content.pictures") {
		comment.Images = append(comment.Images, getString(i, "imgSrc"))
	}
	comment.Mention = getObject(obj, "content.atNameToMid", "m").m
	var emotes = make(map[string]string)
	for k, v := range getObject(obj, "content.emotes", "").m {
		emotes[k] = getString(v, "url")
	}
	comment.Emotes = emotes
	comment.Level = int8(toInt(getString(obj, "member.level")))
	comment.Member = getString(obj, "member.vipType") == "1"

	var cardBg = (getObject(obj, "memberV2.userSailing.cardBg", "m").m)
	comment.CardBg = &emptyCardBg
	comment.CardBg.Num = toInt(getString(cardBg, "fan.number"))
	comment.CardBg.Prefix = getString(cardBg, "fan.numPrefix")
	comment.CardBg.Color = getString(cardBg, "fan.color")
	comment.CardBg.Name = getString(cardBg, "name")
	comment.CardBg.ID = toInt(getString(cardBg, "id"))
	comment.CardBg.Image = getString(cardBg, "image")
	if comment.CardBg.Num == 0 {
		comment.CardBg = nil
	}
	var pandent = getObject(obj, "memberV2.userSailing.pendent", "m").m
	comment.Pendant = &emptyPendant
	comment.Pendant.Name = getString(pandent, "name")
	comment.Pendant.Image = getString(pandent, "image")
	comment.Pendant.ID = getInt(pandent, "id")

	if comment.Pendant.Name == "" {
		comment.Pendant = nil
	}

	var medal = getObject(obj, "memberV2.medal", "m").m
	comment.Medal = &emptyMedal
	comment.Medal.Name = getString(medal, "medal_name")
	comment.Medal.Level = int8(getInt(medal, "level"))

	if comment.Medal.Name == "" {
		comment.Medal = nil
	}

	return comment

}
func (client *BiliClient) GetComment(oid int64, cursor string, type0 int, mode int) ([]Reply, string) {
	u := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/wbi/main?oid=%d&type=%d&mode=%d&pagination_str=%s", oid, type0, mode, fmt.Sprintf("{\"offset\":\"%s\"}", cursor))
	url, _ := url2.Parse(u)
	signed, _ := client.WBI.SignQuery(url.Query(), time.Now())
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/v2/reply/wbi/main?" + signed.Encode())
	obj := CommentResponse{}
	json.Unmarshal(res.Body(), &obj)
	var list = make([]Reply, 0)
	for _, reply := range obj.Data.Replies {
		var comment = Reply{}
		comment.UID = reply.Mid
		comment.Text = reply.Content.Message
		comment.ReplyID = reply.ReplyID
		comment.UName = reply.Member.Uname
		comment.Face = reply.Member.Avatar
		comment.Likes = reply.Like
		comment.ID = reply.ReplyID
		comment.Type = type0
		comment.Replies = reply.Count
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Reply = make([]Reply, 0)
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

// 获取评论
func (client *BiliClient) GetCommentRPC(oid int64, cursor string, type0 int, sort ...ReplySort) ([]Reply, string) {

	var template = `
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
	var processed = fmt.Sprintf(template, oid, type0, mode, cursor)
	msg := dynamic.NewMessage(ProtoMap[ProtoType.Reply_MainListReq])
	msg.UnmarshalJSON([]byte(processed))
	payload, _ := msg.Marshal()

	res, err := client.Resty.R().SetBody(payload).Post("https://app.bilibili.com/bilibili.main.community.reply.v1.Reply/MainList")
	if err != nil {
		fmt.Println(err)
	}
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)

	var array []Reply
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
					//array = append(array, subs...)
					o.Reply = append(o.Reply, subs...)
				}
				if len(subs) == 0 || offset == "" {
					break
				}
			}
		} else if len(o.Reply) != 0 {
			/*
				for _, comment := range o.Reply {
					array = append(array, comment)
				}

			*/
		}

		array = append(array, o)
	}

	if getBool(obj, "cursor.isEnd") {
		gotCursor = ""
	}
	return array, gotCursor

	//client.Resty.R().SetBody()
}

// 获取二级评论
func (client *BiliClient) GetReplyRPC(oid int64, root int64, type0 int, sort ...ReplySort) ([]Reply, string) {
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
	msg := dynamic.NewMessage(ProtoMap[ProtoType.Reply_DetailListReq])
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

// 获取二级评论
func (client *BiliClient) GetReply(oid int64, root int64, cursor string, type0 int) ([]Reply, string) {
	u := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/detail?oid=%d&type=%d&root=%d&web_location=333.1365&pagination_str=%s", oid, type0, root, url2.QueryEscape(fmt.Sprintf("{\"offset\":\"%s\"}", cursor)))
	obj := CommentResponse{}
	res, _ := client.Resty.R().Get(u)
	json.Unmarshal(res.Body(), &obj)
	var list = make([]Reply, 0)
	var raw = getArray(res.Body(), "data.root.replies")
	for i, reply := range obj.Data.Root.Replies {
		var comment = Reply{}
		comment.UID = reply.Mid
		comment.Text = reply.Content.Message
		comment.UName = reply.Member.Uname
		comment.Face = reply.Member.Avatar
		comment.Likes = reply.Like
		comment.ID = reply.ReplyID
		comment.Replies = len(reply.Replies)
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Type = type0
		comment.ReplyID = root
		comment.OID = oid
		comment.Root = root
		comment.Level = int8(reply.Member.LevelInfo.CurrentLevel)
		var cardBg = (getObject(raw[i], "member.user_sailing_v2.card_bg", "m").m)
		comment.CardBg = &emptyCardBg
		comment.CardBg.Num = getInt(cardBg, "fans.num")
		comment.CardBg.Prefix = getString(cardBg, "fans.num_prefix")
		comment.CardBg.Color = getString(cardBg, "fans.num_color")
		comment.CardBg.Name = getString(cardBg, "name")
		comment.CardBg.ID = getInt(cardBg, "id")
		comment.CardBg.Image = getString(cardBg, "image")
		if comment.CardBg.Num == 0 {
			comment.CardBg = nil
		}
		var pandent = getObject(raw[i], "member.user_sailing_v2.pendent", "m").m
		comment.Pendant = &emptyPendant
		comment.Pendant.Name = getString(pandent, "name")
		comment.Pendant.Image = getString(pandent, "image")
		comment.Pendant.ID = getInt(pandent, "id")

		if comment.Pendant.Name == "" {
			comment.Pendant = nil
		}

		var medal = getObject(raw[i], "member.fans_detail.medal", "m").m
		comment.Medal = &emptyMedal
		comment.Medal.Name = getString(medal, "medal_name")
		comment.Medal.Level = int8(getInt(medal, "level"))

		if comment.Medal.Name == "" {
			comment.Medal = nil
		}

		var emotes = make(map[string]string)
		for k, v := range getObject(raw[i], "content.emotes", "m").m {
			emotes[k] = getString(v, "url")
		}
		for _, j := range getArray(raw[i], "content.pictures") {
			comment.Images = append(comment.Images, getString(j, "imgSrc"))
		}
		comment.Mention = getObject(raw[i], "content.at_name_to_mid_str", "m").m
		comment.Emotes = (emotes)
		list = append(list, comment)
	}

	var off = obj.Data.Cursor.PaginationReply.NextOffset
	if getBool(obj, "data.cursor.isEnd") {
		off = ""
	}
	return list, off

}

// 获取二级评论
func (comment *Reply) GetReply(offset string, client *BiliClient) ([]Reply, string) {
	return client.GetReply(comment.OID, comment.ReplyID, offset, comment.Type)
}

// 获取评论
func (archive *Dynamic) GetComments(cursor string, client *BiliClient, sort ...ReplySort) ([]Reply, string) {
	t := CommentType.Dynamic
	if archive.BV != "" {
		t = CommentType.Video
	}
	var array []Reply
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

// 设置自己主页公告
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

// 粉丝牌牌
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

// 获取粉丝/关注的数量
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

// 获取比较详细的用户信息
func (client *BiliClient) GetUser(id int64) (User, int) {
	var wg sync.WaitGroup
	var user = User{}
	wg.Add(2)
	wg.Go(func() {
		res, _ := client.Resty.R().Get("https://api.bilibili.com/x/web-interface/card?mid=" + toString(id))
		var obj interface{}
		json.Unmarshal(res.Body(), &obj)
		user.UID = id
		/*
			if getInt(obj, "code") == -404 {
				return user, 3
			}
			if getInt(obj, "code") == -352 {
				return user, 2
			}
			if getString(obj, "message") != "OK" {
				return user, 2
			}

		*/
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
		wg.Done()
	})
	wg.Go(func() {

		res, _ := client.Resty.R().Get("https://api.bilibili.com/x/space/notice?mid=" + toString(id))
		var obj map[string]interface{}
		json.Unmarshal(res.Body(), &obj)
		user.Announce = getString(obj, "data")

		wg.Done()
	})

	wg.Wait()
	return user, 1
}

// 批量获取头像
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

// 获取头像
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

// 解析专栏
func (client *BiliClient) GetArticle(cv int64, callback func(string), rawResponse func(string2 string)) Article {
	if cv > 1500000000 {
		cv = int64(client.Opus2Cid(cv))
	}
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
		article.ID = int(cv)
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
		for _, v := range getArray(obj, "data.image_urls") {
			article.Images += v.(string) + ","
		}
	} else {
		if callback != nil {
			callback("risk")
		}
	}
	return article
}

func (client *BiliClient) GetUserArticle(uid int64, off string) []Article {
	var dst []Article
	res, _ := client.Resty.R().Get(fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/opus/feed/space?host_mid=%d&page=3&offset=%s&type=article", uid, off))
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	for _, i := range getArray(obj, "data.items") {
		dst = append(dst, Article{
			DynamicID: toInt64(getString(i, "opus_id")),
		})
	}
	return dst
}

func (client *BiliClient) Opus2Cid(oid int64) int {
	var msg = dynamic.NewMessage(ProtoMap[ProtoType.OpusDetailReq])
	var template = `
{"oid":%d,"pattern":"outer","share_mode":8}
`
	template = fmt.Sprintf(template, oid)

	msg.UnmarshalJSON([]byte(template))
	payload, e := msg.Marshal()
	if e != nil {
		fmt.Println(e)
	}
	response, _ := client.Resty.R().SetBody(payload).Post("https://app.bilibili.com/bilibili.app.dynamic.v2.Opus/OpusDetail")
	var obj map[string]interface{}
	json.Unmarshal(response.Body(), &obj)
	return toInt(getString(obj, "opusItem.oid"))
}

// 需要登录
func (client *BiliClient) GetReactionList(oid int64, page int) []ActionItem {
	res, _ := client.Resty.R().Get(fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/detail/reaction?id=%d&offset={\"page\":%d}", oid, page))
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	var results []ActionItem
	for _, i := range getArray(obj, "data.items") {
		var item ActionItem
		item.OID = oid
		item.UName = getString(i, "name")
		item.UID = toInt64(getString(i, "mid"))
		results = append(results, item)
	}
	if len(results) == 0 {
		time.Now()
	}
	return results
}

func (client *BiliClient) GetDynamicByUIDRPC(uid int64, page int) ([]Dynamic, bool) {

	var jsonStr = `
			{
			    "host_uid": %d,
				"page": %d,
				"local_time":8,
				"from":"space"

			}
			`
	var req = fmt.Sprintf(jsonStr, uid, page)

	msg := dynamic.NewMessage(ProtoMap[ProtoType.DynSpaceReq])
	err := msg.UnmarshalJSON([]byte(req))
	//err := msg.UnmarshalJSON([]byte(req))
	if err != nil {
		fmt.Println(err)
	}
	payload, _ := msg.Marshal()
	//payload, _ = os.ReadFile("DynSpace.")
	response, _ := client.Resty.R().SetBody(payload).Post("https://app.bilibili.com/bilibili.app.dynamic.v2.Dynamic/DynSpace")
	var obj map[string]interface{}
	json.Unmarshal(response.Body(), &obj)

	return nil, false
}

// 点赞列表
// 2是普通图文类型，8是视频类型
// 不支持DYNAMIC_TEXT,
func (client *BiliClient) GetLikeListRPC(oid int64, typo int, rid int64, page int, offset int64) ([]ActionItem, bool) {
	var results []ActionItem
	var jsonStr = `
		{
		    "dynamic_id": %d,
		    "dyn_type": %d,
		    "rid": %d,
			"page": %d
			
		}
		`

	var req = fmt.Sprintf(jsonStr, oid, typo, rid, page)
	msg := dynamic.NewMessage(ProtoMap[ProtoType.LikeListReq])
	err := msg.UnmarshalJSON([]byte(req))
	if err != nil {
		fmt.Println(err)
	}
	payload, _ := msg.Marshal()
	response, _ := client.Resty.R().SetBody(payload).Post("https://app.bilibili.com/bilibili.app.dynamic.v2.Dynamic/LikeList")
	var obj map[string]interface{}
	json.Unmarshal(response.Body(), &obj)
	for _, i := range getArray(obj, "list") {
		var o ActionItem
		o.UID = toInt64(getString(i, "mid"))
		o.OID = oid
		o.UName = getString(i, "author.name")
		results = append(results, o)
	}
	return results, getBool(obj, "hasMore")
}

// 获取当前ip
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

func (client *BiliClient) GetUserLocation(uid int64) string {
	if client.UID == 0 && client.Options.AccessKey == "" {
		res, _ := client.Resty.R().Get("https://api.bilibili.com//x/tv/dynamic/v2/single?mid=" + toString(uid))
		var o map[string]interface{}
		json.Unmarshal(res.Body(), &o)
		var videos = getArray(o, "data.list.cards")
		if videos == nil || len(videos) == 0 {
			return ""
		}
		var aid = getInt64(videos[0], "card_id")
		res, _ = client.Resty.R().Get("https://api.bilibili.com//x/tv/ugc/view?aid=" + toString(aid))
		var obj map[string]interface{}
		json.Unmarshal(res.Body(), &obj)
		return getString(obj, "data.pub_location")
	} else {
		res, _ := client.Resty.R().Get(appSign(fmt.Sprintf("https://app.bilibili.com/x/v2/space?access_key=%s&appkey=1d8b6e7d45233436&build=8770300&c_locale=zh-Hans_CN&channel=xiaomi&client_attr=0&disable_rcmd=0&fnval=84944&fnver=0&force_host=0&fourk=1&from=0&local_time=8&mobi_app=android&platform=android&player_net=1&qn=32&qn_policy=1&s_locale=zh-Hans_CN&ts=1770111164&vmid=%d&voice_balance=1", client.Options.AccessKey, uid)))
		var obj map[string]interface{}
		json.Unmarshal(res.Body(), &obj)
		for _, i := range getArray(obj, "data.card.space_tag") {
			if getString(i, "type") == "location" {
				return strings.Replace(getString(i, "title"), "IP属地：", "", 1)
			}
		}
		return ""
	}

}

func (client *BiliClient) GetCharge(uid int64) []ChargeInfo {

	var array []ChargeInfo

	var url = fmt.Sprintf("https://api.bilibili.com/x/upower/up/member/rank/v2?pn=1&privilege_type=10&ps=100&up_mid=%d", uid)

	res, _ := client.Resty.R().Get(url)
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	for _, o := range getArray(obj, "data.level_info") {
		array = append(array, ChargeInfo{
			Name:  getString(o, "name"),
			Price: float64(getInt(o, "price") / 100.0),
			Count: getInt(o, "member_total"),
		})
	}

	return array
}

func (client *BiliClient) SearchUser(keyword string) []User {
	res, _ := client.Resty.R().Get(fmt.Sprintf("https://app.bilibili.com/x/v2/car/search/type?keyword=%s&search_type=bili_user", url2.QueryEscape(keyword)))
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	var dst []User
	for _, i := range getArray(obj, "data.up_items") {
		dst = append(dst, User{
			UID:   getInt64(i, "mid"),
			UName: getString(i, "name"),
			Bio:   getString(i, "desc"),
			Face:  getString(i, "face"),
			Fans:  uint(getInt(i, "fans_count")),
		})
	}
	return dst
}

func GetArticle() {

}
