package parser

import (
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParse_OnlyText(t *testing.T) {
	is := is.New(t)
	input, err := os.ReadFile("./testdata/only_text.html")
	is.NoErr(err)

	expectedContent, err := os.ReadFile("./testdata/only_text_expected.md")
	is.NoErr(err)
	expectedDate := time.Date(2024, time.February, 16, 7, 8, 34, 0, time.UTC)

	info, err := ParsePost(input)
	is.NoErr(err)

	is.Equal(info.Content, string(expectedContent))
	is.True(info.Date.Equal(expectedDate))
	is.True(len(info.ImagesLink) == 0)
	is.Equal(int64(116), info.ID)
}

func TestParse_SingleImage(t *testing.T) {
	is := is.New(t)
	input, err := os.ReadFile("./testdata/single_image.html")
	is.NoErr(err)

	expectedDate := time.Date(2024, time.January, 30, 17, 0, 33, 0, time.UTC)

	info, err := ParsePost(input)
	is.NoErr(err)

	is.Equal(info.Content, "")
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 1)
	is.Equal(info.ImagesLink[0], "https://cdn-example.com/single_image.jpg")
	is.Equal(int64(114), info.ID)
}

func TestParse_TextWithOneImage(t *testing.T) {
	is := is.New(t)
	input, err := os.ReadFile("./testdata/text_with_one_image.html")
	is.NoErr(err)

	expectedContent, err := os.ReadFile("./testdata/text_with_one_image_expected.md")
	is.NoErr(err)
	expectedDate := time.Date(2024, time.January, 3, 19, 39, 2, 0, time.UTC)

	info, err := ParsePost(input)
	is.NoErr(err)

	is.Equal(info.Content, string(expectedContent))
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 1)
	is.Equal(info.ImagesLink[0], "https://cdn-example.com/text_with_one_image.jpg")

	is.Equal(int64(111), info.ID)
}

func TestParse_TextWithMultipleImages(t *testing.T) {
	is := is.New(t)
	input, err := os.ReadFile("./testdata/text_with_multiple_images.html")
	is.NoErr(err)

	expectedDate := time.Date(2024, time.February, 16, 18, 58, 53, 0, time.UTC)

	info, err := ParsePost(input)
	is.NoErr(err)

	is.Equal(info.Content, "Text with multiple images")
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 3)
	is.Equal(info.ImagesLink[0], "https://cdn-example.com/text_with_multiple_images_1.jpg")
	is.Equal(info.ImagesLink[1], "https://cdn-example.com/text_with_multiple_images_2.jpg")
	is.Equal(info.ImagesLink[2], "https://cdn-example.com/text_with_multiple_images_3.jpg")

	is.Equal(int64(2171), info.ID)
}
