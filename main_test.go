package bili

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

func getClient() *BiliClient {
	file, _ := os.ReadFile("cookie.txt")

	var cookie = string(file)
	//cookie = ""
	client := NewClient(cookie, ClientOptions{
		//HttpProxy: "[2401:b60:2a:c186:2871:8c58:763e:45ff]:8080",
		ProxyUser: "fg27msTTyo",
		ProxyPass: "PZ8u9Pr2oz",
		//NoCookie:  true,
	})

	return client

}

func TestDynamic(t *testing.T) {
	//获取用户动态
	client := getClient()
	client.Options.NoCookie = true

	var array []Dynamic

	/*

		dynamics, offset := client.GetDynamicsByUser(3493118494116797)
		for _, dynamic := range dynamics {
			array = append(array, dynamic)
		}
		dynamics, _ = client.GetDynamicsByUser(3493118494116797, offset)
		for _, dynamic := range dynamics {
			array = append(array, dynamic)
		}


	*/

	offset := ""
	for {
		dynamics, offset0 := client.GetDynamicsByUser(1352996769, offset)

		if offset0 != "-1" {
			offset = offset0
			for _, dynamic := range dynamics {
				array = append(array, dynamic)
			}
		}
		if "" == offset0 {
			break
		}
		fmt.Println(len(array))
		time.Sleep(1 * time.Second)
	}
	PrintJSON(array)

}
func TestLiveHistory(t *testing.T) {
	client := getClient()
	client.GetHistory(23174842)
}
func TestLiveDanmaku(t *testing.T) {
	//直播websocket消息
	client := getClient()
	go func() {
		for {
			time.Sleep(15 * time.Second)
			fmt.Printf("Miss %d Hit %d Rate %2f\n", miss, hit, (float64(hit))/(float64(miss+hit)))
		}
	}()
	client.TraceLive("https://live.bilibili.com/21402309", PrintLiveMsg, HashHandler)

}

func TestVideoDetail(t *testing.T) {
	//视频详细信息
	var client = getClient()
	var video = client.GetVideo("BV1qb411i73")
	PrintJSON(video)
}

func TestUserArchive(t *testing.T) {
	//用户稿件列表
	var client = getClient()
	var videos = client.GetVideoByUser(504140200, 1, false)
	PrintJSON(videos)
	time.Sleep(1 * time.Second)
	videos = client.GetVideoByUser(504140200, 2, false)
	PrintJSON(videos)
	time.Sleep(1 * time.Second)
	videos = client.GetVideoByUser(504140200, 3, false)
	PrintJSON(videos)
}

func TestVideoComment(t *testing.T) {
	//视频评论区
	var client = getClient()
	var video = client.GetVideoByUser(504140200, 1, true)[0]
	comments, _ := video.getComments("", client)

	PrintJSON(comments)
}

func TestDynamicComment(t *testing.T) {
	//获取动态评论内容
	var client = getClient()
	dyn, _ := client.GetDynamicsByUser(232392815)
	var dst = []Comment{}
	var off = ""
	var count = 0
	for {
		var array, o = getClient().GetComment(dyn[0].CommentID, off, dyn[0].CommentType, 3)
		off = o
		time.Sleep(500 * time.Millisecond)
		for _, comment := range array {
			var tmp = Comment{}
			copier.Copy(&tmp, comment)
			dst = append(dst, tmp)
		}

		count = count + len(array)
		fmt.Println(count)
		if len(array) == 0 || off == "" {
			break
		}
	}
	//PrintJSON(dst)
	if dyn[0].Comments == len(dst) {
		fmt.Println("Count Match!")
	} else {
		fmt.Println("Count Not Match!")
	}
}

func TestGetLocation(t *testing.T) {
	var client = getClient()
	PrintJSON(client.GetLocation())          //获取当前ip信息
	PrintJSON(client.GetLocation("8.8.8.8")) //获取指定ip信息
}
func TestGetArticle(t *testing.T) {
	//那两个回调不用管，历史遗留问题
	PrintJSON(getClient().GetArticle(27148899, nil, nil))
}
func TestGetAreas(t *testing.T) {
	//获取所有直播分区
	PrintJSON(getClient().GetAreas())
}

