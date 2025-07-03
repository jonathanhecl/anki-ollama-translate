# Anki Ollama Translate

A simple tool to translate Anki cards using Ollama.

## Important

- Only works with old format apkg files. Export your collection to apkg from Anki.

## Usage

```sh
anki-ollama-translate <apkg> [OPTIONS]
```

### Options

- `-check`
  - Check all fields before translation.
- `-field="<field_name>"`
  - Select field to translate.
- `-model="<model_name>"`
  - Select Ollama model to translate. Default: llama3.2
- `-to="<language>"`
  - Select language to translate to. Default: español neutro
- `-h, --help`
  - Show this help message.


