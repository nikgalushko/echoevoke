package parser

import (
	"io"
	"log"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

type PostInfo struct {
	Content    string
	Author     string
	Date       time.Time
	ImagesLink []string
}

func Parse(data io.Reader) (PostInfo, error) {
	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return PostInfo{}, err
	}

	var selectionErr error
	ret := PostInfo{}
	doc.Find("div.tgme_widget_message_text[dir='auto']").Each(func(i int, s *goquery.Selection) {
		converter := md.NewConverter("", true, nil)
		html, err := s.Html()
		if err != nil {
			selectionErr = err
			log.Println("[ERROR] Failed to get HTML from selection", err)
			return
		}

		markdown, err := converter.ConvertString(html)
		if err != nil {
			selectionErr = err
			log.Println("[ERROR] Failed to convert HTML to markdown", err)
			return
		}

		ret.Content = markdown
	})

	return ret, selectionErr
}
