package pixiv_public_types

type IllustDetailPagesRespItem struct {
	Urls struct {
		ThumbMini string `json:"thumb_mini"`
		Small     string `json:"small"`
		Regular   string `json:"regular"`
		Original  string `json:"original"`
	} `json:"urls"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type IllustDetailPagesResp = BaseResp[[]*IllustDetailPagesRespItem]

type IllustDetailResp = BaseResp[*Illust]
