package brec

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap/zapcore"
)

type Event struct {
	Type      EventType           `json:"EventType"`
	TimeStamp string              `json:"EventTimestamp"`
	ID        string              `json:"EventId"`
	Data      jsoniter.RawMessage `json:"EventData"`
}

func (e *Event) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("EventType", string(e.Type))
	encoder.AddString("EventTimestamp", e.TimeStamp)
	encoder.AddString("EventId", e.ID)
	encoder.AddByteString("EventData", e.Data)
	return nil
}

func (e *Event) GetTimestamp() (time.Time, error) {
	return time.Parse(TimestampLayout, e.TimeStamp)
}

type EventDataSession struct {
	SessionID string `json:"SessionId"`
	EventDataBase
}

type EventDataFileOpen struct {
	RelativePath string `json:"RelativePath"`
	FileOpenTime string `json:"FileOpenTime"`
	SessionID    string `json:"SessionId"`
	EventDataBase
}

type EventDataFileClose struct {
	RelativePath  string  `json:"RelativePath"`
	FileSize      uint64  `json:"FileSize"`
	Duration      float64 `json:"Duration"`
	FileOpenTime  string  `json:"FileOpenTime"`
	FileCloseTime string  `json:"FileCloseTime"`
	SessionID     string  `json:"SessionId"`
	EventDataBase
}

type EventDataBase struct {
	RoomID           uint64 `json:"RoomId"`
	ShortID          uint64 `json:"ShortId"`
	StreamerName     string `json:"Name"`
	Title            string `json:"Title"`
	AreaNameParent   string `json:"AreaNameParent"`
	AreaNameChild    string `json:"AreaNameChild"`
	Recording        bool   `json:"Recording"`
	Streaming        bool   `json:"Streaming"`
	DanmakuConnected bool   `json:"DanmakuConnected"`
}

type EventType string

const (
	EventTypeSessionStarted EventType = "SessionStarted"
	EventTypeFileOpening    EventType = "FileOpening"
	EventTypeFileClosed     EventType = "FileClosed"
	EventTypeSessionEnded   EventType = "SessionEnded"
	EventTypeStreamStarted  EventType = "StreamStarted"
	EventTypeStreamEnded    EventType = "StreamEnded"

	TimestampLayout = "2006-01-02T15:04:05.9999999Z07:00"
)
