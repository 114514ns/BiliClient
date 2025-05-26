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
	client := NewClient(cookie, ClientOptions{})

	/*
		client := NewAnonymousClient(ClientOptions{
			HttpProxy:       "156.226.170.127:8080",
			ProxyUser:       "fg27msTTyo",
			ProxyPass:       "PZ8u9Pr2oz",
			RandomUserAgent: true,
		})

	*/

}
