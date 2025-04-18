package bili

import (
	"encoding/json"
	"fmt"
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
	Replys  int
	Avatar  string
	Reply   []Comment
	ReplyID int64
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

func (client *BiliClient) GetDynamicsByUser(user string, offset string) ([]Archive, string) {
	url := "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset&host_mid=" + user + "&timezone_offset=-480&features=itemOpusStyle"
	u, _ := url2.Parse(url)
	signed, _ := client.WBI.SignQuery(u.Query(), time.Now())
	res, _ := client.Resty.R().Get("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?" + signed.Encode())
	obj := UserDynamic{}
	json.Unmarshal(res.Body(), &obj)
	var archives = make([]Archive, 0)
	for _, item := range obj.Data.Items {
		p, _ := ParseDynamic(item)
		archives = append(archives, p)
	}
	return archives, obj.Data.Offset
}
func ParseDynamic(item DynamicItem) (Archive, Archive) {
	var Type = item.Type
	var orig = Archive{}
	var userName = item.Modules.ModuleAuthor.Name
	var archive = Archive{}
	archive.UName = userName
	archive.UID = item.Modules.ModuleAuthor.Mid
	archive.Aid, _ = strconv.ParseInt(item.IDStr, 10, 64)
	if Type == "DYNAMIC_TYPE_FORWARD" { //转发
		archive.Type = "f"
		archive.ID = item.IDStr
		var txt = ""
		for _, node := range item.Modules.ModuleDynamic.Desc.Nodes {
			txt = txt + node.Text
			txt = txt + "\n"
		}
		//orig, _ = ParseDynamic(*item.Orig, false)
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
	comment.Replys = len(reply.Replies)
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
		comment.Replys = reply.Count
		comment.Time = time.Unix(int64(reply.Ctime), 0)
		comment.Reply = make([]Comment, 0)
		for _, response := range reply.Replies {
			comment.Reply = append(comment.Reply, parseComment(response))
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
		comment.Replys = len(reply.Replies)
		comment.Time = time.Unix(int64(reply.Ctime), 0)

		list = append(list, comment)
	}

	return list

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
