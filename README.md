# Anki Ollama Translate

A simple tool to translate Anki cards using Ollama (https://ollama.com/).

You can use it to translate your Anki decks to another language using AI models.

## ⚠️ Important for exports

- This tool only works with Anki package files (.apkg) exported in the legacy format.
- **How to export your deck correctly in Anki**:
  1. Go to `Decks` > `Export...`
  2. Select your deck from the dropdown
  3. Set format to "Anki Deck Package (*.apkg)"
  4. **IMPORTANT**: Check the option "Compatibility with older Anki versions (slower/larger files)"
  5. Make sure "Include media" is checked if your cards have images/audio
  6. Click "Export" and save the file

## ⚠️ Important for import
  1. Open Anki
  2. Go to `File` > `Import...`
  3. Select your deck from the dropdown
  4. Set format to "Anki Deck Package (*.apkg)"
  5. Make sure "Include media" is checked if your cards have images/audio
  6. Click "Import" and save the file

If you have any issues with the import, try to do Tools -> Check Database first.

## 🛠️ Usage

```sh
anki-ollama-translate <apkg> [OPTIONS]
```

### Options

- `-check`
  - Check all fields before translation. This is useful to see the fields names. (Can be combined with -type and -field)
- `-type="<type_name>"`
  - Select type to translate. Default: all types (Can be combined with -check)
- `-field="<field_name>"`
  - Select field to translate. (Can be combined with -check)
- `-model="<model_name>"`
  - Select Ollama model to translate. Default: llama3.2
- `-from="<language>"`
  - Select language to translate from. Default: auto-detect
- `-to="<language>"`
  - Select language to translate to. Default: español neutro
- `-ask`
  - To ask for a manual translation when the AI translation cannot do it.
- `-v`
  - Enable verbose mode. This can make the process slower.
- `-h, --help`
  - Show this help message.

## Understanding the check results

When you use the `-check` option, the tool will show you the type and fields names and their content. This is useful to see the fields names.
Ej. `Type / Field [ Order ] Content`

## Requirements

- Ollama (https://ollama.com/) installed and running.
- OPTIONAL: Llama3.2 model installed in Ollama.
- Windows, Linux or macOS.

## Recommended models

- phi4 (less errors, but slower)
- llama3.2 (faster, but more errors)

## Download

You can download the latest release from [Releases](https://github.com/jonathanhecl/anki-ollama-translate/releases).

## 📝 License

MIT License
