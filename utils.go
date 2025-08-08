package bili

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jhump/protoreflect/dynamic"
	"golang.org/x/net/html"
	"math"
	"math/rand"
	url2 "net/url"
	"os"
	"path/filepath"
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
	s     string
	i     int
	i64   int64
	f32   float32
	f64   float64
	array []interface{}
	v     bool
	m     map[string]interface{}
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
func getArray(obj interface{}, path string) []interface{} {
	return getObject(obj, path, "array").array
}
func getBool(obj interface{}, path string) bool {
	return getObject(obj, path, "bool").v
}
func getObject(obj interface{}, path string, typo string) JsonType {
	var array = strings.Split(path, ".")
	inner, ok := obj.(map[string]interface{})
	if !ok {
		return JsonType{}
	}
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
				if t.Kind() == reflect.Slice {
					if typo == "array" {
						st.array = value.([]interface{})
					}
				}
				if t.Kind() == reflect.Bool {
					st.v = value.(bool)
				}
				if t.Kind() == reflect.Map {
					st.m = value.(map[string]interface{})
				}
			}

			return st
		} else {

			if inner[s] == nil {
				return st
			}
			inner = inner[s].(map[string]interface{})
		}
	}
	return st
}

// from https://socialsisteryi.github.io/bilibili-API-collect/docs/misc/bvid_desc.html#golang
var (
	XOR_CODE = int64(23442827791579)
	MAX_CODE = int64(2251799813685247)
	CHARTS   = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
	PAUL_NUM = int64(58)
)

func swapString(s string, x, y int) string {
	chars := []rune(s)
	chars[x], chars[y] = chars[y], chars[x]
	return string(chars)
}

func Bvid2Avid(bvid string) (avid int64) {
	s := swapString(swapString(bvid, 3, 9), 4, 7)
	bv1 := string([]rune(s)[3:])
	temp := int64(0)
	for _, c := range bv1 {
		idx := strings.IndexRune(CHARTS, c)
		temp = temp*PAUL_NUM + int64(idx)
	}
	avid = (temp & MAX_CODE) ^ XOR_CODE
	return
}

func Avid2Bvid(avid int64) (bvid string) {
	arr := [12]string{"B", "V", "1"}
	bvIdx := len(arr) - 1
	temp := (avid | (MAX_CODE + 1)) ^ XOR_CODE
	for temp > 0 {
		idx := temp % PAUL_NUM
		arr[bvIdx] = string(CHARTS[idx])
		temp /= PAUL_NUM
		bvIdx--
	}
	raw := strings.Join(arr[:], "")
	bvid = swapString(swapString(raw, 3, 9), 4, 7)
	return
}

func ChunkSlice[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	var result [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}
	return result
}
func getAurora(uid uint64) string {
	if uid == 0 {
		return ""
	}

	// 1. 将 UID 转为字符串再转为字节数组
	midBytes := []byte(strconv.FormatUint(uid, 10))

	// 2. 和 "ad1va46a7lza" 每字节进行异或
	key := []byte("ad1va46a7lza")
	resultBytes := make([]byte, len(midBytes))
	for i, b := range midBytes {
		resultBytes[i] = b ^ key[i%len(key)]
	}

	// 3. Base64 编码，去掉 padding
	encoded := base64.RawStdEncoding.EncodeToString(resultBytes)
	return encoded
}
func getFawkes() string {
	msg := dynamic.NewMessage(protoMap[ProtoType.Metadata_FawkesReq])
	msg.TrySetFieldByName("appkey", "android")
	msg.TrySetFieldByName("env", "prod")
	msg.TrySetFieldByName("session_id", randomHex(8))
	bytes, _ := msg.Marshal()
	return base64.StdEncoding.EncodeToString(bytes)
}
func getMetadata() string {
	msg := dynamic.NewMessage(protoMap[ProtoType.Metadata])
	msg.TrySetFieldByName("mobi_app", "android")
	msg.TrySetFieldByName("build", 8430300)
	msg.TrySetFieldByName("channel", "alifenfa")
	msg.TrySetFieldByName("platform", "android")
	msg.TrySetFieldByName("buvid", "")
	bytes, _ := msg.Marshal()
	return base64.StdEncoding.EncodeToString(bytes)
}
func getBUVID() string {
	var hash = strings.ToUpper(randomHex(32))
	var chars = strings.Split(hash, "")
	return "XU" + chars[2] + chars[12] + chars[22] + hash
}
func getFP() string {
	return randomHex(64)
}
func getDevice(buvid string, fp string) string {
	msg := dynamic.NewMessage(protoMap[ProtoType.Device])
	msg.TrySetFieldByName("app_id", 1)
	msg.TrySetFieldByName("build", 7420400)
	msg.TrySetFieldByName("buvid", buvid)
	msg.TrySetFieldByName("platform", "android")
	msg.TrySetFieldByName("mobi_app", "android")
	msg.TrySetFieldByName("device", "")
	msg.TrySetFieldByName("channel", "alifenfa")
	msg.TrySetFieldByName("brand", "XIAOMI")
	msg.TrySetFieldByName("model", "android")
	msg.TrySetFieldByName("osver", "15")
	msg.TrySetFieldByName("fp_local", fp)
	msg.TrySetFieldByName("fp_remote", fp)
	msg.TrySetFieldByName("version_name", "8.43.0")
	msg.TrySetFieldByName("fp", fp)
	msg.TrySetFieldByName("fts", time.Now().Unix())
	msg.TrySetFieldByName("guest_id", "")
	bytes, _ := msg.Marshal()
	return base64.StdEncoding.EncodeToString(bytes)

}
func randomHex(length int) string {
	var table = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	var hex = ""
	for i := 0; i < length; i++ {
		hex = hex + table[rand.Intn(len(table))]
	}
	return hex
}
func collectProtoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
