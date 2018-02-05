package chat

// Message receive/send to websocket client
//
type Message struct {
	//ID      uint64 `json:"id"`
	//Type    int    `json:"type"`
	//From    string `json:"from"`
	//To      string `json:"to"`
	ID        uint64 `json:"-"`
	Timestamp int64  `json:"timestamp"`
	Content   string `json:"content"`
	From      string `json:"from"`
}
