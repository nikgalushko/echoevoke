package parser

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestParse(t *testing.T) {
	is := is.New(t)
	input, err := os.Open("./testdata/only_text.html")
	is.NoErr(err)
	t.Cleanup(func() { input.Close() })

	expectedContent, err := os.ReadFile("./testdata/only_text_expected.md")
	is.NoErr(err)

	info, err := Parse(input)
	is.NoErr(err)

	is.Equal(info.Content, string(expectedContent))
}
