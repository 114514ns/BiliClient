package bili

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"net/url"
	"strings"
)

type ClientOptions struct {
	HttpProxy       string
	ProxyUser       string
	ProxyPass       string
	RandomUserAgent bool
}

type BiliClient struct {
	Cookie    string
	Resty     *resty.Client
	WBI       *WBI
	UserAgent string
	UID       int64
	Options   ClientOptions
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.3",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.6",
	"Mozilla/5.0 (Linux; Android 13; SM-S908U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone14,3; U; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/602.1.50 (KHTML, like Gecko) Version/10.0 Mobile/19A346 Safari/602.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:138.0) Gecko/20100101 Firefox/138.0",
	"Mozilla/5.0 (Windows 7 Enterprise; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6099.71 Safari/537.36",
	"Mozilla/5.0 (Windows Server 2012 R2 Standard; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5975.80 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.5672 Safari/537.36",
}

func NewClient(cookie string, options ClientOptions) *BiliClient {
	client := &BiliClient{Cookie: cookie, Resty: resty.New(), Options: options}
	setupClient(client, cookie)
	return client
}
func NewAnonymousClient(options ClientOptions) *BiliClient {
	client := &BiliClient{Resty: resty.New(), Options: options}
	setupClient(client, "")
	return client
}
func setupClient(client *BiliClient, cookie string) {
	if client.Options.HttpProxy != "" {
		proxyURL, _ := url.Parse(fmt.Sprintf("http://%s:%s@%s", client.Options.ProxyUser, client.Options.ProxyPass, client.Options.HttpProxy))
		client.Resty.SetProxy(proxyURL.String())
	}
	client.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
	client.Resty.OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
		if client.Options.RandomUserAgent {
			request.Header.Set("User-Agent", userAgents[rand.Uint32()%uint32(len(userAgents))])
		} else {
			request.Header.Set("User-Agent", client.UserAgent)
		}

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
