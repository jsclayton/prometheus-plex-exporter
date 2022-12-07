package plex

type Library struct {
	Name string
	ID   string
	Type string

	Server *Server

	DurationTotal int64
	StorageTotal  int64
}

func isLibraryDirectoryType(directoryType string) bool {
	switch directoryType {
	case
		"movie",
		"show",
		"artist":
		return true
	}
	return false
}
