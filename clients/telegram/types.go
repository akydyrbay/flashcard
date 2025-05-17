package telegram

type Update struct {
	ID      int    `json:"update_id"`
	Message string `json:"message"`
}

type UpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}
