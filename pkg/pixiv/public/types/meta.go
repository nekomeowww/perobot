package pixiv_public_types

import "time"

type Global struct {
	Token    string `json:"token"`
	Services struct {
		Booth    string `json:"booth"`
		Sketch   string `json:"sketch"`
		VroidHub string `json:"vroidHub"`
		Accounts string `json:"accounts"`
	} `json:"services"`
	OneSignalAppID     string `json:"oneSignalAppId"`
	PublicPath         string `json:"publicPath"`
	CommonResourcePath string `json:"commonResourcePath"`
	Development        bool   `json:"development"`
	UserData           struct {
		ID            string `json:"id"`
		PixivID       string `json:"pixivId"`
		Name          string `json:"name"`
		ProfileImg    string `json:"profileImg"`
		ProfileImgBig string `json:"profileImgBig"`
		Premium       bool   `json:"premium"`
		XRestrict     int    `json:"xRestrict"`
		Adult         bool   `json:"adult"`
		SafeMode      bool   `json:"safeMode"`
		IllustCreator bool   `json:"illustCreator"`
		NovelCreator  bool   `json:"novelCreator"`
	} `json:"userData"`
	AdsData  interface{} `json:"adsData"`
	MiscData struct {
		Consent struct {
			Gdpr bool `json:"gdpr"`
		} `json:"consent"`
		PolicyRevision bool `json:"policyRevision"`
		Grecaptcha     struct {
			RecaptchaEnterpriseScoreSiteKey string `json:"recaptchaEnterpriseScoreSiteKey"`
		} `json:"grecaptcha"`
		Info struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			CreateDate string `json:"createDate"`
		} `json:"info"`
		IsSmartphone bool `json:"isSmartphone"`
	} `json:"miscData"`
	Premium struct {
		PopularSearch bool `json:"popularSearch"`
		AdsHide       bool `json:"adsHide"`
	} `json:"premium"`
	Mute []interface{} `json:"mute"`
}

type Preload struct {
	Timestamp time.Time          `json:"timestamp"`
	Illust    map[string]*Illust `json:"illust"`
	User      map[string]*User   `json:"user"`
}
