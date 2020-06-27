package sources

import "io"

type IBackupSourceConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

type IBackupSource interface {
	GetReader() io.Reader
}
