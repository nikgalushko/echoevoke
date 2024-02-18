package parser

import (
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestParse(t *testing.T) {
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
}
