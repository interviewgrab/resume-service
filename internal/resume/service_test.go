package resume

import (
	"os"
	"strings"
	"testing"
)

func TestParsePDF(t *testing.T) {
	pdf := getFile(t)
	parsedText, err := parsePDF(pdf)
	if err != nil {
		t.Errorf("Error parsing pdf")
	}
	t.Log(parsedText)
}

func getFile(t *testing.T) []byte {
	files, err := listFileNames(".")
	if err != nil || len(files) == 0 {
		t.Errorf("Cannot find any pdf")
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err == nil {
			return body
		}
	}
	t.Errorf("Error reading pdf files")
	return nil
}

func listFileNames(folderPath string) ([]string, error) {
	var fileNames []string

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return fileNames, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".pdf") {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}