func TestGetFansClub(t *testing.T) {
	//主播粉丝团
	list := getClient().GetFansClub("3493080636328319", 300, nil)
	PrintJSON(list)
}
func TestGetOnline(t *testing.T) {
	//直播间在线榜单
	var list = getClient().GetOnline("26854650", "3493118494116797")
	PrintJSON(list)
}
func TestGetGuard(t *testing.T) {
	//主播大航海
	var list = getClient().GetGuard("26854650", "3493118494116797", 300)
	PrintJSON(list)
}

func TestGetFollowing(t *testing.T) {
	//用户关注列表
	var list = getClient().GetFollowing("451537183", 300)
	PrintJSON(list)
}

func TestLiveStream(t *testing.T) {
	//直播流
	var client = getClient()
	//var stream = client.GetLiveStream(strconv.Itoa(client.GetAreaLiveByPage(9, 1)[0].Room))
	var stream = client.GetLiveStream(30931147)
	fmt.Println(stream)
}
func TestTraceLiveStream(t *testing.T) {
	var client = getClient()
	client.TraceStream(21399314, "dst.mp3", false)
}
func TestAreaLivers(t *testing.T) {
	//获取分区内开播的直播间
	var total = 0
	var page = 1
	for {
		var list = getClient().GetAreaLiveByPage(9, page)
		for _, liver := range list {
			fmt.Println("" +
				"\"" + strconv.Itoa(liver.Room) + "\",")
		}
		page++
		total += len(list)
		if len(list) == 0 {
			break
		}
		time.Sleep(time.Second * 2)
	}
	//PrintJSON(total)
}

func TestFansMedal(t *testing.T) {
	//查询用户粉丝牌，需要登录
	PrintJSON(getClient().GetFansMedal("3461575518193817"))
}

func TestSearchVideo(t *testing.T) {
	var opinion = SearchOption{
		Keyword:   "虚拟主播",
		BeginTime: time.Date(2025, 2, 5, 0, 0, 0, 0, time.Local),
		EndTime:   time.Date(2025, 2, 6, 0, 0, 0, 0, time.Local),
	}
	PrintJSON(getClient().SearchVideo(opinion))
}

func TestVideoStream(t *testing.T) {
	//获取视频流
	array := getClient().GetVideoStream("BV1TMYqzEEa1", 31728668205)
	//audio := array[0]
	PrintJSON(array)
}

func TestDownloadVideo(t *testing.T) {
	client := getClient()
	array := client.GetVideoStream("BV1eB7QzFEHB", 30291396850)
	PrintJSON(array[2])
	client.DownloadVideo(array[2], "/mnt/share/Stream")

}

func TestSign(t *testing.T) {
	var u = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=544853"
	var u0, _ = url.Parse(u)
	var client = getClient()
	query, _ := client.WBI.SignQuery(u0.Query(), time.Unix(1752042465, 0))
	PrintJSON(query)
}
func PrintLiveMsg(action FrontLiveAction) {
	var medalTag = ""
	if action.MedalName != "" {
		medalTag = fmt.Sprintf("[%s]", action.MedalName)
	}
	var levelTag = ""
	if action.MedalLevel != 0 {
		levelTag = fmt.Sprintf("[LV%d]", action.MedalLevel)
	}
	if action.ActionName == "msg" {
		fmt.Printf("%s%s[%s]   %s\n", medalTag, levelTag, action.FromName, action.Extra)
	}
	if action.ActionName == "gift" {
		var giftName = action.Extra
		if giftName == "" {
			giftName = action.GiftName
		}
		if action.Extra == "" {
			fmt.Printf("[%s]  投喂了%d个 %s   %f元\n", action.FromName, action.GiftAmount, giftName, action.GiftPrice)
		} else {
			//Extra里第一个是盲盒名字，第二个是盲盒价格，逗号分隔
			//GiftName和GiftPrice是爆出来的礼物的信息

			fmt.Printf("%s  %ss  [%s]  打开了%d个 %s 爆出%s   %f元\n", medalTag, levelTag, action.FromName, action.GiftAmount, giftName, action.GiftName, action.GiftPrice)
		}

	}
	if action.ActionName == "enter" {
		//fmt.Printf("[%s] 进入直播间\n", action.FromName)
	}
}

func PrintJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "", "    ")
	fmt.Println(string(bytes))
}

func ffplay(stream string) {
	cmd := exec.Command("ffplay", stream)
	out, _ := cmd.CombinedOutput()
	fmt.Printf("%s\n", string(out))
}
