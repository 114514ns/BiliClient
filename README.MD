# BiliClient
# Intro

## Usage
### 创建Client
`创建默认Cookie的Client`
~~~go

client := NewClient(cookie, ClientOptions{

HttpProxy: "0.0.0.0:8080",
ProxyUser: "admin",
ProxyPass: "PZ8u9Pr2oz",

RandomUserAgent: true,
})
~~~
`使用指定Cookie创建Client`
~~~go
client := NewClient("")
~~~
### 直播相关

~~~go
func (client BiliClient) GetGuardByPage(room string, liver string, page int) []LiveUser{} //分页获取大航海成员
func (client BiliClient) GetFansClub(liver string, delay int, onError func(msg string)){} []LiveUser //分页获取粉丝团成员
func (client BiliClient) GetOnline(room string, liver string){} []LiveUser //获取直播间在线用户
func (client BiliClient) GetLiveStream(room string){} string //获取直播流
func (client BiliClient) TraceLive(room string, onMessage func(action FrontLiveAction), onChange func(state string)) {} //当直播间有消息或者状态改变时会触发回调
~~~


### 视频相关
~~~go
func (client BiliClient) GetVideo(bv string){} (result []Video) //解析视频，如果有分p视频，则一并返回
func (client BiliClient) GetVideoByUser(mid int64, page int, byHot bool){} (result []Video) //获取用户的稿件
func (client BiliClient) GetVideoStream(bv string, part int) []string {} //获取视频流
func (client *BiliClient) SearchVideo(opinion SearchOption) []Video {} //搜索视频
~~~

### 空间相关
~~~go
func (client BiliClient) GetCollection(user string, page int){} map[string]string //获取用户收藏夹列表
func (client BiliClient) GetFollowing(user string, delay int){} map[string]string  //获取用户关注列表
func (client *BiliClient) GetDynamicsByUser(user string,cursor string){} []Archive string //获取用户动态
func (client *BiliClient) SetAnnouce(content string){} //设置主页公告
func (client *BiliClient) GetFansMedal(id string){} []Medal //获取用户粉丝牌
func (client *BiliClient) GetComment(oid int64, cursor string, type0 int){} []Comment //获取评论
func (client *BiliClient) GetStats(uid int64){} map[string]int //获取用户稿件播放量和关注，粉丝量
func (client *BiliClient) GetLocation(){} Location //获取当前ip信息
func (client *BiliClient) GetLocation(ip string){} Location //获取指定ip信息
func (client *BiliClient) GetArticle(cv int64) Archive{} //获取专栏信息和内容
~~~

具体用法参考main.test.go

# Referer
接口文档：[bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect)
<br/>
wbi算法：[CuteReimu/bilibili/](https://github.com/CuteReimu/bilibili/)
<br/>
dm算法：[src](https://github.com/guozhigq/pilipala/blob/34fb5802710e4ccd21a4b536518595a6580d40d5/lib/http/member.dart#L77)