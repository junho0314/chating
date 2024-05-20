package model

type ChatingRoom struct {
	RoomId   string `json:"roomId"`
	RoomName string `json:"roomName"`
	IsUsed   bool   `json:"isUsed"`
}
