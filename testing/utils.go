package testing

import (
	"os"
	"path/filepath"
)

func LoadJson(fileName string) []byte {
	// Load the JSON file
	jsonPath := filepath.Join("testdata", fileName+".json")
	jsonData, _ := os.ReadFile(jsonPath)
	return jsonData
}

func LoadJsonString(fileName string) string {
	return string(LoadJson(fileName))
}
