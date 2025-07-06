package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jonathanhecl/gollama"
	"github.com/jonathanhecl/gotimeleft"
	_ "modernc.org/sqlite"
)

var (
	origApkg        string
	newApkgOutput   string
	tempDB          string
	fieldSelected   string = ""
	fieldSelectedID int8   = -1
	sequenceID      int64  = -1
	modelSelected   string = "llama3.2"
	version         string = "1.0.4"
	toLanguage      string = "español neutro"
	askTranslation  bool   = false
)

func printUsage() {
	fmt.Println("Usage: anki-ollama-translate <apkg> [OPTIONS]")
	fmt.Println("Options:")
	fmt.Println("  -check \tCheck all fields before translation.")
	fmt.Println("  -field=\"<field_name>\" \tSelect field to translate.")
	fmt.Println("  -model=\"<model_name>\" \tSelect Ollama model to translate. Default: llama3.2")
	fmt.Println("  -to=\"<language>\" \tSelect language to translate to. Default: español neutro")
	fmt.Println("  -ask \tAsk for manual translation when it's not complete.")
	fmt.Println("  -h, --help \tShow this help message.")
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
		} else if strings.HasPrefix(arg, "-field=") {
			fieldSelected = arg[len("-field="):]
		} else if strings.HasPrefix(arg, "-model=") {
			modelSelected = arg[len("-model="):]
		} else if strings.HasPrefix(arg, "-to=") {
			toLanguage = arg[len("-to="):]
		} else if strings.HasPrefix(arg, "-ask") {
			askTranslation = true
		} else if strings.HasPrefix(arg, "-") {
			fmt.Println("❌ Invalid parameter:", arg)
			printUsage()
		} else {
			origApkg = normalizeFileName(arg, ".apkg")
		}
	}

	if !fileExists(origApkg) {
		fmt.Println("❌ APKG not found:", origApkg)
		return
	}

	newApkgOutput = normalizeFileName(origApkg, "_"+toLanguage+".apkg")

	// Ollama
	ctx := context.Background()
	g := gollama.New(modelSelected)
	if found, _ := g.HasModel(ctx, modelSelected); !found {
		fmt.Println("❌ Ollama model not found:", modelSelected)
		if err := g.PullIfMissing(ctx); err != nil {
			fmt.Println("❌ Error pulling Ollama model:", err)
			return
		}
		fmt.Println("✅ Ollama model downloaded:", modelSelected)
	} else {
		fmt.Println("✅ Ollama model found:", modelSelected)
	}

	// Requirements
	if !check && fieldSelected == "" {
		fmt.Println("❌ Invalid parameters: -check or -field are required.")
		printUsage()
	}

	fmt.Println("✅ APKG found:", origApkg)

	tempDB = normalizeFileName(origApkg, "_temp.anki2")
	if err := unzipCollection(origApkg, tempDB); err != nil {
		fmt.Println("❌ Error unzipping APKG:", err)
		return
	}

	db, err := sql.Open("sqlite", tempDB)
	if err != nil {
		fmt.Println("❌ Error opening SQLite database:", err)
		return
	}
	if err = db.Ping(); err != nil {
		fmt.Println("❌ Error pinging SQLite database:", err)
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			fmt.Println("❌ Error closing SQLite database:", err)
		}

		if err := os.Remove(tempDB); err != nil {
			fmt.Println("❌ Error removing temporary SQLite database:", err)
		}
	}()

	if check {
		checkFields(db)
		return
	}

	if fieldSelected == "" {
		fmt.Println("❌ Field not selected. Use -field=\"<field_name>\" to select a field. Use -check to check all fields if you are not sure.")
		return
	}

	fieldSelectedID = findFieldID(db, fieldSelected)
	if fieldSelectedID == -1 {
		fmt.Println("❌ Field not found:", fieldSelected)
		return
	}

	fmt.Println("✅ Field found:", fieldSelected, "[", fieldSelectedID, "]")

	lines := extractLines(db, fieldSelectedID)

	progress := gotimeleft.Init(len(lines))
	progress.Value(0)

	for i, line := range lines {
		if progress.GetValue()%25 == 0 {
			fmt.Printf("\nTranslation progress: %s %s lines (%s) - Total time: %s - Time left: %s\n", progress.GetProgressBar(50), progress.GetProgressValues(), progress.GetProgress(0), progress.GetTimeSpent().String(), progress.GetTimeLeft().String())
		}

		progress.Step(1)
		lines[i] = translateLine(g, i, line, "")
	}

	fmt.Printf("\nTranslation completed.\n")

	if err := applyTranslations(db, lines); err != nil {
		fmt.Println("❌ Error applying translations:", err)
		return
	}

	if err := repackApkg(tempDB, newApkgOutput); err != nil {
		fmt.Println("❌ Error repacking APKG:", err)
		return
	}
	fmt.Println("✔ New APKG generated:", newApkgOutput)
}

func findFieldID(db *sql.DB, fieldName string) int8 {
	var raw string
	err := db.QueryRow("SELECT models FROM col").Scan(&raw)
	if err != nil {
		panic(err)
	}

	var models map[string]interface{}
	err = json.Unmarshal([]byte(raw), &models)
	if err != nil {
		panic(err)
	}

	for _, modelData := range models {
		model := modelData.(map[string]interface{})
		fields := model["flds"].([]interface{})
		for i, f := range fields {
			name := f.(map[string]interface{})["name"].(string)
			if strings.EqualFold(name, "Sequence") {
				sequenceID = int64(i)
			}
			if strings.EqualFold(name, fieldName) {
				return int8(i)
			}
		}
	}

	return -1
}

