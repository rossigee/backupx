package destinations

import "io"

type IBackupDestinationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

type IBackupDestination interface {
	GetWriter() io.Writer
}
