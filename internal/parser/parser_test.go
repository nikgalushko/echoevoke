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
	expectedDate := time.Date(2024, time.January, 30, 17, 0, 33, 0, time.UTC)

	info, err := Parse(input)
	is.NoErr(err)

	is.Equal(info.Content, string(expectedContent))
	is.True(info.Date.Equal(expectedDate))

	is.True(len(info.ImagesLink) == 1)
	is.Equal(info.ImagesLink[0], "https://cdn4.cdn-telegram.org/file/hGbVB30bIbZGyRIOrxI-4TJjZLPhubd2Hsxfw1w2PBLm_HrzEiqRwFfErVW9tC3mhxKXk9J9i_C3gXlEKtEJVcbCJCQ6UUpIE_WZ7HS7sR2VqrVGKpzbSu-g-WMGWPhkzuQMRgWA59A53zqW-CWkxVcOU4a9EhVJmvDHgKM0P4Ivo4tYZMO63T49dEqNoRoDXts9XD6jRFC0KHxEs_YT0EulhqcA4oELbLw8oejPURdeu57DCVKatqo0tsHozdyho9zjAb0boTMFc_qRpp4pzftcCFqH0zUDskLbYrm8HriXVLu2cjseAO7EJbS7Hjx6sCVpIBbopSz2JiNg-e2GYQ.jpg")
}
