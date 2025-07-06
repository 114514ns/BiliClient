package bili

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	json "github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strings"
)

type ClientOptions struct {
	HttpProxy       string
	ProxyUser       string
	ProxyPass       string
	RandomUserAgent bool
	ResetConnection bool
	NoCookie        bool
}

type BiliClient struct {
	Cookie    string
	Resty     *resty.Client
	WBI       *WBI
	UserAgent string
	UID       int64
	Options   ClientOptions
	Address   string
}
type protoType0 struct {
	Reply_MainListReply string
	Reply_MainListReq   string
	Metadata_FawkesReq  string
	Metadata            string
	Device              string
}

var ProtoType protoType0
var protoMap = make(map[string]*desc.MessageDescriptor)

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
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.132 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 13; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.136 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; U; Android 13; en-US; SM-G991U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.136 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/118.0.2088.76 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.132 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 13; OnePlus 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.132 Mobile Safari/537.36",
	"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/537.36 (KHTML, like Gecko) Firefox/118.0 Safari/537.36",
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

	_, err := os.Open("bilibili")
	//client.Resty.SetRetryCount(15)
	/*
		tr, err := srt.NewSpoofedRoundTripper(
			// Reference for more: https://bogdanfinn.gitbook.io/open-source-oasis/tls-client/client-options
			tlsclient.WithRandomTLSExtensionOrder(), // needed for Chrome 107+
			tlsclient.WithClientProfile(profiles.Firefox_135),
		)
		client.Resty.SetTransport(tr)

	*/
	//client.Resty.SetRetryWaitTime(time.Microsecond * 1500)
	if err == nil {
		parser := protoparse.Parser{}
		{

			ProtoType.Reply_MainListReq = "Reply.MainListReq"
			ProtoType.Reply_MainListReply = "Reply.MainListReply"
			ProtoType.Metadata_FawkesReq = "Metadata.FawkesReq"
			ProtoType.Metadata = "Metadata"
			ProtoType.Device = "Device"
			ProtoType.Metadata_FawkesReq = "Metadata.FawkesReq"

			fds, _ := parser.ParseFiles("bilibili/main/community/reply/v1.proto")
			fd := fds[0]
			protoMap[ProtoType.Reply_MainListReply] = fd.FindMessage("bilibili.main.community.reply.v1.MainListReply")
			protoMap[ProtoType.Reply_MainListReq] = fd.FindMessage("bilibili.main.community.reply.v1.MainListReq")

			fds, _ = parser.ParseFiles("bilibili/metadata/fawkes.proto")
			protoMap[ProtoType.Metadata_FawkesReq] = fds[0].FindMessage("bilibili.metadata.fawkes.FawkesReq")
			fds, _ = parser.ParseFiles("bilibili/metadata.proto")
			protoMap[ProtoType.Metadata] = fds[0].FindMessage("bilibili.metadata.Metadata")
			fds, _ = parser.ParseFiles("bilibili/metadata/device.proto")
			protoMap[ProtoType.Device] = fds[0].FindMessage("bilibili.metadata.device.Device")
		}
	}
	if client.Options.HttpProxy != "" {
		proxyURL, _ := url.Parse(fmt.Sprintf("http://%s:%s@%s", client.Options.ProxyUser, client.Options.ProxyPass, client.Options.HttpProxy))
		client.Resty.SetProxy(proxyURL.String())
	}
	client.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
	client.Resty.SetCookieJar(nil)
	client.Resty.OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
		if strings.Contains(request.URL, "bilibili.main") {
			var buvid = getBUVID()
			var fp = getFP()
			//var spilt = strings.Split(request.URL, "/")
			//request.SetHeader(":authority", "app.bilibili.com")
			//request.SetHeader(":method", "POST")
			//request.SetHeader(":path", "/"+spilt[3]+"/"+spilt[4])
			//request.SetHeader(":scheme", "https")
			request.SetHeader("accept-encoding", "identity")
			request.SetHeader("grpc-encoding", "gzip")
			request.SetHeader("grpc-accept-encoding", "gzip")
			request.SetHeader("env", "prod")
			request.SetHeader("app-key", "android")
			request.SetHeader("user-agent", "")
			request.SetHeader("x-bili-aurora-eid", getAurora(uint64(client.UID)))
			request.SetHeader("x-bili-mid", toString(client.UID))
			request.SetHeader("x-bili-aurora-zone", "")
			request.SetHeader("x-bili-gaia-vtoken", "")
			request.SetHeader("x-bili-ticket", "")
			request.SetHeader("x-bili-metadata-bin", getMetadata())
			request.SetHeader("x-bili-device-bin", getDevice(buvid, fp))
			request.SetHeader("x-bili-network-bin", "CAEaBjQ2MDAwMA")
			request.SetHeader("x-bili-restriction-bin", "")
			request.SetHeader("x-bili-locale-bin", "CggKAnpoGgJDThIICgJ6aBoCQ04")
			request.SetHeader("x-bili-exps-bin", "")
			request.SetHeader("buvid", buvid)
			request.SetHeader("x-bili-fawkes-req-bin", getFawkes())
			request.SetHeader("content-type", "application/grpc")
			var payload = request.Body.([]byte)

			frame := make([]byte, 5+len(payload))
			frame[0] = 0x00
			binary.BigEndian.PutUint32(frame[1:5], uint32(len(payload)))
			copy(frame[5:], payload)
			request.SetBody(frame)

		} else {

			if client.Options.RandomUserAgent {
				request.Header.Set("User-Agent", randomUserAgent())
			} else {
				request.Header.Set("User-Agent", client.UserAgent)
			}
			request.SetHeader("accept-language", "zh-CN,zh;q=0.9")
			request.SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")

		}

		var ref = ""
		if strings.Contains(request.URL, "electron") {
			ref = "client"
		}
		if !client.Options.NoCookie {
			request.Header.Set("Cookie", client.Cookie)
		}
		request.Header.Set("Referer", "https://www.bilibili.com/"+ref)

		return nil
	})
	client.Resty.OnAfterResponse(func(_ *resty.Client, response *resty.Response) error {
		url0 := response.Request.URL
		if strings.Contains(url0, "bilibili.main") {
			var raw = response.Body()[5:]
			var dist []byte
			if response.Body()[0:][0] == 0x01 {
				gr, _ := gzip.NewReader(bytes.NewReader(raw)) //gzip
				dist, _ = io.ReadAll(gr)
			} else {
				dist = raw
			}

			var typo = ""
			if strings.Contains(url0, "Reply/MainList") {

				typo = ProtoType.Reply_MainListReply
			}
			msg := dynamic.NewMessage(protoMap[typo])
			err = msg.Unmarshal(dist)
			if err != nil {
				panic(err)
			}
			jsonBytes, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				panic(err)
			}

			response.SetBody(jsonBytes)
		}
		return nil
	})
	if cookie == "" && !client.Options.NoCookie {
		r, _ := client.Resty.R().Get("https://space.bilibili.com/208259/")
		for _, s := range r.Header().Values("Set-Cookie") {
			cookie += strings.Split(s, ";")[0] + ";"
		}
		if cookie != "" {
			cookie = cookie[0 : len(cookie)-1]
		}

	}
	client.Cookie = cookie
	client.WBI = NewDefaultWbi()
	client.WBI.WithRawCookies(cookie)
	client.WBI.doInitWbi()
	client.UID = client.selfUID()
	client.Address = client.GetLocation().Address
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
	res, err := client.Resty.R().SetHeader("Cookie", client.Cookie).Get("https://api.bilibili.com/x/web-interface/nav")

	if err != nil {
		fmt.Println(err)
	}
	var self = SelfInfo{}
	json.Unmarshal(res.Body(), &self)
	return self.Data.Mid
}

