package resume

import (
	"bytes"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
	"strings"
)

func parsePDF(fileContent []byte) (string, error) {
	r := bytes.NewReader(fileContent)
	pdfReader, err := model.NewPdfReader(r)
	if err != nil {
		return "", err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}
	var textBuilder strings.Builder

	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return "", err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}

		pageText, err := ex.ExtractText()
		if err != nil {
			return "", err
		}

		textBuilder.WriteString(pageText)
	}

	return textBuilder.String(), nil
}
