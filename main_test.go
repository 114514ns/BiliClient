package bili

import (
	"testing"
)

func TestDynamic(t *testing.T) {
	//file, _ := os.ReadFile("cookie.txt")
	//client := NewClient(string(file))

	client := NewAnonymousClient(ClientOptions{
		HttpProxy: "*:8080",
		ProxyUser: "fg27msTTyo",
		ProxyPass: "PZ8u9Pr2oz",
	})

	//client.GetArticle(41359318)
	client.GetLocation("38.49.46.132")

}
