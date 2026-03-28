package bili

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/gorilla/websocket"
)

type Certificate struct {
	Uid      int64  `json:"uid"`
	RoomId   int    `json:"roomid"`
	Key      string `json:"key"`
	Protover int    `json:"protover"`
	Cookie   string `json:"buvid"`
	Type     int    `json:"type"`
	UUID     string `jsonLqueue_uuid`
}

func buildMessage(str string, opCode int) []byte {
	buffer := new(bytes.Buffer)
	totalSize := uint32(16 + len(str)) // 封包总大小
	headerLength := uint16(16)         // 头部长度
	protocolVersion := uint16(1)       // 协议版本
	operation := uint32(opCode)        // 操作码
	sequence := uint32(1)              // sequence

	binary.Write(buffer, binary.BigEndian, totalSize)
	binary.Write(buffer, binary.BigEndian, headerLength)
	binary.Write(buffer, binary.BigEndian, protocolVersion)
	binary.Write(buffer, binary.BigEndian, operation)
	binary.Write(buffer, binary.BigEndian, sequence)
	buffer.Write([]byte(str))

	return buffer.Bytes()
}

var GiftPrice = map[string]float32{}
var GiftPic = make(map[string]string)
var mu sync.RWMutex

func (client BiliClient) FillGiftPrice(room string, area int, parent int) map[string]float32 {
	res, _ := client.Resty.R().Get("https://api.live.bilibili.com/xlive/web-room/v1/giftPanel/roomGiftList?platform=pc&room_id=" + room + "&area_id=" + strconv.Itoa(area) + "&area_parent_id" + strconv.Itoa(parent))
	var gift = GiftList{}
	json.Unmarshal(res.Body(), &gift)
	for i := range gift.Data.GiftConfig.BaseConfig.List {
		var item = gift.Data.GiftConfig.BaseConfig.List[i]

		if strings.Contains(item.Name, "盲盒") {
			res, _ := client.Resty.R().Get("https://api.live.bilibili.com/xlive/general-interface/v1/blindFirstWin/getInfo?gift_id=" + strconv.Itoa(item.ID))

			var box = GiftBox{}
			json.Unmarshal(res.Body(), &box)
			for i2 := range box.Data.Gifts {
				var item0 = box.Data.Gifts[i2]
				mu.Lock()
				GiftPrice[item0.GiftName] = float32(item0.Price) / 1000.0
				GiftPic[item0.GiftName] = item0.Picture
				mu.Unlock()
			}
		} else {
			mu.Lock()
			GiftPrice[item.Name] = float32(item.Price) / 1000.0
			GiftPic[item.Name] = item.Picture
			mu.Unlock()
		}

	}
	for i := range gift.Data.GiftConfig.RoomConfig {
		var item = gift.Data.GiftConfig.RoomConfig[i]
		mu.Lock()
		GiftPrice[item.Name] = float32(item.Price) / 1000.0
		mu.Unlock()
	}
	return GiftPrice
}

func (client *BiliClient) SendMessage(msg string, room int, onResponse func(string)) {

	//u := "https://api.live.bilibili.com/msg/send?"
	//url3, _ := url.Parse(u)
	//signed, _ := client.WBI.SignQuery(url3.Query(), time.Now())
	st := `{"appId":100,"platform":5}`
	body := fmt.Sprintf("bubble=0&msg=%s&color=16777215&mode=1&room_type=0&jumpfrom=71001&reply_mid=0&reply_attr=0&replay_dmid=&statistics=%s&fontsize=25&rnd=%d&roomid=%d&csrf=%s&csrf_token=%s", msg, st, time.Now().Unix()/1000, room, client.CSRF(), client.CSRF())
	res, _ := client.Resty.R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetBody(body).Post("https://api.live.bilibili.com/msg/send?" /*+ signed.Encode()*/)
	fmt.Println(res.String())
	if onResponse != nil {
		onResponse(res.String())
	}
}

