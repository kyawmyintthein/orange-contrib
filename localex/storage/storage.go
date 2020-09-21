package storage

type Storage interface {
	GetLocalizedMessage(string, string) string
}
