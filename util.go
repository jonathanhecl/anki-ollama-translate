package main

import (
	"os"
	"path/filepath"
)

func fileExists(pathName string) bool {
	_, err := os.Stat(pathName)
	return !os.IsNotExist(err)
}

func normalizeFileName(pathName string, newExt string) string {
	if newExt[0] != '.' {
		newExt = "." + newExt
	}
	ext := filepath.Ext(pathName)
	if ext != newExt {
		if len(ext) == 0 {
			return pathName + newExt
		}
		return pathName[:len(pathName)-len(ext)] + newExt
	}
	return pathName
}
