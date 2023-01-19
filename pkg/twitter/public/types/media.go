package twitter_public_types

type EntityMediaSize struct {
	H      int    `json:"h"`
	W      int    `json:"w"`
	Resize string `json:"resize"`
}

type EntityMediaSizes struct {
	Large  EntityMediaSize `json:"large"`
	Medium EntityMediaSize `json:"medium"`
	Small  EntityMediaSize `json:"small"`
	Thumb  EntityMediaSize `json:"thumb"`
}

type EntityMediaFeature struct {
	Faces []*EntityMediaFace `json:"faces"`
}

type EntityMediaFeatures struct {
	Large  *EntityMediaFeature `json:"large,omitempty"`
	Medium *EntityMediaFeature `json:"medium,omitempty"`
	Small  *EntityMediaFeature `json:"small,omitempty"`
	Orig   *EntityMediaFeature `json:"orig,omitempty"`
}

type EntityMediaFace struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type EntityMediaOriginalInfo struct {
	Height     int               `json:"height"`
	Width      int               `json:"width"`
	FocusRects []EntityMediaFace `json:"focus_rects"`
}

type EntityMedia struct {
	DisplayURL    string                  `json:"display_url"`
	ExpandedURL   string                  `json:"expanded_url"`
	IDStr         string                  `json:"id_str"`
	Indices       []int                   `json:"indices"`
	MediaURLHTTPS string                  `json:"media_url_https"`
	Type          EntityMediaType         `json:"type"`
	URL           string                  `json:"url"`
	Features      EntityMediaFeatures     `json:"features"`
	Sizes         EntityMediaSizes        `json:"sizes"`
	OriginalInfo  EntityMediaOriginalInfo `json:"original_info"`
}

type EntityMediaType string

const (
	TweetLegacyExtendedEntityMediaTypePhoto       EntityMediaType = "photo"
	TweetLegacyExtendedEntityMediaTypeVideo       EntityMediaType = "video"
	TweetLegacyExtendedEntityMediaTypeAnimatedGIF EntityMediaType = "animated_gif"
)

type ExtendedEntityMedia struct {
	EntityMedia

	MediaKey      string `json:"media_key"`
	ExtMediaColor struct {
		Palette []struct {
			Percentage float64 `json:"percentage"`
			Rgb        struct {
				Blue  int `json:"blue"`
				Green int `json:"green"`
				Red   int `json:"red"`
			} `json:"rgb"`
		} `json:"palette"`
	} `json:"ext_media_color"`
	ExtMediaAvailability struct {
		Status string `json:"status"`
	} `json:"ext_media_availability"`
	VideoInfo *ExtendedEntityMediaVideoInfo `json:"video_info,omitempty"`
}

type ExtendedEntityMediaVideoInfo struct {
	AspectRatio []int                             `json:"aspect_ratio"`
	Variants    []ExtendedEntityMediaVideoVariant `json:"variants"`
}

type ExtendedEntityMediaVideoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}
