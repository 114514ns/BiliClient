package bili

import (
	"fmt"
	json "github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"strings"
	"sync"
	"testing"
	"time"
)

var restyClient = resty.New()
var miss = 0
var hit = 0
var mutex sync.Mutex

type UserMapping struct {
	UID   int64
	UName string
	Hash  string
}

var cache = make(map[string]UserMapping)

func TestHashableLiveDanmaku(t *testing.T) {
	//直播websocket消息
	client := getClient()
	go func() {
		for {
			time.Sleep(15 * time.Second)
			fmt.Printf("Miss %d Hit %d Rate %2f\n", miss, hit, (float64(hit))/(float64(miss+hit)))
		}
	}()
	go func() {
		for {
			time.Sleep(2 * time.Second)
			for _, action := range client.GetHistory(31405244) {
				cache[action.Hash] = UserMapping{UName: action.FromName, UID: action.FromId}
			}
		}
	}()
	client.TraceLive("https://live.bilibili.com/31405244", PrintLiveMsg, HashHandler)

}

func HashHandler(hash string, name *string, id *int64) {
	type Response struct {
		List []struct {
			UID int64

			UName string
		} `json:"list"`
	}
	res, _ := restyClient.R().Get("http://127.0.0.1:8081/hash?hash=" + hash)
	v, ok := cache[hash]
	if ok {
		mutex.Lock()
		hit++
		mutex.Unlock()
		*name = v.UName
		*id = v.UID
		return
	}
	var obj Response
	json.Unmarshal([]byte(res.String()), &obj)
	if len(obj.List) < 1 {
		fmt.Println("[Hash] Not found")
		mutex.Lock()
		miss++
		mutex.Unlock()
	} else {
		mutex.Lock()
		hit++
		mutex.Unlock()
		if len(obj.List) >= 2 {
			fmt.Println("[Hash] Multi Hash Found")
			var found = false
			for _, s := range obj.List {
				if strings.HasPrefix(s.UName, (*name)[0:1]) {
					fmt.Println("[Hash] Select " + s.UName)
					*name = s.UName
					*id = s.UID
					found = true
					break
				}
			}
			if !found {
				time.Now()
			}
		} else {
			*name = obj.List[0].UName
			*id = obj.List[0].UID
		}
	}
}
