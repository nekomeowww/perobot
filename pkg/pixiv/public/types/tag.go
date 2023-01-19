package pixiv_public_types

type Tag struct {
	Tag         string `json:"tag"`
	Locked      bool   `json:"locked"`
	Deletable   bool   `json:"deletable"`
	UserID      string `json:"userId"`
	Translation struct {
		En string `json:"en"`
	} `json:"translation,omitempty"`
	UserName string `json:"userName"`
}

type Tags struct {
	AuthorID string `json:"authorId"`
	IsLocked bool   `json:"isLocked"`
	Tags     []*Tag `json:"tags"`
	Writable bool   `json:"writable"`
}
