package parser

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

type PostInfo struct {
	ID         int64
	Content    string
	Date       time.Time
	ImagesLink []string
}

func ParsePage(data []byte) ([]PostInfo, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var posts []PostInfo
	doc.Find("div.tgme_widget_message_wrap").Each(func(i int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			log.Println("[ERROR] failed to get HTML from selection", err)
			return
		}

		info, err := ParsePost([]byte(html))
		if err != nil {
			log.Println("[ERROR] failed to parse post", err)
			return
		}

		posts = append(posts, info)
	})

	return posts, nil
}

// ParsePost returns the content, date and images of a post
func ParsePost(data []byte) (PostInfo, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
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

	postID, err := selectPostID(doc)
	if err != nil {
		return PostInfo{}, err
	}

	return PostInfo{Content: content, Date: date, ImagesLink: images, ID: postID}, nil
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

// select data-post by attribute
func selectPostID(doc *goquery.Document) (postID int64, err error) {
	var (
		dataPost string
		exists   bool
	)
	doc.Find("div.tgme_widget_message[data-post]").Each(func(i int, s *goquery.Selection) {
		dataPost, exists = s.Attr("data-post")
		if !exists {
			err = errors.New("data-post attribute not found")
			log.Println("[ERROR] failed to get data-post attribute", err)
		}
	})
	if err != nil {
		return
	}

	parts := strings.Split(dataPost, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("expected 2 parts, got %d", len(parts))
		log.Println("[ERROR] failed to split data-post", err)
		return

	}
	postID, err = strconv.ParseInt(parts[1], 10, 64)
	return
}

// selectDate returns the date of the post
func selectDate(doc *goquery.Document) (date time.Time, err error) {
	doc.Find("span.tgme_widget_message_meta").Find("[datetime]").Each(func(i int, s *goquery.Selection) {
		date, err = time.Parse("2006-01-02T15:04:05-07:00", s.AttrOr("datetime", ""))
		if err != nil {
			log.Println("[ERROR] failed to parse date", err)
		}
	})
	return
}

var backgroundImageRe = regexp.MustCompile(`background-image:url\s?\('(.*)'\)`)

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