func (client *BiliClient) GetHistory(room int) []FrontLiveAction {

	var array []FrontLiveAction
	var u = "https://api.live.bilibili.com/xlive/web-room/v1/dM/gethistory?roomid=" + strconv.Itoa(room)
	res, _ := client.Resty.R().Get(u)
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	var process []interface{}
	process = append(process, getArray(obj, "data.admin")...)
	process = append(process, getArray(obj, "data.room")...)
	for _, i := range process {
		var action FrontLiveAction
		action.ActionName = "msg"
		action.FromId = getInt64(i, "uid")
		action.FromName = getString(i, "nickname")
		action.Extra = getString(i, "text")
		action.Time, _ = time.Parse(time.DateTime, getString(i, "timeline"))
		action.HonorLevel = int8(getInt(i, "wealth_level"))
		action.GuardLevel = int8(getInt(i, "guard_level"))
		action.Face = getString(i, "user.base.face")
		action.Hash = fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(strconv.FormatInt(action.FromId, 10))))
		medals := getObject(i, "user.medal", "object").m
		if medals != nil {
			action.MedalName = getString(medals, "name")
			action.MedalColor = getString(medals, "v2_medal_color_start")
			action.MedalLevel = int8(getInt(medals, "level"))
		}
		array = append(array, action)
	}
	return array

}

