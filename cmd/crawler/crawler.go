package crawler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"best_go/cmd/model"

)



type crawler struct {
	r		 model.Requeser
	maxDepth uint64
	res      chan model.CrawlResult
	visited  map[string]bool
	mu       sync.RWMutex
}

func NewCrawler(r model.Requeser, maxDepth uint64)(*crawler){
	return &crawler{
		r:		  r,
		maxDepth: maxDepth,
		res:      make(chan model.CrawlResult),
		visited:  make(map[string]bool),
		mu:       sync.RWMutex{},
	}
}

func (c *crawler) AddMaxDepth() {
	atomic.AddUint64(&c.maxDepth, 2)
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
	select {
	case <-ctx.Done(): // Если контекст завершен - прекращаем выполнение
		return
	default:
		body, err := c.r.Get(ctx, url)
		if err != nil {
			c.res <- model.CrawlResult{Err: err}
			return
		}
		b, urls := body.ParsePager()
		
		for _, u := range urls {
			if c.visited[u] {
				fmt.Println("Skip, url")
				return
			
			}else{
				go c.Scan(ctx, u, depth-1)
			}		
		
		c.res <- model.CrawlResult{
			Title: b,
			URL:   url,
		}

		c.visited[url] = true
		}
	}

}

func (c *crawler) ChanResult() <-chan model.CrawlResult{
	return c.res
}
