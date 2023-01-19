package pixiv_public_types

type User struct {
	UserID        string          `json:"userId"`
	Name          string          `json:"name"`
	Image         string          `json:"image"`
	ImageBig      string          `json:"imageBig"`
	Premium       bool            `json:"premium"`
	IsFollowed    bool            `json:"isFollowed"`
	IsMypixiv     bool            `json:"isMypixiv"`
	IsBlocking    bool            `json:"isBlocking"`
	Background    *UserBackground `json:"background"`
	SketchLiveID  interface{}     `json:"sketchLiveId"`
	Partial       int             `json:"partial"`
	AcceptRequest bool            `json:"acceptRequest"`
	SketchLives   []interface{}   `json:"sketchLives"`
}

type UserBackground struct {
	Repeat    interface{} `json:"repeat"`
	Color     interface{} `json:"color"`
	URL       string      `json:"url"`
	IsPrivate bool        `json:"isPrivate"`
}
