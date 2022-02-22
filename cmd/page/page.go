package page

import (
	"best_go/cmd/model"
	"io"

	"github.com/PuerkitoBio/goquery"
)

type page struct {
	doc *goquery.Document
}

func NewPage (raw io.Reader)(model.Page, error){
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil{
		return nil, err
	}
	return &page{doc: doc}, nil
}

func (p *page) ParsePager() (string, []string){
	title := p.doc.Find("title").First().Text()

	var urls []string
	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			urls = append(urls, url)
		}
	})
	return title, urls
}




