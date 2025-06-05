package bili

import (
	"fmt"
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

}

func PrintLiveMsg(action FrontLiveAction) {
	if action.ActionName == "msg" {
		fmt.Printf("[%s]   %s\n", action.FromName, action.Extra)
	}
}
