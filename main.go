package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	// "github.com/flimzy/anki"
)

const (
	version = "1.0.0"
)

func main() {
	fmt.Println("ANKI Ollama Translate v" + version)

	if err := unzipFile("test.apkg", "test"); err != nil {
		fmt.Println(err)
		return
	}

	// akpg, err := anki.ReadFile("test.apkg")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer akpg.Close()

	// collection, err := akpg.Collection()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(collection)
}

func unzipFile(src, dest string) error {
	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		if _, err := os.Stat(dest); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(dest, 0755); err != nil {
					return err
				}
			}
		}

		fileWriter, err := os.Create(dest + "/" + file.Name)
		if err != nil {
			return err
		}
		defer fileWriter.Close()

		_, err = io.Copy(fileWriter, fileReader)
		if err != nil {
			return err
		}
	}
	return nil
}
