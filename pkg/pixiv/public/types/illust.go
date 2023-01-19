package pixiv_public_types

import "time"

type IllustURLs struct {
	Mini     string `json:"mini"`
	Thumb    string `json:"thumb"`
	Small    string `json:"small"`
	Regular  string `json:"regular"`
	Original string `json:"original"`
}

type IllustFanboxPromotion struct {
	UserName        string `json:"userName"`
	UserImageURL    string `json:"userImageUrl"`
	ContentURL      string `json:"contentUrl"`
	Description     string `json:"description"`
	ImageURL        string `json:"imageUrl"`
	ImageURLMobile  string `json:"imageUrlMobile"`
	HasAdultContent bool   `json:"hasAdultContent"`
}

type IllustExtraData struct {
	Meta struct {
		Title              string `json:"title"`
		Description        string `json:"description"`
		Canonical          string `json:"canonical"`
		AlternateLanguages struct {
			Ja string `json:"ja"`
			En string `json:"en"`
		} `json:"alternateLanguages"`
		DescriptionHeader string `json:"descriptionHeader"`
		Ogp               struct {
			Description string `json:"description"`
			Image       string `json:"image"`
			Title       string `json:"title"`
			Type        string `json:"type"`
		} `json:"ogp"`
		Twitter struct {
			Description string `json:"description"`
			Image       string `json:"image"`
			Title       string `json:"title"`
			Card        string `json:"card"`
		} `json:"twitter"`
	} `json:"meta"`
}

type IllustZoneConfigItem struct {
	URL string `json:"url"`
}

type IllustZoneConfig struct {
	Responsive     IllustZoneConfigItem `json:"responsive"`
	Rectangle      IllustZoneConfigItem `json:"rectangle"`
	Five00X500     IllustZoneConfigItem `json:"500x500"`
	Header         IllustZoneConfigItem `json:"header"`
	Footer         IllustZoneConfigItem `json:"footer"`
	ExpandedFooter IllustZoneConfigItem `json:"expandedFooter"`
	Logo           IllustZoneConfigItem `json:"logo"`
	Relatedworks   IllustZoneConfigItem `json:"relatedworks"`
}

type IllustTitleCaptionTranslation struct {
	WorkTitle   interface{} `json:"workTitle"`
	WorkCaption interface{} `json:"workCaption"`
}

type Illust struct {
	IllustID                string                         `json:"illustId"`
	IllustTitle             string                         `json:"illustTitle"`
	IllustComment           string                         `json:"illustComment"`
	ID                      string                         `json:"id"`
	Title                   string                         `json:"title"`
	Description             string                         `json:"description"`
	IllustType              int                            `json:"illustType"`
	CreateDate              time.Time                      `json:"createDate"`
	UploadDate              time.Time                      `json:"uploadDate"`
	Restrict                int                            `json:"restrict"`
	XRestrict               int                            `json:"xRestrict"`
	Sl                      int                            `json:"sl"`
	Urls                    IllustURLs                     `json:"urls"`
	Tags                    *Tags                          `json:"tags"`
	Alt                     string                         `json:"alt"`
	StorableTags            []string                       `json:"storableTags"`
	UserID                  string                         `json:"userId"`
	UserName                string                         `json:"userName"`
	UserAccount             string                         `json:"userAccount"`
	UserIllusts             map[string]*UserIllusts        `json:"userIllusts"`
	LikeData                bool                           `json:"likeData"`
	Width                   int                            `json:"width"`
	Height                  int                            `json:"height"`
	PageCount               int                            `json:"pageCount"`
	BookmarkCount           int                            `json:"bookmarkCount"`
	LikeCount               int                            `json:"likeCount"`
	CommentCount            int                            `json:"commentCount"`
	ResponseCount           int                            `json:"responseCount"`
	ViewCount               int                            `json:"viewCount"`
	BookStyle               any                            `json:"bookStyle"`
	IsHowto                 bool                           `json:"isHowto"`
	IsOriginal              bool                           `json:"isOriginal"`
	ImageResponseOutData    []interface{}                  `json:"imageResponseOutData"`
	ImageResponseData       []interface{}                  `json:"imageResponseData"`
	ImageResponseCount      int                            `json:"imageResponseCount"`
	PollData                interface{}                    `json:"pollData"`
	SeriesNavData           interface{}                    `json:"seriesNavData"`
	DescriptionBoothID      interface{}                    `json:"descriptionBoothId"`
	DescriptionYoutubeID    interface{}                    `json:"descriptionYoutubeId"`
	ComicPromotion          interface{}                    `json:"comicPromotion"`
	FanboxPromotion         *IllustFanboxPromotion         `json:"fanboxPromotion"`
	ContestBanners          []interface{}                  `json:"contestBanners"`
	IsBookmarkable          bool                           `json:"isBookmarkable"`
	BookmarkData            interface{}                    `json:"bookmarkData"`
	ContestData             interface{}                    `json:"contestData"`
	ZoneConfig              *IllustZoneConfig              `json:"zoneConfig"`
	ExtraData               *IllustExtraData               `json:"extraData"`
	TitleCaptionTranslation *IllustTitleCaptionTranslation `json:"titleCaptionTranslation"`
	IsUnlisted              bool                           `json:"isUnlisted"`
	Request                 interface{}                    `json:"request"`
	CommentOff              int                            `json:"commentOff"`
	AiType                  int                            `json:"aiType"`
}

type UserIllusts struct {
	ID                      string                         `json:"id"`
	Title                   string                         `json:"title"`
	IllustType              int                            `json:"illustType"`
	XRestrict               int                            `json:"xRestrict"`
	Restrict                int                            `json:"restrict"`
	Sl                      int                            `json:"sl"`
	URL                     string                         `json:"url"`
	Description             string                         `json:"description"`
	Tags                    []string                       `json:"tags"`
	UserID                  string                         `json:"userId"`
	UserName                string                         `json:"userName"`
	Width                   int                            `json:"width"`
	Height                  int                            `json:"height"`
	PageCount               int                            `json:"pageCount"`
	IsBookmarkable          bool                           `json:"isBookmarkable"`
	BookmarkData            interface{}                    `json:"bookmarkData"`
	Alt                     string                         `json:"alt"`
	TitleCaptionTranslation *IllustTitleCaptionTranslation `json:"titleCaptionTranslation"`
	CreateDate              time.Time                      `json:"createDate"`
	UpdateDate              time.Time                      `json:"updateDate"`
	IsUnlisted              bool                           `json:"isUnlisted"`
	IsMasked                bool                           `json:"isMasked"`
	ProfileImageURL         string                         `json:"profileImageUrl"`
	AiType                  int                            `json:"aiType"`
}