func (client *BiliClient) GetLiveOnlineRank(room int, uid0 ...int64) []User {
	var dst []User
	var uid int64 = 0
	if len(uid0) > 0 {
		uid = uid0[0]
	} else {
		res, _ := client.Resty.R().Get("https://api.live.bilibili.com/room/v1/Room/get_info?room_id=%d" + toString(int64(room)))
		var obj map[string]interface{}
		json.Unmarshal(res.Body(), &obj)
		uid = getInt64(res, "data.uid")
	}
	var url = fmt.Sprintf("https://api.live.bilibili.com/xlive/general-interface/v1/rank/queryContributionRank?ruid=%d&room_id=%d", uid, room)
	res, _ := client.Resty.R().Get(url)
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	for _, i := range getArray(obj, "data.item") {
		dst = append(dst, User{
			Face:  getString(i, "face"),
			UID:   getInt64(i, "uid"),
			UName: getString(i, "name"),
		})
	}

	return dst
}
func randomQueueString() string {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Float64()
	str := strconv.FormatFloat(randNum, 'f', -1, 64)
	numStr := strconv.FormatInt(int64(randNum*1e15), 36)
	_ = str
	if len(numStr) < 8 {
		return numStr
	}
	return numStr[len(numStr)-8:]
}
func (client BiliClient) TraceLive(room string, onMessage func(action FrontLiveAction), hashHandler func(hash string, name *string, id *int64)) {
	if strings.Contains(room, "https") {
		room = strings.Split(room, "/")[3:][0]
	}
	url0 := "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + room + "&is_anchor=true&isGaiaAvoided=true"
	query, _ := url.Parse(url0)
	signed, _ := client.WBI.SignQuery(query.Query(), time.Now())
	res, _ := client.Resty.R().SetHeader("Cookie", client.Cookie).Get("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?" + signed.Encode())
	var liveInfo = LiveInfo{}
	json.Unmarshal(res.Body(), &liveInfo)
	ticker := time.NewTicker(45 * time.Second)
	var dialer = &websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}
	var host = liveInfo.Data.HostList[0].Host
	host = "ali-sh-live-comet-08.chat.bilibili.com"
	u := url.URL{Scheme: "wss", Host: host + ":2245", Path: "/sub"}
	var header = http.Header{}
	header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	c, _, err := dialer.Dial(u.String(), header)
	go func() {
		var cer = Certificate{}
		//cer.Uid = client.selfUID()
		id, _ := strconv.Atoi(room)
		cer.RoomId = id
		cer.Type = 2
		cer.Key = liveInfo.Data.Token
		cer.Cookie = strings.Replace(client.Cookie, "buvid3=", "", 1)
		cer.Protover = 3
		cer.UUID = randomQueueString()
		cerJson, _ := json.Marshal(&cer)

		err := c.WriteMessage(websocket.TextMessage, buildMessage(string(cerJson), 7))
		if err != nil {
			return
		}
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				time.Sleep(2 * time.Second)
				c, _, _ = dialer.Dial(u.String(), header)
				if c != nil {
					c.WriteMessage(websocket.TextMessage, buildMessage(string(cerJson), 7))

				}
				time.Sleep(5 * time.Second)
				continue
			}
			reader := io.NewSectionReader(bytes.NewReader(message), 16, int64(len(message)-16))
			brotliReader := brotli.NewReader(reader)
			var decompressedData bytes.Buffer
			var msg = ""
			_, err0 := io.Copy(&decompressedData, brotliReader)
			if err0 != nil {
				msg = string(message)
			} else {
				msg = string(decompressedData.Bytes())
			}
			buffer := bytes.NewReader([]byte(msg))

			for {
				if buffer.Len() < 16 {
					break
				}

				var totalSize uint32
				var headerLength uint16
				var protocolVersion uint16
				var operation uint32
				var sequence uint32

				binary.Read(buffer, binary.BigEndian, &totalSize)
				binary.Read(buffer, binary.BigEndian, &headerLength)
				binary.Read(buffer, binary.BigEndian, &protocolVersion)
				binary.Read(buffer, binary.BigEndian, &operation)
				binary.Read(buffer, binary.BigEndian, &sequence)
				if buffer.Len() < int(totalSize-16) {
					break
				}
				msgData := make([]byte, totalSize-16)
				buffer.Read(msgData)

				var obj = string(msgData)
				var action = LiveAction{}
				action.LiveRoom = room
				action.GiftPrice = 0
				action.GiftAmount = 0
				var text = LiveText{}
				json.Unmarshal(msgData, &text)
				var front = FrontLiveAction{}
				parsed := true
				front.Emoji = make(map[string]string)
				front.Time = time.Now()
				if strings.Contains(obj, "DANMU_MSG") && !strings.Contains(obj, "RECALL_DANMU_MSG") { // 弹幕
					action.ActionName = "msg"
					action.Hash = text.Info[0].([]interface{})[7].(string)
					action.FromName = text.Info[2].([]interface{})[1].(string)
					e1, ok := text.Info[0].([]interface{})[13].(map[string]interface{})
					if ok {
						e2, ok := e1["emoticon_unique"].(string)
						if ok {
							front.Emoji[strings.Replace(e2, "upower_", "", 1)] = e1["url"].(string)
						}
					}

					var o interface{}
					json.Unmarshal([]byte(text.Info[0].([]interface{})[15].(map[string]interface{})["extra"].(string)), &o)
					e, ok := o.(map[string]interface{})["emots"]
					if e != nil {
						emots := e.(map[string]interface{})
						if len(emots) != 0 {
							for s, i := range emots {
								front.Emoji[s] = i.(map[string]interface{})["url"].(string)
							}
						}
					}

					action.FromId = int64(text.Info[2].([]interface{})[0].(float64))
					action.HonorLevel = int8(text.Info[16].([]interface{})[0].(float64))
					action.Extra = text.Info[1].(string)
					value, ok := text.Info[0].([]interface{})[15].(map[string]interface{})
					if strings.Contains(action.Extra, "[") {
						time.Now()
					}
					if ok {
						user, exists := value["user"].(map[string]interface{})
						if exists {
							base, exists := user["base"].(map[string]interface{})
							if exists {
								face, exists := base["face"]
								if exists {
									front.Face = face.(string)
								}
							}
							medal, exists := user["medal"].(map[string]interface{})
							if exists {
								name, exists := medal["name"]
								if exists {
									action.MedalName = name.(string)
								}
								level, exists := medal["level"]
								if exists {
									action.MedalLevel = int8(level.(float64))
								}
								guardLevel, exists := medal["guard_level"]
								if exists {
									action.GuardLevel = int8(guardLevel.(float64))
								}
								color, exists := medal["v2_medal_color_start"]
								if exists {
									front.MedalColor = color.(string)
								}

							}
						}
					}

				} else if strings.Contains(obj, "SEND_GIFT") { //送礼物
					var info = GiftInfo{}
					json.Unmarshal(msgData, &info)
					action.ActionName = "gift"
					action.FromName = info.Data.Uname
					action.GiftName = info.Data.GiftName
					action.MedalLevel = int8(info.Data.Medal.Level)
					action.MedalName = info.Data.Medal.Name
					action.HonorLevel = info.Data.HonorLevel
					action.FromId = info.Data.SenderUinfo.UID
					front.MedalColor = fmt.Sprintf("#%06X", info.Data.Medal.Color)
					price := float64(GiftPrice[info.Data.GiftName]) * float64(info.Data.Num)
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", price), 64)
					action.GiftPrice = float32(result)
					if price == 0 {
						price = float64(info.Data.Price) / 1000 * float64(info.Data.Num)
						action.GiftPrice = float32(price)
					}
					action.GiftAmount = int16(info.Data.Num)
					if info.Data.Parent.GiftName != "" {
						action.Extra = info.Data.Parent.GiftName + "," + strconv.Itoa(info.Data.Parent.Price/1000)
					}
					front.Face = info.Data.Face
					front.GiftPicture = GiftPic[info.Data.GiftName]
				} else if text.Cmd == "INTERACT_WORD" { //进入直播间

					var enter = EnterLive{}
					json.Unmarshal(msgData, &enter)
					action.FromId = enter.Data.UID
					action.FromName = enter.Data.Uname
					action.ActionName = "enter"
					front.LiveAction = action
					//log.Printf("[%s] 进入直播间", action.FromName)

				} else if strings.Contains(obj, "PREPARING") {
					c.Close()
				} else if text.Cmd == "LIVE" {

				} else if strings.Contains(obj, "SUPER_CHAT_MESSAGE") { //SC
					var sc = SuperChatInfo{}
					json.Unmarshal(msgData, &sc)

					action.ActionName = "sc"
					action.FromName = sc.Data.UserInfo.Uname
					action.FromId = sc.Data.Uid
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sc.Data.Price), 64)
					action.GiftPrice = float32(result)

					action.GiftAmount = 1
					action.Extra = sc.Data.Message
				} else if strings.Contains(obj, "GUARD_BUY") { //上舰
					var guard = GuardInfo{}
					json.Unmarshal(msgData, &guard)
					action.FromId = guard.Data.Uid
					action.ActionName = "guard"
					action.FromName = guard.Data.Username
					action.GiftName = guard.Data.GiftName
					switch action.GiftName {
					case "舰长":
						action.GiftPrice = float32(138 * guard.Data.Num)
					case "提督":
						action.GiftPrice = float32(1998 * guard.Data.Num)
					case "总督":
						action.GiftPrice = float32(19998 * guard.Data.Num)
					}

				} else if text.Cmd == "WATCHED_CHANGE" {
					var obj = Watched{}
					json.Unmarshal(msgData, &obj)
				} else if text.Cmd == "ONLINE_RANK_COUNT" {
					//fmt.Println(text)
				} else {
					//log.Println(string(msgData))
				}
				front.LiveAction = action
				if parsed {
					if onMessage != nil {
						if hashHandler != nil && front.ActionName == "msg" {
							go func() {
								hashHandler(front.Hash, &front.FromName, &front.FromId)
								onMessage(front)
							}()
						} else {
							if front.ActionName != "" {
								onMessage(front)
							}
						}

					}
				}
				if buffer.Len() < 16 {
					break
				}

			}
			if !strings.Contains(msg, "[object") {

				//log.Println(msg)
			}

		}
	}()
	for {
		select {
		case <-ticker.C:
			err = c.WriteMessage(websocket.TextMessage, buildMessage("[object Object]", 2))
			//lives[roomId].LastActive = time.Now().Unix() + 3600*8
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func (client *BiliClient) TraceStream(room int, dst0 string, onlyAudio ...bool) {
	if len(onlyAudio) == 0 {
		onlyAudio = append(onlyAudio, false)
	}
	var stream = client.GetLiveStream(room, "fmp4", onlyAudio[0])
	var ticker = time.NewTicker(time.Second * 2)
	var dst, _ = os.Create(dst0)
	writer := bufio.NewWriter(dst)
	var m = make(map[string]bool)
	//var m = make(map[string]bool)
	var u, _ = url.Parse(stream)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			str, _ := client.Resty.R().Get(stream)
			for _, s := range strings.Split(str.String(), "\n") {
				if !strings.HasPrefix(s, "#") {
					_, ok := m[s]
					if !ok {
						path := u.Path
						split := strings.Split(path, "/")
						var d = ""
						for i, s2 := range split {
							if i != len(split)-1 {
								d += s2 + "/"
							}
						}
						r, _ := client.Resty.R().Get("https://" + u.Host + d + s)
						writer.Write(r.Body())
						writer.Flush()
						m[s] = true
					}
				}
			}
		}
	}
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
