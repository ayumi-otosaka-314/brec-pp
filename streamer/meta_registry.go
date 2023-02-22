package streamer

type MetaRegistry interface {
	GetAvatar(streamerUID uint64, sessionID string) (string, error)
	GetThumbnail(roomID uint64, sessionID string) (string, error)
}
