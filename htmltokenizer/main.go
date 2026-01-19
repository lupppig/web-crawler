package main

import (
	"bytes"

	"golang.org/x/net/html"
)

func extractBodyText(htmlContent []byte) string {
	tokenizer := html.NewTokenizer(bytes.NewReader(htmlContent))

	inBody := false
	var result string

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return result

		case html.StartTagToken, html.SelfClosingTagToken:
			t := tokenizer.Token()
			if t.Data == "body" {
				inBody = true
			}

		case html.EndTagToken:
			t := tokenizer.Token()
			if t.Data == "body" {
				inBody = false
			}

		case html.TextToken:
			if inBody {
				text := tokenizer.Token().Data
				result += text + " "
			}
		}
	}
}
