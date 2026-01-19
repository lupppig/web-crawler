package parser

import (
	"bytes"
	"io"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type PageData struct {
	Title string
	Body  string
	Links []string
}

var skipTags = map[string]bool{
	"script":     true,
	"style":      true,
	"noscript":   true,
	"template":   true,
	"iframe":     true,
	"canvas":     true,
	"svg":        true,
	"meta":       true,
	"link":       true,
	"head":       true,
	"object":     true,
	"embed":      true,
	"javascript": true,
	"nav":        true,
	"footer":     true,
	"form":       true,
	"span":       true,
	"img":        true,
}

func Parse(content []byte, baseURL string) (*PageData, error) {
	reader := bytes.NewReader(content)
	tokenizer := html.NewTokenizer(reader)

	page := &PageData{
		Links: make([]string, 0),
	}

	var (
		inBody  bool
		inTitle bool
		words   []string
	)

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			}
			return nil, tokenizer.Err()
		}

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if _, ok := skipTags[token.Data]; ok {
				continue
			}
			if token.Data == "title" {
				inTitle = true
			}
			if token.Data == "body" {
				inBody = true
			}
			if token.Data == "a" {
				href := getHref(token, baseURL)
				if href != "" {
					page.Links = append(page.Links, href)
				}
			}

		case html.EndTagToken:
			token := tokenizer.Token()
			if token.Data == "title" {
				inTitle = false
			}
			if token.Data == "body" {
				inBody = false
			}

		case html.TextToken:
			if inTitle && page.Title == "" {
				page.Title = strings.TrimSpace(tokenizer.Token().Data)
			}
			if inBody && len(words) < 1000 {
				text := strings.TrimSpace(tokenizer.Token().Data)
				tokens := tokenize(text)
				remaining := 1000 - len(words)
				if len(tokens) > remaining {
					words = append(words, tokens[:remaining]...)
				} else {
					words = append(words, tokens...)
				}
			}
		}
	}
	page.Body = strings.Join(words, " ")
	return page, nil
}

func tokenize(text string) []string {
	var body strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) {
			body.WriteRune(r)
		} else {
			body.WriteByte(' ')
		}
	}
	return strings.Fields(body.String())
}

func getHref(tt html.Token, base string) string {
	for _, attr := range tt.Attr {
		if attr.Key != "href" {
			continue
		}
		val := strings.TrimSpace(attr.Val)
		if isSameDomain(base, val) {
			return val
		}
	}
	return ""
}

func isSameDomain(base, link string) bool {
	baseURL, err1 := url.Parse(base)
	linkURL, err2 := url.Parse(link)

	if err1 != nil || err2 != nil {
		return false
	}

	if !linkURL.IsAbs() {
		return true
	}

	host := strings.TrimPrefix(linkURL.Hostname(), "www.")
	baseHost := strings.TrimPrefix(baseURL.Hostname(), "www.")
	return baseHost == host
}
