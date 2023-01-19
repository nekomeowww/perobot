package twitter_public_types

const (
	GuestBearerToken = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
)

type GuestActivateResp struct {
	GuestToken string `json:"guest_token"`
}
