package bili

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"strings"
)

type BiliClient struct {
	Cookie    string
	Resty     *resty.Client
	WBI       *WBI
	UserAgent string
	UID       int64
}

func NewClient(cookie string) *BiliClient {
	client := &BiliClient{Cookie: cookie, Resty: resty.New()}
	setupClient(client, cookie)
	return client
}
func NewAnonymousClient() *BiliClient {
	client := &BiliClient{Resty: resty.New()}
	setupClient(client, "")
	return client
}
func setupClient(client *BiliClient, cookie string) {
	client.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
	client.Resty.OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
		request.Header.Set("User-Agent", client.UserAgent)
		request.Header.Set("Referer", "https://www.bilibili.com/")
		//request.Header.Set("Cookie", client.Cookie)
		return nil
	})
	if cookie == "" {
		r, _ := client.Resty.R().Get("https://space.bilibili.com/208259/")
		for _, s := range r.Header().Values("Set-Cookie") {
			cookie += strings.Split(s, ";")[0] + ";"
		}
		cookie = cookie[0 : len(cookie)-1]
	}
	client.Cookie = cookie
	client.Resty.OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
		request.Header.Set("Cookie", client.Cookie)
		return nil
	})
	client.WBI = NewDefaultWbi()
	client.WBI.WithRawCookies(cookie)
	client.WBI.doInitWbi()
	client.UID = client.selfUID()
}
func (client *BiliClient) CSRF() string {
	split := strings.Split(client.Cookie, ";")
	jct := ""
	for _, s := range split {
		if strings.Contains(s, "bili_jct=") {
			jct = strings.Replace(s, "bili_jct=", "", 1)
		}
	}
	jct = jct[1:len(jct)]
	return jct
}
func (client BiliClient) selfUID() int64 {
	res, _ := client.Resty.R().SetHeader("Cookie", client.Cookie).Get("https://api.bilibili.com/x/web-interface/nav")

	var self = SelfInfo{}
	json.Unmarshal(res.Body(), &self)
	return self.Data.Mid
}
