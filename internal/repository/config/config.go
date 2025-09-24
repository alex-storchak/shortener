package config

const (
	DefaultFileStoragePath = "../../data/file_db.txt"
)

type FileStorage struct {
	Path string `env:"FILE_STORAGE_PATH"`
}

type Config struct {
	FileStorage FileStorage
}
