package bili

type GuardListResponse struct {
	Data struct {
		List []GuardResponseItem `json:"list"`
		Top  []GuardResponseItem `json:"top3"`
		Info struct {
			Total int `json:"num"`
		} `json:"info"`
	} `json:"data"`
}
type GuardResponseItem struct {
	Days int16 `json:"accompany"`
	Info struct {
		UID  int64 `json:"uid"`
		User struct {
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"base"`
		Medal struct {
			Name       string `json:"name"`
			Level      int8   `json:"level"`
			ColorDec   int    `json:"color_start"`
			GuardLevel int8   `json:"guard_level"`
			Color      string `json:"v2_medal_color_start"`
		} `json:"medal"`
	} `json:"uinfo"`
}
type FansClubResponse struct {
	Message string `json:"message"`
	Data    struct {
		Item []struct {
			UID   int64  `json:"uid"`
			UName string `json:"name"`
			Score int    `json:"score"`
			Level int8   `json:"level"`
			Medal struct {
				Type  int8   `json:"guard_level"`
				Name  string `json:"name"`
				Level int8   `json:"level"`
			} `json:"uinfo_medal"`
		} `json:"item"`
	} `json:"data"`
}
type OnlineWatcherResponse struct {
	Data struct {
		Item []struct {
			UID   int64  `json:"uid"`
			Name  string `json:"name"`
			Face  string `json:"face"`
			Guard int8   `json:"guard_level"`
			Days  int16  `json:"days"`
			UInfo struct {
				Medal struct {
					Color string `json:"v2_medal_color_start"`
					Name  string `json:"name"`
					Level int8   `json:"level"`
				} `json:"medal"`
			} `json:"uinfo"`
		} `json:"item"`
		Count int `json:"count"`
	} `json:"data"`
}
type LiveStreamResponse struct {
	Data struct {
		Time        int64 `json:"live_time"`
		PlayurlInfo struct {
			Playurl struct {
				Stream []struct {
					ProtocolName string `json:"protocol_name"`
					Format       []struct {
						FormatName string `json:"format_name"`
						Codec      []struct {
							CodecName string `json:"codec_name"`
							CurrentQn int    `json:"current_qn"`
							AcceptQn  []int  `json:"accept_qn"`
							BaseUrl   string `json:"base_url"`
							UrlInfo   []struct {
								Host      string `json:"host"`
								Extra     string `json:"extra"`
								StreamTtl int    `json:"stream_ttl"`
							} `json:"url_info"`
							HdrQn     interface{} `json:"hdr_qn"`
							DolbyType int         `json:"dolby_type"`
							AttrName  string      `json:"attr_name"`
							HdrType   int         `json:"hdr_type"`
						} `json:"codec"`
						MasterUrl string `json:"master_url"`
					} `json:"format"`
				} `json:"stream"`
			} `json:"playurl"`
		} `json:"playurl_info"`
	} `json:"data"`
}
type CollectionList struct {
	Data struct {
		List []struct {
			Title string `json:"title"`
			ID    int    `json:"id"`
		}
	}
}
type FansList struct {
	Data struct {
		List []struct {
			Mid                string `json:"mid"`
			Attribute          int    `json:"attribute"`
			Uname              string `json:"uname"`
			Face               string `json:"face"`
			AttestationDisplay struct {
				Type int    `json:"type"`
				Desc string `json:"desc"`
			} `json:"attestation_display"`
		} `json:"list"`
	} `json:"data"`
	Ts        int64  `json:"ts"`
	RequestID string `json:"request_id"`
}
type UserDynamic struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Items  []DynamicItem `json:"items"`
		Offset string        `json:"offset"`
	} `json:"data"`
}
type DynamicItem struct {
	IDStr   string       `json:"id_str"`
	Orig    *DynamicItem `json:"orig"`
	Modules struct {
		ModuleDynamic struct {
			Major struct {
				Archive struct {
					Aid   string `json:"aid"`
					Badge struct {
						BgColor string      `json:"bg_color"`
						Color   string      `json:"color"`
						IconURL interface{} `json:"icon_url"`
						Text    string      `json:"text"`
					} `json:"badge"`
					Bvid  string `json:"bvid"`
					Cover string `json:"cover"`
					Desc  string `json:"desc"`
					Stat  struct {
						Danmaku string `json:"danmaku"`
						Play    string `json:"play"`
					} `json:"stat"`
					Title string `json:"title"`
					Type  int    `json:"type"`
				} `json:"archive"`
				Opus struct {
					Pics []struct {
						URL string `json:"url"`
					} `json:"pics"`
					Summary struct {
						Text string `json:"text"`
					} `json:"summary"`
				} `json:"opus"`
				Desc struct {
					Text string `json:"text"`
				} `json:"desc"`
				Type string `json:"type"`
			} `json:"major"`
			Topic interface{} `json:"topic"`
			Desc  struct {
				Nodes []struct {
					Text string `json:"text"`
				} `json:"rich_text_nodes"`
			} `json:"desc"`
		} `json:"module_dynamic"`
		ModuleAuthor struct {
			Name      string `json:"name"`
			Mid       int64  `json:"mid"`
			TimeStamp int64  `json:"pub_ts"`
		} `json:"module_author"`
	} `json:"modules"`
	Type string `json:"type"`
}
type VideoResponse struct {
	Data struct {
		Cover     string `json:"pic"`
		Title     string `json:"title"`
		Duration  int    `json:"duration"`
		PublishAt int64  `json:"pubdate"`
		Desc      string `json:"desc"`
		Owner     struct {
			Mid  int64  `json:"mid"`
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"owner"`
		Stat struct {
			View     int `json:"view"`
			Reply    int `json:"reply"`
			Coin     int `json:"coin"`
			Share    int `json:"share"`
			Like     int `json:"like"`
			Danmaku  int `json:"danmaku"`
			Favorite int `json:"favorite"`
		} `json:"stat"`
		Pages []struct {
			Cid      int    `json:"cid"`
			Title    string `json:"part"`
			Duration int    `json:"duration"`
		}
	} `json:"data"`
}
type VideoListResponse struct {
	Message string `json:"message"`
	Data    struct {
		List struct {
			Vlist []struct {
				Comment     int    `json:"comment"`
				Typeid      int    `json:"typeid"`
				Play        int    `json:"play"`
				Pic         string `json:"pic"`
				Subtitle    string `json:"subtitle"`
				Description string `json:"description"`
				Title       string `json:"title"`
				Review      int    `json:"review"`
				Author      string `json:"author"`
				Mid         int    `json:"mid"`
				Created     int    `json:"created"`
				Length      string `json:"length"`
				VideoReview int    `json:"video_review"`
				Bvid        string `json:"bvid"`
			} `json:"vlist"`
		} `json:"list"`
		Page struct {
			Pn    int `json:"pn"`
			Ps    int `json:"ps"`
			Count int `json:"count"`
		} `json:"page"`
		EpisodicButton struct {
			Text string `json:"text"`
			Uri  string `json:"uri"`
		} `json:"episodic_button"`
		IsRisk      bool        `json:"is_risk"`
		GaiaResType int         `json:"gaia_res_type"`
		GaiaData    interface{} `json:"gaia_data"`
	} `json:"data"`
}
type AreaLiverListResponse struct {
	Data struct {
		More int8 `json:"has_more"`
		List []struct {
			Room       int    `json:"roomid"`
			ParentArea string `json:"parent_name"`
			Area       string `json:"area_name"`
			Title      string `json:"title"`
			UName      string `json:"uname"`
			UID        int64  `json:"uid"`
			Watch      struct {
				Num int `json:"num"`
			} `json:"watched_show"`
		} `json:"list"`
	} `json:"data"`
}
type AreaListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Data []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
			List []struct {
				Id         string `json:"id"`
				ParentId   string `json:"parent_id"`
				ParentName string `json:"parent_name"`
				Name       string `json:"name"`
				Icon       string `json:"pic"`
			} `json:"list"`
		} `json:"data"`
		Expid int `json:"expid"`
	} `json:"data"`
}
type Dash struct {
	Data struct {
		Dash0 struct {
			Video []struct {
				Link string `json:"base_url"`
			} `json:"video"`
			Audio []struct {
				Link string `json:"base_url"`
			} `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
}

// LIVE
type LiveInfo struct {
	Data struct {
		Group            string  `json:"group"`
		BusinessID       int     `json:"business_id"`
		RefreshRowFactor float64 `json:"refresh_row_factor"`
		RefreshRate      int     `json:"refresh_rate"`
		MaxDelay         int     `json:"max_delay"`
		Token            string  `json:"token"`
		HostList         []struct {
			Host    string `json:"host"`
			Port    int    `json:"port"`
			WssPort int    `json:"wss_port"`
			WsPort  int    `json:"ws_port"`
		} `json:"host_list"`
	} `json:"data"`
}

type GiftList struct {
	Data struct {
		GiftConfig struct {
			BaseConfig struct {
				List []struct {
					ID      int    `json:"id"`
					Name    string `json:"name"`
					Price   int    `json:"price"`
					Picture string `json:"webp"`
				} `json:"list"`
			} `json:"base_config"`
			RoomConfig []struct {
				Name  string `json:"name"`
				Price int    `json:"price"`
			} `json:"room_config"`
		} `json:"gift_config"`
	} `json:"data"`
}
type GiftInfo struct {
	Cmd  string `json:"cmd"`
	Data struct {
		GiftName string `json:"giftName"`
		Num      int    `json:"num"`
		Price    int    `json:"price"`
		Parent   struct {
			Price    int    `json:"original_gift_price"`
			GiftName string `json:"original_gift_name"`
		} `json:"blind_gift"`
		ReceiveUserInfo struct {
			UID   int    `json:"uid"`
			Uname string `json:"uname"`
		} `json:"receive_user_info"`
		SenderUinfo struct {
			Base struct {
				Name string `json:"name"`
			} `json:"base"`
			UID int `json:"uid"`
		} `json:"sender_uinfo"`
		UID   int `json:"uid"`
		Medal struct {
			Name  string `json:"name"`
			Level int    `json:"level"`
			Color int    `json:"medal_color"`
		}
		Uname string `json:"uname"`
		Face  string `json:"face"`
	} `json:"data"`
}
type LiveText struct {
	Cmd  string        `json:"cmd"`
	DmV2 string        `json:"dm_v2"`
	Info []interface{} `json:"info"`
}
type GiftBox struct {
	Data struct {
		Gifts []struct {
			Price    int    `json:"price"`
			GiftName string `json:"gift_name"`
			Picture  string `json:"webp"`
		} `json:"gifts"`
	} `json:"data"`
}
type LiveAction struct {
	FromName   string
	FromId     string
	LiveRoom   string
	ActionName string
	GiftName   string
	GiftPrice  float32
	GiftAmount int16
	Extra      string
	MedalName  string
	MedalLevel int8
	GuardLevel int8
}
type FrontLiveAction struct {
	LiveAction
	Face        string
	UUID        string
	MedalColor  string
	GiftPicture string
}
type RoomInfo struct {
	Data struct {
		LiveTime     string `json:"live_time"`
		UID          int    `json:"uid"`
		Title        string `json:"title"`
		Area         string `json:"area_name"`
		AreaId       int    `json:"area_id"`
		ParentAreaId int    `json:"parent_area_id"`
		Face         string `json:"user_cover"`
	} `json:"data"`
}
type Live struct {
	Title    string
	StartAt  int64
	EndAt    int64
	UserName string
	UserID   string
	Area     string
	RoomId   int
	Money    float64 `gorm:"type:decimal(7,2)"`
	Message  int
	Watch    int
}
type EnterLive struct {
	Cmd  string `json:"cmd"`
	Data struct {
		UID       int    `json:"uid"`
		Uname     string `json:"uname"`
		FansMedal struct {
			MedalName string `json:"medal_name"`
			Level     int    `json:"medal_level"`
		} `json:"fans_medal"`
	} `json:"data"`
}

type LiverInfo struct {
	Data struct {
		Info struct {
			Uname string `json:"uname"`
		} `json:"info"`
	} `json:"data"`
}

type SuperChatInfo struct {
	Data struct {
		Message  string  `json:"message"`
		Price    float64 `json:"price"`
		Uid      int     `json:"uid"`
		UserInfo struct {
			Uname string `json:"uname"`
		} `json:"user_info"`
	} `json:"data"`
}

type GuardInfo struct {
	Data struct {
		Uid        int    `json:"uid"`
		Username   string `json:"username"`
		GuardLevel int    `json:"guard_level"`
		Num        int    `json:"num"`
		GiftName   string `json:"gift_name"`
	} `json:"data"`
}
type Watched struct {
	Data struct {
		Num       int    `json:"num"`
		TextSmall string `json:"text_small"`
		TextLarge string `json:"text_large"`
	} `json:"data"`
}
type SelfInfo struct {
	Data struct {
		Mid int64 `json:"mid"`
	} `json:"data"`
}
type CommentResponse struct {
	Message string `json:"message"`
	Data    struct {
		Cursor struct {
			IsBegin         bool `json:"is_begin"`
			Prev            int  `json:"prev"`
			Next            int  `json:"next"`
			IsEnd           bool `json:"is_end"`
			PaginationReply struct {
				NextOffset string `json:"next_offset"`
			} `json:"pagination_reply"`
			SessionId   string `json:"session_id"`
			Mode        int    `json:"mode"`
			ModeText    string `json:"mode_text"`
			AllCount    int    `json:"all_count"`
			SupportMode []int  `json:"support_mode"`
			Name        string `json:"name"`
		} `json:"cursor"`
		Replies []struct {
			Rpid      int64  `json:"rpid"`
			Oid       int64  `json:"oid"`
			Type      int    `json:"type"`
			Mid       int64  `json:"mid"`
			Root      int    `json:"root"`
			Parent    int    `json:"parent"`
			Dialog    int    `json:"dialog"`
			Count     int    `json:"count"`
			Rcount    int    `json:"rcount"`
			State     int    `json:"state"`
			Fansgrade int    `json:"fansgrade"`
			Attr      int    `json:"attr"`
			Ctime     int    `json:"ctime"`
			MidStr    string `json:"mid_str"`
			OidStr    string `json:"oid_str"`
			RpidStr   string `json:"rpid_str"`
			RootStr   string `json:"root_str"`
			ParentStr string `json:"parent_str"`
			DialogStr string `json:"dialog_str"`
			Like      int    `json:"like"`
			Action    int    `json:"action"`
			Member    struct {
				Mid       string `json:"mid"`
				Uname     string `json:"uname"`
				Avatar    string `json:"avatar"`
				LevelInfo struct {
					CurrentLevel int `json:"current_level"`
				} `json:"level_info"`
			} `json:"member"`
			Content struct {
				Message string        `json:"message"`
				Members []interface{} `json:"members"`
				JumpUrl struct {
				} `json:"jump_url"`
				Pictures []struct {
					ImgSrc string `json:"img_src"`
				} `json:"pictures,omitempty"`
				PictureScale float64 `json:"picture_scale,omitempty"`
			} `json:"content"`
			Replies      []interface{} `json:"replies"`
			DynamicIdStr string        `json:"dynamic_id_str"`
			DynamicId    int64         `json:"dynamic_id,omitempty"`
		} `json:"replies"`
		Upper struct {
			Mid int `json:"mid"`
		} `json:"upper"`
		Note             int         `json:"note"`
		EsportsGradeCard interface{} `json:"esports_grade_card"`
		Callbacks        interface{} `json:"callbacks"`
		ContextFeature   string      `json:"context_feature"`
	} `json:"data"`
}
