package bili

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	file, _ := os.ReadFile("cookie.txt")
	client := NewClient(string(file))
	//client := NewAnonymousClient()
	//fmt.Println(client.GetLiveStream("31015070"))
	fmt.Println(client.GetFace(451537183))
}
