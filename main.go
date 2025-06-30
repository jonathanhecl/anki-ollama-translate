package main

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

var (
	origApkg      string
	tempDB        string
	fieldSelected int8 = -1

	//
	version string = "1.0.0"
)

// const (
// 	origApkg         = "Jlab's beginner course.apkg"
// 	tempDB           = "Jlab's beginner course.anki2"
// 	exportedEN       = "output_en.txt"
// 	translatedES     = "output_es.txt"
// 	newApkgOutput    = "deck_traducido.apkg"
// 	fieldToTranslate = "RemarksBack"
// )

func printUsage() {
	fmt.Println("Usage: anki-ollama-translate <apkg>")
	os.Exit(1)
}

func main() {

	fmt.Println("anki-ollama-translate v" + version)
	if len(os.Args) < 2 {
		printUsage()
	}

	var check bool = false

	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, "-h") || strings.EqualFold(arg, "--help") {
			printUsage()
		} else if strings.EqualFold(arg, "-check") {
			check = true
		} else if strings.HasPrefix(arg, "-") {
			fmt.Println("❌ Parámetro inválido:", arg)
			printUsage()
		} else {
			origApkg = normalizeFileName(arg, ".apkg")
		}
	}

	if !fileExists(origApkg) {
		fmt.Println("❌ No se encontró el archivo APKG:", origApkg)
		os.Exit(1)
	}

	fmt.Println("✅ Archivo APKG:", origApkg)

	return

	// // Paso 1: descomprimir el archivo APKG y extraer collection.anki2
	// if err := unzipCollection(origApkg, tempDB); err != nil {
	// 	panic(err)
	// }

	// // Paso 2: abrir la base de datos
	// db, err := sql.Open("sqlite", tempDB)
	// if err != nil {
	// 	fmt.Println("❌ Error abriendo la base SQLite:", err)
	// 	os.Exit(1)
	// }
	// if err = db.Ping(); err != nil {
	// 	fmt.Println("❌ No se puede acceder a la base SQLite:", err)
	// 	os.Exit(1)
	// }
	// defer db.Close()

	// var raw string
	// err = db.QueryRow("SELECT models FROM col").Scan(&raw)
	// if err != nil {
	// 	panic(err)
	// }

	// var models map[string]interface{}
	// err = json.Unmarshal([]byte(raw), &models)
	// if err != nil {
	// 	panic(err)
	// }

	// listFields := []string{}

	// for _, modelData := range models {
	// 	model := modelData.(map[string]interface{})
	// 	fmt.Println(model["name"])
	// 	fields := model["flds"].([]interface{})
	// 	for i, f := range fields {
	// 		name := f.(map[string]interface{})["name"].(string)
	// 		listFields = append(listFields, fmt.Sprintf("[%s] %d = %s", model["name"].(string), i, name))
	// 		if name == fieldToTranslate {
	// 			fieldSelected = int8(i)
	// 			break
	// 		}
	// 	}
	// }

	// if fieldSelected == -1 {
	// 	fmt.Println("❌ No se encontró el campo a traducir:", fieldToTranslate)
	// 	fmt.Println("Los campos disponibles son:")
	// 	for _, v := range listFields {
	// 		fmt.Println(v)
	// 	}
	// 	os.Exit(1)
	// }

	// // Paso 3A: si no existe traducción, extraer los reversos
	// if _, err := os.Stat(translatedES); os.IsNotExist(err) {
	// 	extractReverses(db, exportedEN)
	// 	fmt.Println("✔ Archivo creado:", exportedEN)
	// 	fmt.Println("→ Ahora traducilo línea por línea y guardalo como:", translatedES)
	// 	return
	// }

	// // Paso 3B: si existe traducción, modificar los reversos
	// if err := applyTranslations(db, translatedES); err != nil {
	// 	panic(err)
	// }

	// // Paso 4: reempacar como nuevo APKG
	// if err := repackApkg(tempDB, newApkgOutput); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("✔ Nuevo APKG generado:", newApkgOutput)
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

func extractReverses(db *sql.DB, outFile string) {
	rows, err := db.Query("SELECT flds FROM notes ORDER BY id")
	if err != nil {
		fmt.Println("❌ Error al hacer SELECT flds FROM notes:", err)
		os.Exit(1)
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var flds string
		if err := rows.Scan(&flds); err != nil {
			fmt.Println("❌ Error al escanear una fila:", err)
			continue
		}
		fields := strings.Split(flds, "\x1f")
		if len(fields) > 1 {
			if int(fieldSelected) < len(fields) {
				lines = append(lines, fields[fieldSelected])
			}
		} else {
			lines = append(lines, "")
		}
	}

	err = os.WriteFile(outFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Println("❌ Error al guardar archivo:", err)
		os.Exit(1)
	}
}

func applyTranslations(db *sql.DB, transFile string) error {
	lines, err := readLines(transFile)
	if err != nil {
		return err
	}

	rows, _ := db.Query("SELECT id, flds FROM notes ORDER BY id")
	defer rows.Close()

	tx, _ := db.Begin()
	idx := 0
	for rows.Next() {
		var id int64
		var flds string
		rows.Scan(&id, &flds)
		fields := strings.Split(flds, "\x1f")
		if len(fields) > 1 && idx < len(lines) {
			fields[fieldSelected] = lines[idx]
		}
		newFlds := strings.Join(fields, "\x1f")
		tx.Exec("UPDATE notes SET flds = ? WHERE id = ?", newFlds, id)
		idx++
	}
	tx.Commit()
	fmt.Printf("✔ Aplicadas %d traducciones.\n", idx)
	return nil
}

func readLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
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
