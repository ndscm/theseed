package seedspa

import (
	"net/http"
	"strings"
)

func containsExtension(filePath string) bool {
	parts := strings.Split(filePath, "/")
	fileName := parts[len(parts)-1]
	return strings.Contains(fileName, ".")
}

type spaFileSystem struct {
	http.FileSystem
	spaFallback string
}

func (cls spaFileSystem) Open(name string) (http.File, error) {
	file, err := cls.FileSystem.Open(name)
	if err == nil {
		return file, nil
	}
	if !containsExtension(name) {
		spaFile, err := cls.FileSystem.Open(cls.spaFallback)
		if err == nil {
			return spaFile, nil
		}
		return nil, err
	}
	return nil, err
}

func SpaServer(root http.FileSystem, spaFallback string) http.Handler {
	spafs := spaFileSystem{
		FileSystem:  root,
		spaFallback: spaFallback,
	}
	return http.FileServer(spafs)
}
