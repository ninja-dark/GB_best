package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Pager interface {
	parsePager(ctx context.Context, url string) (body string, urls []string, err error)
}

type page struct {
	URL string
}

func (p *page) parsePage(ctx context.Context, url string) (string, []string, error) {
	select {
	case <-ctx.Done():
		return "", nil, nil
	default:
		cl := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", nil, err
		}
		body, err := cl.Do(req)
		if err != nil {
			return "", nil, err
		}
		defer body.Body.Close()

		doc, err := goquery.NewDocumentFromReader(body.Body)
		if err != nil {
			return "", nil, err
		}

		title := doc.Find("title").First().Text()

		var urls []string

		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			url, ok := s.Attr("href")
			if ok {
				urls = append(urls, url)
			}

		})
		return title, urls, nil
	}

}

type CrawlResult struct {
	Err   error
	Title string
	Url   string
}

type Crawler interface {
	Scan(ctx context.Context, url string, depth uint64)
	ChanResult() <-chan CrawlResult
}

type crawler struct {
	maxDepth uint64
	res      chan CrawlResult
	visited  map[string]bool
	mu       sync.RWMutex
}


func (c *crawler) AddMaxDepth(increase uint64) {
	atomic.AddUint64(&c.maxDepth, increase)
}

func (c *crawler) Scan(ctx context.Context, url string, depth uint64) {
	if depth <= 0 {
		return
	}
	c.mu.RLock()
	store := c.visited[url]
	c.mu.RUnlock()
	if store {
		fmt.Println("Skip, url")
		return
	}

	page := page{
		URL: url,
	}
	select {
	case <-ctx.Done(): //Если контекст завершен - прекращаем выполнение
		return
	default:
		body, urls, err := page.parsePage(ctx, url)
		if err != nil {
			c.res <- CrawlResult{Err: err}
			return
		}

		for _, u := range urls {
			if c.visited[u] {
				fmt.Println("Skip, url")
				return
			} else {
				go c.Scan(ctx, u, depth-1)
			}
		}

		c.res <- CrawlResult{
			Title: body,
			Url:   url,
		}

		c.visited[url] = true
	}

}

func (c *crawler) ChanResult() <-chan CrawlResult {
	return c.res
}

type Config struct {
	MaxDepth   uint64
	MaxResults int
	MaxErrors  int
	Increase   uint64
	Url        string
	Timeout    int
}

func main() {
	cfg := Config{
		MaxDepth:   3,
		MaxResults: 10,
		MaxErrors:  500,
		Increase:   2,
		Url:        "https://telegram.com",
		Timeout:    10,
	}
	var mu sync.RWMutex

	cr := &crawler{
		maxDepth: cfg.MaxDepth,
		res:      make(chan CrawlResult),
		visited:  make(map[string]bool),
		mu:       mu,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	go cr.Scan(ctx, cfg.Url, cfg.MaxDepth)
	go processResult(ctx, cancel, cr, cfg)

	sigCh := make(chan os.Signal) //Создаем канал для приема сигналов
	signal.Notify(sigCh, syscall.SIGINT)

	sigChDepth := make(chan os.Signal) //Создаем канал для прима сингнала
	signal.Notify(sigChDepth, syscall.SIGUSR1)

	for {
		select {
		case <-ctx.Done(): //Если всё завершили - выходим
			return
		case <-sigCh:
			cancel() //Если пришёл сигнал SigInt - завершаем контекст
		case <-sigChDepth:
			cr.AddMaxDepth(cfg.Increase)
		}

	}

}

func processResult(ctx context.Context, cancel func(), cr Crawler, cfg Config) {
	var maxResult, maxErrors = cfg.MaxResults, cfg.MaxErrors

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-cr.ChanResult():
			if msg.Err != nil {
				maxErrors--
				if maxErrors <= 0 {
					fmt.Println("max number of errors received")
					return
				}
			} else {
				fmt.Println(msg)
				maxResult--
				if maxResult <= 0 {
					fmt.Println("max number of errors received")
					cancel()
					return
				}
			}
		}
	}
}
