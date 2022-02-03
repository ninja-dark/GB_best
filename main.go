package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/PuerkitoBio/goquery"
)

type Page interface {
	Page(url string) (body string, urls []string, err error)
}

type URLs struct {
	c   map[string]bool
	mux sync.Mutex
}

func (store *URLs) setVisited(url string) bool {
	store.mux.Lock()
	_, exists := store.c[url]
	defer store.mux.Unlock()

	if exists {
		return true
	} else {
		store.c[url] = true
		return false
	}

}
func NewPage(raw io.Reader) (doc *goquery.Document, error error) {
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
func Get(url string) (b *goquery.Document, error error) {
	cl := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	body, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer body.Body.Close()
	b, err = NewPage(body.Body)
	if err != nil {
		fmt.Errorf("Error, %s", err)
	}
	return b, nil
}

func Getlinks(url string) []string {
	var urls []string
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil
	}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			urls = append(urls, url)
		}

	})
	return urls
}

func Crawl(url string, depth uint64, store URLs, wg *sync.WaitGroup) {
	defer wg.Done()
	if depth <= 0 {
		return
	}
	visited := store.setVisited(url)
	if visited {
		fmt.Println("Skip, url")
		return
	}

	urls := Getlinks(url)
//	var increase uint64
//	increase = 2
	sigChDepth := make(chan os.Signal)         //Создаем канал для прима сингнала
	signal.Notify(sigChDepth, syscall.SIGUSR1) // Подписываемся на сигнал SIGUSR1

/*	for {
		select {
		case <-sigChDepth:
			AddMaxDepth(depth, increase)
		}

	}
*/
	fmt.Printf("found: %q\n", url)

	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, store, wg)
	}
	return

}
func AddMaxDepth(depth uint64, increase uint64) {
	atomic.AddUint64(&depth, increase)
}

func main() {
	store := URLs{c: make(map[string]bool)}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go Crawl("https://telegram.org/", 3, store, wg)
	wg.Wait()
}

type fakeFetcher map[string]*Resuilt

type Resuilt struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}
