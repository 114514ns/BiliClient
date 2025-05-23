package bili

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"math"
	"math/rand"
	url2 "net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func FormatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	secs := (duration % time.Minute) / time.Second

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
func GenerateBase64RandomString(minLength, maxLength int) string {
	// Initialize random source
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate random length
	length := minLength
	if maxLength > minLength {
		length += r.Intn(maxLength - minLength + 1)
	}

	// Generate random bytes
	randomBytes := make([]byte, length)
	for i := 0; i < length; i++ {
		// Generate values between 0x20 and 0x7F
		randomBytes[i] = byte(r.Intn(0x60) + 0x20)
	}

	// Encode to base64 and return
	return base64.StdEncoding.EncodeToString(randomBytes)
}
func nextNumber(s string, index int) int64 {
	var sum int64 = 0
	for i := index; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			sum = sum*10 + int64(s[i]-'0')
		} else {
			if sum != 0 {
				break
			}
		}
	}
	return sum
}
func appendDM(url0 string, client *BiliClient) string {
	u, _ := url2.Parse(url0)
	htmlRes, _ := client.Resty.R().Get("https://space.bilibili.com/2/upload/video")
	reader := bytes.NewReader(htmlRes.Body())
	root, _ := html.Parse(reader)
	find := goquery.NewDocumentFromNode(root).Find("script")
	access := ""
	dmImgStr := GenerateBase64RandomString(16, 64)
	dmImgStr = dmImgStr[0 : len(dmImgStr)-2]
	cover := GenerateBase64RandomString(32, 128)
	cover = cover[0 : len(cover)-2]
	inter := `{"ds":[],"wh":[0,0,0],"of":[0,0,0]}`
	find.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "access_id") {
			access, _ = url2.PathUnescape(s.Text())
		}
	})
	var i interface{}
	json.Unmarshal([]byte(access), &i)
	m, _ := i.(map[string]interface{})
	access, _ = m["access_id"].(string)
	u.RawQuery = fmt.Sprintf(u.RawQuery+"&w_webid=%s&dm_img_list=[]&dm_img_str=%s&dm_cover_img_str=%s&dm_img_inter=%s", access, dmImgStr, cover, inter)
	return u.String()
}
func appendWts(url0 string, client *BiliClient) string {
	u, _ := url2.Parse(url0)
	signed, _ := client.WBI.SignQuery(u.Query(), time.Now())
	return "https://" + u.Host + "/" + u.Path + "?" + signed.Encode()
}
func toString(i int64) string {
	return strconv.FormatInt(i, 10)
}
func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

type JsonType struct {
	s   string
	i   int
	i64 int64
	f32 float32
	f64 float64
}

func getInt(obj interface{}, path string) int {
	return getObject(obj, path, "int").i
}
func getInt64(obj interface{}, path string) int64 {
	return getObject(obj, path, "int64").i64
}
func getString(obj interface{}, path string) string {
	return getObject(obj, path, "string").s
}
func getObject(obj interface{}, path string, typo string) JsonType {
	var array = strings.Split(path, ".")
	inner := obj.(map[string]interface{})
	var st = JsonType{}
	for i, s := range array {
		if i == len(array)-1 {

			value := inner[s]
			if value != nil {
				var t = reflect.TypeOf(value)
				if t.Kind() == reflect.String {
					st.s = value.(string)
				}
				if t.Kind() == reflect.Int {
					st.i, _ = value.(int)
				}
				if t.Kind() == reflect.Int64 {
					if value.(int64) > math.MaxInt {
						st.i64 = value.(int64)
					} else {
						st.i = value.(int)
					}

				}
				if t.Kind() == reflect.Float64 {
					if typo == "int" {
						st.i = int(value.(float64))
					}
					if typo == "int64" {
						st.i64 = int64(value.(float64))
					}
				}
			}

			return st
		} else {
			inner = inner[s].(map[string]interface{})
		}
	}
	return st
}
