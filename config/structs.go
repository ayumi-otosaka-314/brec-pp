package config

import "time"

type Root struct {
	Server  Server
	Discord Discord
	Storage Storage
}

type Server struct {
	ListenAddress string
	Timeout       time.Duration
	Paths         HandlerPaths
}

type HandlerPaths struct {
	RecordUpload string
}

type Discord struct {
	WebhookURL string
}

type Storage struct {
	RootPath    string
	GoogleDrive GoogleDrive
}

type GoogleDrive struct {
	Timeout          time.Duration
	CredentialPath   string
	ReservedCapacity uint64
	ParentFolderID   string
}
