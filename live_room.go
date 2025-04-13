package bili

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Certificate struct {
	Uid      int64  `json:"uid"`
	RoomId   int    `json:"roomid"`
	Key      string `json:"key"`
	Protover int    `json:"protover"`
	Cookie   string `json:"buvid"`
	Type     int    `json:"type"`
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

func (client BiliClient) FillGiftPrice(room string, area int, parent int) {
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
}

func (client *BiliClient) SendMessage(msg string, room int) {

	u := "https://api.live.bilibili.com/msg/send?"
	url3, _ := url.Parse(u)
	signed, _ := client.WBI.SignQuery(url3.Query(), time.Now())
	st := `{"appId":100,"platform":5}`
	body := fmt.Sprintf("bubble=0&msg=%s&color=16777215&mode=1&room_type=0&jumpfrom=71001&reply_mid=0&reply_attr=0&replay_dmid=&statistics=%s&font_size=25&rnd=%d&roomid=%d&csrf=%s&csrf_token=%s", msg, st, time.Now().Unix()/1000, room, client.CSRF(), client.CSRF())
	res, _ := client.Resty.R().SetHeader("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundarytdNYxtgGKJBK9qSv").SetBody(body).Post("https://api.live.bilibili.com/msg/send?" + signed.Encode())
	fmt.Println(res.String())
}

func (client BiliClient) TraceLive(room string, onMessage func(action FrontLiveAction), onChange func(state string)) {
	url0 := "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?type=0&id=" + room
	res, _ := client.Resty.R().Get(url0)
	var liveInfo = LiveInfo{}
	json.Unmarshal(res.Body(), &liveInfo)
	ticker := time.NewTicker(45 * time.Second)
	var dialer = &websocket.Dialer{
		Proxy:            nil,
		HandshakeTimeout: 45 * time.Second,
	}
	u := url.URL{Scheme: "wss", Host: liveInfo.Data.HostList[0].Host + ":2245", Path: "/sub"}
	c, _, err := dialer.Dial(u.String(), nil)
	go func() {
		var cer = Certificate{}
		cer.Uid = client.UID
		id, _ := strconv.Atoi(room)
		cer.RoomId = id
		cer.Type = 2
		cer.Key = liveInfo.Data.Token
		cer.Cookie = strings.Replace(client.Cookie, "buvid3=", "", 1)
		cer.Protover = 3
		cerJson, _ := json.Marshal(&cer)

		err := c.WriteMessage(websocket.TextMessage, buildMessage(string(cerJson), 7))
		if err != nil {
			return
		}
		for {
			_, message, err := c.ReadMessage()
			if err != nil {

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
				if strings.Contains(obj, "DANMU_MSG") && !strings.Contains(obj, "RECALL_DANMU_MSG") { // 弹幕
					action.ActionName = "msg"
					action.FromName = text.Info[2].([]interface{})[1].(string)
					action.FromId = strconv.Itoa(int(text.Info[2].([]interface{})[0].(float64)))
					action.Extra = text.Info[1].(string)
					value, ok := text.Info[0].([]interface{})[15].(map[string]interface{})
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
					action.FromId = strconv.Itoa(info.Data.SenderUinfo.UID)
					front.MedalColor = fmt.Sprintf("#%06X", info.Data.Medal.Color)
					price := float64(GiftPrice[info.Data.GiftName]) * float64(info.Data.Num)
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", price), 64)
					action.GiftPrice = float32(result)
					action.GiftAmount = int16(info.Data.Num)
					if info.Data.Parent.GiftName != "" {
						action.Extra = info.Data.Parent.GiftName + "," + strconv.Itoa(info.Data.Parent.Price/1000)
					}
					front.Face = info.Data.Face
					front.GiftPicture = GiftPic[info.Data.GiftName]
				} else if strings.Contains(obj, "INTERACT_WORD") { //进入直播间

					var enter = EnterLive{}
					json.Unmarshal(msgData, &enter)
					action.FromId = strconv.Itoa(enter.Data.UID)
					action.FromName = enter.Data.Uname
					action.ActionName = "enter"
					front.LiveAction = action
					//db.Create(&action)

				} else if strings.Contains(obj, "PREPARING") {

				} else if text.Cmd == "LIVE" {

				} else if strings.Contains(obj, "SUPER_CHAT_MESSAGE") { //SC
					var sc = SuperChatInfo{}
					json.Unmarshal(msgData, &sc)

					action.ActionName = "sc"
					action.FromName = sc.Data.UserInfo.Uname
					action.FromId = strconv.Itoa(sc.Data.Uid)
					result, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sc.Data.Price), 64)
					action.GiftPrice = float32(result)

					action.GiftAmount = 1
					action.Extra = sc.Data.Message
				} else if strings.Contains(obj, "GUARD_BUY") { //上舰
					var guard = GuardInfo{}
					json.Unmarshal(msgData, &guard)
					action.FromId = strconv.Itoa(guard.Data.Uid)
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
				} else {
					parsed = false
				}
				front.LiveAction = action
				if parsed {
					onMessage(front)
				}
				if buffer.Len() < 16 {
					break
				}

			}
			if !strings.Contains(msg, "[object") {

				//log.Printf("Received: %s", substr(msg, 16, len(msg)))
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
