package bilibili

type LiveInfo struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    *LiveInfoData `json:"data"`
}

type LiveInfoData struct {
	Room     *RoomInfo     `json:"room_info"`
	Streamer *StreamerInfo `json:"anchor_info"`
}

type RoomInfo struct {
	UID                uint64     `json:"uid"`
	RoomID             uint64     `json:"room_id"`
	ShortID            uint64     `json:"short_id"`
	Title              string     `json:"title"`
	Cover              string     `json:"cover"`
	Background         string     `json:"background"`
	Description        string     `json:"description"`
	LiveStatus         LiveStatus `json:"live_status"`
	LiveStartTimestamp uint64     `json:"live_start_time"`
	Keyframe           string     `json:"keyframe"`
}

type LiveStatus int8

const (
	LiveStatusInactive          LiveStatus = 0
	LiveStatusActive            LiveStatus = 1
	LiveStatusStreamingArchives LiveStatus = 2
)

type StreamerInfo struct {
	Base *StreamerBaseInfo `json:"base_info"`
}

type StreamerBaseInfo struct {
	Name   string `json:"uname"`
	Avatar string `json:"face"`
	Gender string `json:"gender"`
}
