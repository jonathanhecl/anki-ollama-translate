package main

import (
	"archive/zip"
	"fmt"
	"io"
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

func unzipCollection(apkgPath, outDBPath string) error {
	r, err := zip.OpenReader(apkgPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "collection.anki2" || f.Name == "collection.anki21" {
			rc, _ := f.Open()
			defer rc.Close()
			out, err := os.Create(outDBPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("no se encontró collection.anki2 en el APKG")
}

func repackApkg(dbPath, outZip string) error {
	newFile, err := os.Create(outZip)
	if err != nil {
		return err
	}
	defer newFile.Close()

	w := zip.NewWriter(newFile)

	// Añadir base de datos modificada
	dbBytes, _ := os.ReadFile(dbPath)
	f, _ := w.Create("collection.anki2")
	f.Write(dbBytes)

	// Añadir media (mínimo válido: archivo vacío)
	f2, _ := w.Create("media")
	f2.Write([]byte("{}"))

	return w.Close()
}
