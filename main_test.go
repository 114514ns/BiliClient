package bili

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/samber/lo"
)

func getClient() *BiliClient {
	file, _ := os.ReadFile("cookie.txt")

	var cookie = string(file)

	file, _ = os.ReadFile("access_key.txt")
	//cookie = ""
	client := NewClient(cookie, ClientOptions{

		ProxyUser: "fg27msTTyo",
		ProxyPass: "PZ8u9Pr2oz",
		AccessKey: string(file),
		NoCookie:  false,
	})

	log.Printf("Login UID : %d\n", client.UID)

	return client

}

func TestDynamic(t *testing.T) {
	//获取用户动态
	client := getClient()
	//client.Options.NoCookie = true

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
		dynamics, offset0 := client.GetDynamicsByUser(82044006, false, offset)

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
		time.Sleep(350 * time.Microsecond)
	}
	PrintJSON(len(array))

}
func TestGetDynamicByUIDRPC(t *testing.T) {
	var client = getClient()
	for {
		client.GetDynamicByUIDRPC(672328094, 1)
		time.Sleep(500 * time.Microsecond)
	}
}

func TestDynamicReaction(t *testing.T) {
	var client = getClient()
	var page = 1
	var results []ActionItem
	print(client.GetReactionList(1121658321759109129, 103))
	for {
		te := client.GetReactionList(1121658321759109129, page)
		page++
		results = append(results, te...)
		if len(te) == 0 {
			break
		}
		time.Sleep(350 * time.Microsecond)
	}

}
func TestDynamicLikeListRPC(t *testing.T) {
	var client = getClient()
	var page = 1
	var results []ActionItem
	var offset int64 = 0
	for {
		array, more := client.GetLikeListRPC(1101369725119627286, 8, 970285943, page, offset)
		page++
		offset = array[len(array)-1].UID
		results = append(results, array...)
		if !more {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}

}
func TestLiveHistory(t *testing.T) {
	client := getClient()
	client.GetHistory(23174842)
}
func TestLiveDanmaku(t *testing.T) {
	//直播websocket消息
	client := getClient()
	//client.TraceLive("https://live.bilibili.com/23770996", PrintLiveMsg, nil)
	client.TraceLive("https://live.bilibili.com/1883943131", PrintLiveMsg, nil)

}

func TestGetRoomInfo(t *testing.T) {
	PrintJSON(getClient().GetRoomInfo(struct {
		UID  int64
		Room int
	}{UID: 0, Room: 23174842}))
}

func TestVideoDetail(t *testing.T) {
	//视频详细信息
	var client = getClient()
	var video = client.GetVideo("BV1wKrTBoEDQ")
	for i, r := range video[0].ParentTitle {
		fmt.Printf("index=%d, rune=%c, unicode=%U, int=%d\n", i, r, r, r)
	}
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
	//var video = client.GetVideoByUser(504140200, 1, true)[0]
	var video = client.GetVideo("BV1gzzYBVE7e")[0]
	var off = ""
	for {
		comments, off0 := video.getComments(off, client)
		PrintJSON(comments)
		if off0 != "" {
			off = off0
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

}

func TestGetCommentRPC(t *testing.T) {
	var client = getClient()
	var off = ""
	var dst []Reply
	for {
		arr, off0 := client.GetCommentRPC(116605392521146, off, 1, REPLY_SORT_HOT)
		dst = append(dst, arr...)
		if off == off0 || off0 == "" {
			break
		} else {
			off = off0
		}
		time.Sleep(300 * time.Millisecond)
	}
	marshal, _ := json.Marshal(dst)
	os.WriteFile("reply.json", marshal, 0666)

}

func TestDynamicComment(t *testing.T) {
	//获取动态评论内容
	var client = getClient()
	dyn, _ := client.GetDynamicsByUser(39922227, true, "")
	var dst = []Reply{}
	var off = ""
	var count = 0
	for {
		var array, o = client.GetCommentRPC(dyn[16].CommentID, off, dyn[16].CommentType, REPLY_SORT_TIME)
		off = o
		time.Sleep(500 * time.Millisecond)
		for _, comment := range array {
			var tmp = Reply{}
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
	var sum = len(dst)
	for _, i := range dst {
		sum += len(i.Reply)
	}
	if dyn[0].Comments == sum {
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
	PrintJSON(getClient().GetArticle(43750847, nil, nil))
}

func TestGetUserArticle(t *testing.T) {
	var client = getClient()
	article := client.GetUserArticle(504140200, "")
	for i := range article {
		article[i] = client.GetArticle(article[i].DynamicID, nil, nil)
		time.Sleep(2 * time.Millisecond * 500)
	}
	PrintJSON(article)
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
	var list = getClient().GetGuard("22637261", "672328094", 100, 3)
	list = append(list, getClient().GetGuard("22637261", "672328094", 100, 4)...)
	list = append(list, getClient().GetGuard("22637261", "672328094", 100, 5)...)
	list = lo.UniqBy(list, func(item LiveUser) int64 {
		return item.UID
	})
	PrintJSON(list)

}

func TestGetFollowing(t *testing.T) {
	//用户关注列表
	var list = getClient().GetFollowing("2", 300)
	PrintJSON(list)
}

func TestLiveStream(t *testing.T) {
	//直播流
	var client = getClient()
	//var stream = client.GetLiveStream(strconv.Itoa(client.GetAreaLiveByPage(9, 1)[0].Room))
	var stream = client.GetLiveStream(30858592, "ts", false)
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
			fmt.Println(liver.UName)
		}
		page++
		total += len(list)
		if len(list) == 0 {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	fmt.Println(total)
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
	array := getClient().GetVideoStream("BV1eySZBGEPV", 0)
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

func TestGetUserLoaction(t *testing.T) {
	PrintJSON(getClient().GetUserLocation(1879130196))

}

func TestGetUser(t *testing.T) {
	user, _ := getClient().GetUser(690608691)
	PrintJSON(user)
}

func TestGetUserFav(t *testing.T) {
	PrintJSON(getClient().GetFavList("3493078891497932", 1))
}

func TestGetUserList(t *testing.T) {
	//PrintJSON(getClient().GetUserCollection(82044006, 1))
}

func TestCharge(t *testing.T) {
	PrintJSON(getClient().GetCharge(342476492))
}

func TestGetCollectionItems(t *testing.T) {
	/*
		var client = getClient()

			var list = client.GetUserCollection(3493078891497932, 1)
			var item = list[len(list)-1]
			PrintJSON(client.GetCollectionItems(item.ID, item.Type, 1, item.UID))



	*/
	//PrintJSON(client.GetCollectionItems(583984, "", 1, 15567349))
}

func TestSearchUser(t *testing.T) {
	var client = getClient()
	PrintJSON(client.SearchUser("塔菲"))
}

func TestGetDynamicByOID(t *testing.T) {
	PrintJSON(getClient().GetDynamicDetail(1204380156358033412))
}

func TestBatchGetFace(t *testing.T) {
	var client = getClient()
	var ids []int64
	ids = append(ids, 527868)
	PrintJSON(client.BatchGetFace(ids))
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
