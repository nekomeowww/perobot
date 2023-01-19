package pixiv_public_types

type BaseResp[B any] struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    B      `json:"body"`
}
