package config

const (
	DefaultFileStoragePath = "../../data/file_db.txt"
)

type Config struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}
