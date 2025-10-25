package crawl

import (
	toMarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ToMarkdown(htmlString string) (string, error) {
	mk, err := toMarkdown.ConvertString(htmlString)
	if err != nil {
		return "", err
	}

	return mk, nil
}
