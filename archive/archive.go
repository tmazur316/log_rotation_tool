package archive

type Archive interface {
	SendFile(filepath string) error
	DeleteFolder(filepath string) error
}
