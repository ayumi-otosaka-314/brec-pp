package config

import "time"

type Root struct {
	Server   Server          `mapstructure:"server" validate:"required"`
	Services ServiceRegistry `mapstructure:"services" validate:"required"`
}

type Server struct {
	ListenAddress string        `mapstructure:"listenAddress" validate:"required"`
	Timeout       time.Duration `mapstructure:"timeout" validate:"required,gt=0"`
	Paths         HandlerPaths  `mapstructure:"paths" validate:"required"`
}

type HandlerPaths struct {
	RecordUpload string `mapstructure:"recordUpload" validate:"required"`
}

type ServiceRegistry struct {
	Default   ServiceEntry           `mapstructure:"default" validate:"required"`
	Streamers []StreamerServiceEntry `mapstructure:"streamers"`
}

type ServiceEntry struct {
	Discord Discord `mapstructure:"discord" validate:"required"`
	Storage Storage `mapstructure:"storage" validate:"required"`
}

type StreamerServiceEntry struct {
	RoomID       uint64 `mapstructure:"roomId" validate:"required,gt=0"`
	ServiceEntry `mapstructure:",squash" validate:"required"`
}

type Discord struct {
	WebhookURL string `mapstructure:"webhookUrl" validate:"required,url"`
}

type Storage struct {
	RootPath    string      `mapstructure:"rootPath" validate:"required,dir"`
	GoogleDrive GoogleDrive `mapstructure:"googleDrive" validate:"required"`
}

type GoogleDrive struct {
	Timeout          time.Duration `mapstructure:"timeout" validate:"required,gt=0"`
	CredentialPath   string        `mapstructure:"credentialPath" validate:"required,file"`
	ReservedCapacity uint64        `mapstructure:"reservedCapacity" validate:"required"`
	ParentFolderID   string        `mapstructure:"parentFolderId" validate:"required"`
}