func checkFields(db *sql.DB) {
	var raw string
	err := db.QueryRow("SELECT models FROM col").Scan(&raw)
	if err != nil {
		panic(err)
	}

	var models map[string]interface{}
	err = json.Unmarshal([]byte(raw), &models)
	if err != nil {
		panic(err)
	}

	fieldName := map[int]string{}

	for _, modelData := range models {
		model := modelData.(map[string]interface{})
		fields := model["flds"].([]interface{})
		for i, f := range fields {
			name := f.(map[string]interface{})["name"].(string)
			if len(fieldName[i]) > 0 {
				continue
			}
			fieldName[i] = name
		}
	}

	rows, err := db.Query("SELECT flds FROM notes ORDER BY id")
	if err != nil {
		fmt.Println("❌ Error on SELECT flds FROM notes:", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var flds string
		if err := rows.Scan(&flds); err != nil {
			fmt.Println("❌ Error scanning row:", err)
			continue
		}
		fields := strings.Split(flds, "\x1f")
		for i, f := range fields {
			if len(fieldName[i]) > 0 {
				fmt.Println(fieldName[i], "[", i, "]", f)
			}
		}
	}

	fmt.Println("✅ All fields checked.")
}

func extractLines(db *sql.DB, fieldSelectedID int8) map[int64]string {
	rows, err := db.Query("SELECT flds FROM notes ORDER BY id")
	if err != nil {
		fmt.Println("❌ Error on SELECT flds FROM notes:", err)
		os.Exit(1)
	}
	defer rows.Close()

	lines := map[int64]string{}
	for rows.Next() {
		var flds string
		if err := rows.Scan(&flds); err != nil {
			fmt.Println("❌ Error scanning row:", err)
			continue
		}
		fields := strings.Split(flds, "\x1f")
		if len(fields) > 1 {
			if int(fieldSelectedID) < len(fields) {
				id, _ := strconv.ParseInt(fields[1], 10, 64)
				lines[id] = fields[fieldSelectedID]
			}
		} else {
			id, _ := strconv.ParseInt(fields[1], 10, 64)
			lines[id] = ""
		}
	}
	return lines
}

func translateLine(g *gollama.Gollama, id int64, originalLine, translatedLine string) string {
	if len(originalLine) < 2 { // Skip short lines
		return originalLine
	}

	translateCtx := context.Background()
	prompt := `You are a translator. You are translating a Anki card.
	* Don't remove any tag, <>, [], :, ->, everything.
	* Don't remove any example of other language.
	* Don't convert any HTML tag to markdown or any other format.
	* Please be as accurate as possible.
	* Return a JSON object with a "Translation" key, in one line.`
	if len(translatedLine) > 0 {
		prompt += `* The original text has ` + strconv.Itoa(len(originalLine)) + ` characters.
		* The result of the previous translation has ` + strconv.Itoa(len(translatedLine)) + ` characters.
		* We believe you missed some translation.
		* Include all the characters you think are missing in the translation.
		* You previous translation was incomplete, try again.
		Previous translation: 
		
		` + translatedLine + `
	`
	}

	prompt += `
Translate the following text to ` + toLanguage + `:

` + originalLine + `
`

	type outputType struct {
		Translation string `description:"Translation"`
	}

	response, err := g.Chat(translateCtx, prompt, gollama.StructToStructuredFormat(outputType{}))
	if err != nil {
		log.Fatal("❌ Error getting translation from Gollama:", err)
		return ""
	}

	var output outputType
	if err := response.DecodeContent(&output); err != nil {
		log.Fatal("❌ Error decoding response:", err)
		return ""
	}

	// fmt.Println("✅ Original:", originalLine)
	if len(output.Translation) < (len(originalLine) / 2) {
		// fmt.Println("❌ Translation too short:", output.Translation)
		if translatedLine == output.Translation {
			if askTranslation {
				userTranslation := getUserTranslation(id, originalLine)
				if len(userTranslation) > 0 {
					return userTranslation
				}
			}
			fmt.Println("❌ Not translated [", id, "]:", originalLine)
			return originalLine // Avoid infinite loop
		}
		return translateLine(g, id, originalLine, output.Translation)
	}
	// fmt.Println("✅ Translation:", output.Translation)

	return output.Translation
}

func applyTranslations(db *sql.DB, lines map[int64]string) error {

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
			if len(fields) > int(fieldSelectedID) {
				fields[fieldSelectedID] = lines[id]
				idx++
			}
		}
		newFlds := strings.Join(fields, "\x1f")
		tx.Exec("UPDATE notes SET flds = ? WHERE id = ?", newFlds, id)

	}

	tx.Commit()
	fmt.Printf("✔ Applied %d translations.\n", idx)
	return nil
}

func getUserTranslation(id int64, originalLine string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Can't translate this line. Please translate it manually.")
		fmt.Println("👁️ Original [", id, "]:", originalLine)
		fmt.Print("✏️ Input your translation: ")
		translation, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\n⚠️ Error reading input:", err)
			continue
		}
		translation = strings.TrimSpace(translation)
		if translation == "" {
			fmt.Println("⚠️ Translation cannot be empty. Please try again.")
			continue
		}
		fmt.Println("✅ Translation:", translation)

		for {
			fmt.Print("Accept translation? (y/n): ")
			accept, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("\n⚠️ Error reading input:", err)
				continue
			}
			accept = strings.TrimSpace(strings.ToLower(accept))
			if accept == "y" {
				return translation
			} else if accept == "n" {
				break
			} else {
				fmt.Println("⚠️ Please enter 'y' or 'n'")
			}
		}
	}
}
