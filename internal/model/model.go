package model

import (
	"context"
)

type CrawlResult struct {
	Err   error
	Title string
	URL   string
}

type Crawler interface {
	Scan(ctx context.Context, url string, depth uint64)
	ChanResult() <-chan CrawlResult
	AddMaxDepth()
}
type Requeser interface{
	Get(ctx context.Context, url string) (Page, error)
}

type Page interface {
	ParsePager() (string, []string)
}

