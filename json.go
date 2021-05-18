package main

type event struct {
	Time   int64  `json:"time"`
	SelfId int64  `json:"self_id"`
	Type   string `json:"post_type"`
}

type message struct {
	GroupID int64  `json:"group_id"`
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
	Sender  sender `json:"sender"`
	MsgType string `json:"message_type"`
}

type sender struct {
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
}
