package parser

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

var backgroundImageRe = regexp.MustCompile(`background-image:url\('(.*)'\)`)

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
	if selectionErr != nil {
		return ret, selectionErr
	}

	doc.Find("span.tgme_widget_message_meta").Find("time.datetime").Each(func(i int, s *goquery.Selection) {
		date, err := time.Parse("2006-01-02T15:04:05-07:00", s.AttrOr("datetime", ""))
		if err != nil {
			selectionErr = err
			log.Println("[ERROR] Failed to parse date", err)
			return
		}
		ret.Date = date
	})
	if selectionErr != nil {
		return ret, selectionErr
	}

	doc.Find("a.tgme_widget_message_photo_wrap").Each(func(i int, s *goquery.Selection) {
		style, exists := s.Attr("style")
		if !exists {
			selectionErr = errors.New("style attribute not found")
			log.Println("[ERROR] Failed to get style attribute", err)
			return
		}

		groups := backgroundImageRe.FindStringSubmatch(style)
		if len(groups) != 2 {
			selectionErr = fmt.Errorf("expected 2 match, got %d", len(groups))
			log.Println("[ERROR] Failed to find image URL", err)
			return
		}

		_, err := url.Parse(groups[1])
		if err != nil {
			selectionErr = err
			log.Println("[ERROR] Invalid parsed image URL: ", groups[1], err)
			return
		}

		ret.ImagesLink = append(ret.ImagesLink, groups[1])
	})

	return ret, selectionErr
}
