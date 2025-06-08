package bili

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"net/url"
	"strconv"
	"time"
)

type Medal struct {
	Name          string `json:"medal_name"`
	Level         int8   `json:"level"`
	ColorDec      int    `json:"medal_color_start"`
	ColorInternal string `json:"v2_medal_color_start"`
	GuardLevel    int8
	Color         string
	LiverUID      int64
	LiverName     string
}
type LiveUser struct {
	UID   int64  `json:"uid"`
	Name  string `json:"name"`
	Face  string `json:"face"`
	Guard int8   `json:"guard_level"`
	Days  int16  `json:"days"`
	Score int
	Medal Medal `json:"medal_info"`
}
type Area struct {
	ParentName string
	ParentId   string
	Name       string
	Id         string
	Icon       string
}

type AreaLiver struct {
	UName string
	UID   int64
	Room  int
	Title string
	Cover string
}

func (client BiliClient) GetGuard(room string, liver string, delay int) []LiveUser {
	var arr = make([]LiveUser, 0)
	var page = 1
	for true {
		part := client.GetGuardByPage(room, liver, page)
		page++
		if len(part) == 0 || len(part) != 30 {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(delay))
	}
	return arr
}
func (client BiliClient) GetGuardByPage(room string, liver string, page int) []LiveUser {
	var arr = make([]LiveUser, 0)
	var url = fmt.Sprintf("https://api.live.com/xlive/app-room/v2/guardTab/topListNew?roomid=%s&page=%s&ruid=%s&page_size=30", room, strconv.Itoa(page), liver)
	res, _ := client.Resty.R().Get(url)
	var r = GuardListResponse{}
	json.Unmarshal(res.Body(), &r)
	for _, s := range r.Data.Top {
		var watcher = LiveUser{}
		watcher.Name = s.Info.User.Name
		watcher.Face = s.Info.User.Face
		watcher.Days = s.Days
		watcher.UID = s.Info.UID
		watcher.Medal.Name = s.Info.Medal.Name
		watcher.Medal.Level = s.Info.Medal.Level
		watcher.Medal.Color = s.Info.Medal.Color
		watcher.Medal.GuardLevel = s.Info.Medal.GuardLevel
		watcher.Guard = s.Info.Medal.GuardLevel
		if page == 1 {
			arr = append(arr, watcher)
		}
	}
	for _, s := range r.Data.List {
		var watcher = LiveUser{}
		watcher.Name = s.Info.User.Name
		watcher.Face = s.Info.User.Face
		watcher.Days = s.Days
		watcher.UID = s.Info.UID
		watcher.Medal.Name = s.Info.Medal.Name
		watcher.Medal.Level = s.Info.Medal.Level
		watcher.Medal.Color = s.Info.Medal.Color
		watcher.Medal.GuardLevel = s.Info.Medal.GuardLevel
		watcher.Guard = s.Info.Medal.GuardLevel
		arr = append(arr, watcher)
	}

	return arr
}
func (client BiliClient) GetFansClub(liver string, delay int, onError func(msg string)) []LiveUser {
	var page = 1
	var list = make([]LiveUser, 0)
	t := "0"
	for {
		u := fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/getFansMembersRank?ruid=%s&page=%s&page_size=30&rank_type=%s&ts=%s", liver, strconv.Itoa(page), t, strconv.FormatInt(time.Now().Unix(), 10))
		res, _ := client.Resty.R().Get(u)
		obj := FansClubResponse{}
		json.Unmarshal(res.Body(), &obj)
		if obj.Message != "0" {
			onError(obj.Message)
			return nil
		} else {
			page++
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
		for _, s := range obj.Data.Item {
			var d = LiveUser{}
			d.Score = s.Score
			d.Medal.GuardLevel = s.Medal.Type
			d.Guard = s.Medal.Type
			d.Medal.Level = s.Medal.Level
			d.UID = s.UID
			d.Name = s.UName
			d.Face = s.Face
			d.Medal.Name = s.Medal.Name
			list = append(list, d)
		}
		if len(obj.Data.Item) == 0 {
			break
		}
	}
	return list
}
func (client BiliClient) GetOnline(room string, liver string) []LiveUser {
	var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/queryContributionRank?ruid=%s&room_id=%s", liver, room)
	res, _ := client.Resty.R().Get(url)
	var o = OnlineWatcherResponse{}
	json.Unmarshal(res.Body(), &o)
	var arr = make([]LiveUser, 0)
	for _, s := range o.Data.Item {
		var watcher = LiveUser{}
		watcher.Name = s.Name
		watcher.Face = s.Face
		watcher.Days = s.Days
		watcher.UID = s.UID
		watcher.Guard = s.Guard
		watcher.Medal.Color = s.UInfo.Medal.Color
		watcher.Medal.Name = s.UInfo.Medal.Name
		watcher.Medal.Level = s.UInfo.Medal.Level
		arr = append(arr, watcher)
	}
	return arr
}
func (client BiliClient) GetLiveStream(room string) string {

	now := time.Now()

	uri, _ := url.Parse("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?qn=10000&protocol=0,1&format=0,1,2&codec=0,1,2&web_location=444.8&room_id=" + room)
	signed, _ := client.WBI.SignQuery(uri.Query(), now)
	res, _ := client.Resty.R().Get("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?" + signed.Encode())
	var s = LiveStreamResponse{}
	json.Unmarshal(res.Body(), &s)
	stream := s.Data.PlayurlInfo.Playurl.Stream
	if stream != nil {
		//Format[0]是ts格式，可以直接拿来拼接，Format[1]是fmp4，需要先把ext-x-map拼到每一个分片前面，好像还有点问题
		obj := stream[len(stream)-1].Format[0].Codec[ /*len(stream[len(stream)-1].Format[0].Codec)-1*/ 0]
		if obj.UrlInfo[0].Host+obj.BaseUrl+obj.UrlInfo[0].Extra == "" {

		}
		return obj.UrlInfo[0].Host + obj.BaseUrl + obj.UrlInfo[0].Extra
	} else {

	}

	return ""

}
func (client BiliClient) GetAreaLiveByPage(area int, page int) []AreaLiver {
	var now = time.Now()
	u, _ := url.Parse(fmt.Sprintf("https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?platform=web&parent_area_id=%d&area_id=0&sort_type=&page=%d&vajra_business_key=&web_location=444.43", area, page))
	s, _ := client.WBI.SignQuery(u.Query(), now)
	res, _ := client.Resty.R().Get("https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?" + s.Encode())
	obj := AreaLiverListResponse{}
	json.Unmarshal(res.Body(), &obj)
	var array = make([]AreaLiver, 0)
	if len(obj.Data.List) == 0 {
		fmt.Println(res.String())
	}
	for _, s2 := range obj.Data.List {
		var liver = AreaLiver{}
		liver.UID = s2.UID
		liver.Room = s2.Room
		liver.Title = s2.Title
		liver.Cover = s2.Cover
		liver.UName = s2.UName
		array = append(array, liver)
	}
	return array
}
func (client BiliClient) GetAreas() []Area {
	u := "https://api.live.bilibili.com/xlive/web-interface/v1/index/getWebAreaList?source_id=2"
	res, _ := client.Resty.R().Get(u)
	obj := AreaListResponse{}
	json.Unmarshal(res.Body(), &obj)
	var arr = make([]Area, 0)
	for _, datum := range obj.Data.Data {
		for _, s := range datum.List {
			var area = Area{}
			copier.Copy(&area, &s)
			arr = append(arr, area)
		}

	}
	return arr
}
