package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"
)

var Headers map[string][]string = map[string][]string{
	"Host":       []string{"wwd.com"},
	"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"},
}
var Timeout = 10 * time.Second

//go:embed static
var Static embed.FS

func sanitiseWwdURL(wwdurl string) (string, error) {
	if !strings.HasPrefix(wwdurl, "https://wwd.com/fashion-news/shows-reviews/gallery") {
		return wwdurl, fmt.Errorf("not a valid wwd.com gallery url")
	}
	return wwdurl, nil
}

func main() {
	listen := flag.String("l", ":8083", "listen address")
	flag.Parse()

	staticFs := fs.FS(Static)
	httpFs, err := fs.Sub(staticFs, "static")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(httpFs)))
	mux.HandleFunc("POST /loadwwd", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Printf("error parsing form: %v", err)
			fmt.Fprintf(w, "<div class=\"error\">error parsing form</div>")
			return
		}
		postUrl := r.PostForm.Get("wwdurl")
		postUrl, err = sanitiseWwdURL(postUrl)
		if err != nil {
			log.Printf("error: %v", err)
			// i'm fine putting the error message on the client-facing page here
			// because the only error it could be is an invalid url one
			fmt.Fprintf(w, "<div class=\"error\">error: %v</div>", err)
			return
		}

		g, err := NewGalleryFromURL(postUrl)
		if err != nil {
			log.Printf("error getting images: %v", err)
			fmt.Fprintf(w, "<div class=\"error\">error getting images</div>")
			return
		}

		imgs := g.Images

		fmt.Fprintf(w, "<div class=\"info\">%s <a href=\"%s\">[^]</a></div>", g.Title, g.Permalink)
		fmt.Fprintf(w, "<div class=\"grid\">")
		for _, s := range imgs {
			fmt.Fprintf(w, "<div><a target=\"_blank\" href=\"%s?w=3000\"><img src=\"%s?w=300&h=450\" alt=\"%s\"></a></div>", s.ImageUrl, s.ImageUrl, s.AltText)
		}
		fmt.Fprintf(w, "</div>")
		return
	})

	http.ListenAndServe(*listen, mux)
}
