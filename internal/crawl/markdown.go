package crawl

import (
	htm "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ToMarkdown(htmlString string) (string, error) {
	markdown, err := htm.ConvertString(htmlString)
	if err != nil {
		return "", err
	}

	return markdown, nil
}
