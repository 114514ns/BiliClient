package bili

import (
	"fmt"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/protobuf/encoding/protowire"
	"log"
	"os"
	"testing"
)

func TestDynamic(t *testing.T) {

	file, _ := os.ReadFile("cookie.txt")

	var cookie = string(file)
	cookie = ""

	//launch()
	/*
		client := NewClient(cookie, ClientOptions{})
		client.TraceLive("22749172", nil, nil)

	*/

	client := NewClient(cookie, ClientOptions{})

	client.GetDynamicsByUser(3493118494116797, "-480")

	// 加载二进制数据
	data, err := os.ReadFile("MainList")
	//data = data[5:]
	/*

		gr, err := gzip.NewReader(bytes.NewReader(data))
		all, err := io.ReadAll(gr)
		os.WriteFile("MainList", all, 0777)
		if err != nil {
			log.Fatalf("读取数据失败: %v", err)
		}


	*/

	getFawkes()
	msg := dynamic.NewMessage(protoMap["Reply.MainListReply"])
	err = msg.Unmarshal(data)
	if err != nil {
		log.Fatalf("反序列化失败: %v", err)
	}

	// 5. 打印结果（可以转为 JSON）
	jsonStr, err := msg.MarshalJSON()
	if err != nil {
		log.Fatalf("转为 JSON 失败: %v", err)
	}
	fmt.Println(string(jsonStr))
}
func isLikelyProtobuf(data []byte) bool {
	// 能否消费字段 tag
	num, typ, n := protowire.ConsumeTag(data)
	if n <= 0 || num > 10000 {
		return false
	}

	// 尝试解值
	switch typ {
	case protowire.VarintType:
		_, n := protowire.ConsumeVarint(data[n:])
		return n > 0
	case protowire.BytesType:
		_, n := protowire.ConsumeBytes(data[n:])
		return n > 0
	default:
		return false
	}
}

func PrintLiveMsg(action FrontLiveAction) {
	if action.ActionName == "msg" {
		fmt.Printf("[%s]   %s\n", action.FromName, action.Extra)
	}
}
