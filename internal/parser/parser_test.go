package parser

import (
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParse_OnlyText(t *testing.T) {
	is := is.New(t)
	input, err := os.Open("./testdata/only_text.html")
	is.NoErr(err)
	t.Cleanup(func() { input.Close() })

	expectedContent, err := os.ReadFile("./testdata/only_text_expected.md")
	is.NoErr(err)
	expectedDate := time.Date(2024, time.February, 16, 7, 8, 34, 0, time.UTC)

	info, err := Parse(input)
	is.NoErr(err)

	is.Equal(info.Content, string(expectedContent))
	is.True(info.Date.Equal(expectedDate))
	is.True(len(info.ImagesLink) == 0)
}

func TestParse_SingleImage(t *testing.T) {
	is := is.New(t)
	input, err := os.Open("./testdata/single_image.html")
	is.NoErr(err)
	t.Cleanup(func() { input.Close() })

	expectedDate := time.Date(2024, time.January, 30, 17, 0, 33, 0, time.UTC)

	info, err := Parse(input)
	is.NoErr(err)

	is.Equal(info.Content, "")
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 1)
	is.Equal(info.ImagesLink[0], "https://cdn-example.com/single_image.jpg")
}

func TestParse_TextWithOneImage(t *testing.T) {
	is := is.New(t)
	input, err := os.Open("./testdata/text_with_one_image.html")
	is.NoErr(err)
	t.Cleanup(func() { input.Close() })

	expectedContent, err := os.ReadFile("./testdata/text_with_one_image_expected.md")
	is.NoErr(err)
	expectedDate := time.Date(2024, time.January, 3, 19, 39, 2, 0, time.UTC)

	info, err := Parse(input)
	is.NoErr(err)

	is.Equal(info.Content, string(expectedContent))
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 1)
	is.Equal(info.ImagesLink[0], "https://cdn-example.com/text_with_one_image.jpg")
}

func TestParse_TextWithMultipleImages(t *testing.T) {
	is := is.New(t)
	input, err := os.Open("./testdata/text_with_multiple_images.html")
	is.NoErr(err)
	t.Cleanup(func() { input.Close() })

	expectedDate := time.Date(2024, time.February, 16, 18, 58, 53, 0, time.UTC)

	info, err := Parse(input)
	is.NoErr(err)

	is.Equal(info.Content, "Text with multiple images")
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 3)
	is.Equal(info.ImagesLink[0], "https://cdn-example.com/text_with_multiple_images_1.jpg")
	is.Equal(info.ImagesLink[1], "https://cdn-example.com/text_with_multiple_images_2.jpg")
	is.Equal(info.ImagesLink[2], "https://cdn-example.com/text_with_multiple_images_3.jpg")
}
