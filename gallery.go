package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Image struct {
	ImageUrl string
	AltText  string
}
type Gallery struct {
	//Season   string
	//Designer string
	//Year     string
	Permalink   string
	Title       string
	Description string
	Images      []Image
}

func getImages(doc *goquery.Document) (urls []Image, err error) {
	d := goquery.CloneDocument(doc)
	d.Find("img").Each(func(i int, s *goquery.Selection) {
		imgalt := s.AttrOr("alt", "")
		imgurl, err := url.Parse(s.AttrOr("src", "null"))
		if err != nil {
			// invalid src?
			log.Fatal(err)
		}
		// remove query string
		imgurl.RawQuery = ""
		surl := imgurl.String()
		// ignore non-jpg (.gif, etc)
		if !strings.HasSuffix(surl, "jpg") {
			return
		}

		urls = append(urls, Image{
			ImageUrl: surl,
			AltText:  imgalt,
		})
	})
	return urls, nil
}

func getMeta(doc *goquery.Document) (title, desc, perma string, err error) {
	d := goquery.CloneDocument(doc)
	title = d.Find("meta[property=\"og:title\"]").First().AttrOr("content", "notitle")
	desc = d.Find("meta[property=\"og:description\"]").First().AttrOr("content", "nodesc")
	perma = d.Find("meta[property=\"og:url\"]").First().AttrOr("content", "nourl")

	return
}

func NewGalleryFromURL(wwdurl string) (g Gallery, err error) {
	c := http.Client{
		Timeout: Timeout,
	}

	req, err := http.NewRequest(http.MethodGet, wwdurl, nil)
	if err != nil {
		return
	}
	req.Header = Headers

	r, err := c.Do(req)
	if err != nil {
		return
	}
	if r.Body == nil {
		return g, fmt.Errorf("no data returned")
	}

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return
	}

	// get meta
	g.Title, g.Description, g.Permalink, err = getMeta(doc)
	if err != nil {
		return
	}

	// get images
	imgs, err := getImages(doc)
	if err != nil {
		return
	}

	g.Images = imgs

	return g, nil
}

func (g Gallery) ImageUrls() (urls []string) {
	for _, i := range g.Images {
		urls = append(urls, i.ImageUrl)
	}
	return urls
}