func (client *BiliClient) ResetResty(cookie ...string) {
	transport, _ := client.Resty.Transport()
	transport.CloseIdleConnections()
	client.Resty = resty.New()

	if len(cookie) > 0 {
		setupClient(client, cookie[0])
		client.Cookie = ""
	} else {
		setupClient(client, client.Cookie)
	}

}
func randomBrowserVersion(browser string) string {
	majorVersion := 117 + rand.Intn(3) // 主版本号范围，例如117到119
	minorVersion := rand.Intn(1000)    // 次版本号可在0到999之间
	return fmt.Sprintf("%s/%d.%d", browser, majorVersion, minorVersion)
}
func randomUserAgent() string {
	var operatingSystems = []string{
		"Windows NT 10.0; Win64; x64",
		"Macintosh; Intel Mac OS X 12_6",
		"Linux; Android 13; Pixel 6",
		"Linux; U; Android 13; SM-G991U",
		"X11; Linux x86_64",
	}

	var devices = []string{
		"Mobile Safari/537.36",
		"Safari/537.36",
		"Mobile/15E148 Safari/604.1",
		"Safari/604.1",
	}
	os := operatingSystems[rand.Intn(len(operatingSystems))]
	device := devices[rand.Intn(len(devices))]

	// 定义浏览器类型
	browserTypes := []string{"Chrome", "Edge", "Firefox", "Safari"}
	browser := browserTypes[rand.Intn(len(browserTypes))]

	browserVersion := randomBrowserVersion(browser)

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) %s %s", os, browserVersion, device)
}
