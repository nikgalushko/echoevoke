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
	Date       time.Time
	ImagesLink []string
}

// Parse returns the content, date and images of a post
func Parse(data io.Reader) (PostInfo, error) {
	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return PostInfo{}, err
	}

	content, err := selectContent(doc)
	if err != nil {
		return PostInfo{}, err
	}

	date, err := selectDate(doc)
	if err != nil {
		return PostInfo{}, err
	}

	images, err := selectImages(doc)
	if err != nil {
		return PostInfo{}, err
	}

	return PostInfo{Content: content, Date: date, ImagesLink: images}, nil
}

// selectContent returns the content of the post as markdown
func selectContent(doc *goquery.Document) (markdown string, err error) {
	doc.Find("div.tgme_widget_message_text[dir='auto']").Each(func(i int, s *goquery.Selection) {
		var html string

		converter := md.NewConverter("", true, nil)
		html, err = s.Html()
		if err != nil {
			log.Println("[ERROR] failed to get HTML from selection", err)
			return
		}

		markdown, err = converter.ConvertString(html)
		if err != nil {
			log.Println("[ERROR] failed to convert HTML to markdown", err)
		}
	})
	return
}

// selectDate returns the date of the post
func selectDate(doc *goquery.Document) (date time.Time, err error) {
	doc.Find("span.tgme_widget_message_meta").Find("time.datetime").Each(func(i int, s *goquery.Selection) {
		date, err = time.Parse("2006-01-02T15:04:05-07:00", s.AttrOr("datetime", ""))
		if err != nil {
			log.Println("[ERROR] failed to parse date", err)
		}
	})
	return
}

// selectImages returns a slice of image URLs found in the post
func selectImages(doc *goquery.Document) (images []string, err error) {
	doc.Find("a.tgme_widget_message_photo_wrap").Each(func(i int, s *goquery.Selection) {
		style, exists := s.Attr("style")
		if !exists {
			err = errors.New("style attribute not found")
			log.Println("[ERROR] failed to get style attribute", err)
			return
		}

		groups := backgroundImageRe.FindStringSubmatch(style)
		if len(groups) != 2 {
			err = fmt.Errorf("expected 2 match, got %d", len(groups))
			log.Println("[ERROR] Failed to find image URL", err)
			return
		}

		_, err = url.Parse(groups[1])
		if err != nil {
			log.Println("[ERROR] Invalid parsed image URL: ", groups[1], err)
			return
		}

		images = append(images, groups[1])
	})

	return
}
