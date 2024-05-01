package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/nikgalushko/echoevoke/assets"
	"github.com/nikgalushko/echoevoke/internal/scrapper"
	"github.com/nikgalushko/echoevoke/internal/storage"
	"github.com/nikgalushko/echoevoke/internal/storage/disk"
)

func main() {
	os.ReadFile("echoevoke.db")

	db, err := sql.Open("sqlite3", "echoevoke.db")
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	defer db.Close()
	err = initDB(db)
	if err != nil {
		panic(err)
	}

	chReg := disk.NewChannelRegistry(db)

	err = chReg.RegisterChannel(context.TODO(), "Ateobreaking")
	if err != nil {
		panic(err)
	}
	err = chReg.RegisterChannel(context.TODO(), "lobste_rs")
	if err != nil {
		panic(err)
	}
	err = chReg.RegisterChannel(context.TODO(), "bbbreaking")
	if err != nil {
		panic(err)
	}

	posts := disk.NewPostsStorage(db)
	images := disk.NewImagesStorage(db)

	s := scrapper.New(posts, scrapper.NewImageDownloader(images))

	//go func() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ret := &strings.Builder{}
		err := showPosts(ret, chReg, posts, images)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		resp := index + ret.String() + indexEnd
		w.Write([]byte(resp))
	})

	http.ListenAndServe(":8080", nil)
	//}()

	for i := 0; i < 0; i++ {
		chs, err := chReg.AllChannels(context.Background())
		if err != nil {
			continue
		}

		for _, c := range chs {
			err = s.Scrape(context.TODO(), c)
			if err != nil {
				slog.Error("failed to scrape", slog.String("value", c), slog.Any("err", err))
				continue
			}
		}
		/*dir := filepath.Join("./", strconv.FormatInt(time.Now().Unix(), 10))
		db.Dump(context.TODO(), dir)*/

		slog.Debug("sleeping for 10 minutes")
		time.Sleep(10 * time.Minute)
	}
}

func initDB(db *sql.DB) error {
	fmt.Println("Initializing SQL tables")

	entries, err := assets.SQL.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("failed to read sql directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			sqlBytes, err := assets.SQL.ReadFile(filepath.Join("sql", entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read sql file: %w", err)
			}

			_, err = db.ExecContext(context.Background(), string(sqlBytes))
			if err != nil {
				return fmt.Errorf("failed to execute sql: %w", err)
			}
		}
	}

	return nil
}

func showPosts(w io.Writer, registry *disk.ChannelRegistry, posts *disk.PostsStorage, images *disk.ImagesStorage) error {
	chs, err := registry.AllChannels(context.TODO())
	if err != nil {
		return err
	}

	from := time.Date(2024, time.February, 22, 0, 0, 0, 0, time.UTC)
	to := time.Now()

	tmpl, err := template.New("container").Parse(channelBlock)
	if err != nil {
		return err
	}
	for i, c := range chs {
		postsData, err := posts.GetPosts(context.TODO(), c, from, to)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				continue
			}
			return err
		}

		t := TemplateItem{
			Channel: c,
			IsLast:  i == len(chs)-1,
		}
		for _, p := range postsData {
			tp := TemplatePost{
				Date: p.Date.Format(time.RFC822),
				Text: p.Message,
			}
			for _, id := range p.Images {
				data, err := images.GetImageByID(context.Background(), id)
				if err != nil {
					return err
				}
				tp.Images = append(tp.Images, base64.StdEncoding.EncodeToString(data))
			}
			t.Posts = append(t.Posts, tp)
		}

		err = tmpl.Execute(w, t)
		if err != nil {
			return err
		}
	}
	return nil
}

type TemplateItem struct {
	Channel string
	Posts   []TemplatePost
	IsLast  bool
}

type TemplatePost struct {
	Date   string
	Text   string
	Images []string
}

const channelBlock = `
{{ $ch := .Channel }}
<details id="{{.Channel}}">
	<summary>{{.Channel}}</summary>
	{{range .Posts}}
		<article>
			<header>
				<cite>{{.Date}}</cite>
				<small>
				<a href="#{{$ch}}" class="contrast">↑</a>
				</small>
			</header>
			<p>{{.Text}}</p>
			{{ if gt (len .Images) 0 }}
				<footer>
				<div class="container-fluid">
				<details>
				<summary>Дополнительные изображения</summary>
				{{ range .Images }}
        			<img src="data:image/png;base64, {{.}}">
				{{ end }}
				</details>
				</div>
				</footer>
			{{ end }}
		</article>
	{{end}}
</details>
{{ if not .IsLast }}
	<hr />
{{ end }}
`
const index = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark" />
<link
  rel="stylesheet"
  href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"
/>
<title>Channel Information</title>
</head>
<body>
<div class="container-fluid" style="max-width: 800px; scroll-behavior: smooth;">
`
const indexEnd = `
</div>
</body>
</html>
`

/*
 */
