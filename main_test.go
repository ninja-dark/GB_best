package main

import (
	"best_go/cmd/crawler"
	"best_go/cmd/model"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testWebPage = "http://yandex.com/"




func TestParsePage(t *testing.T){
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(100)*time.Millisecond)
	page := model.Page{
		URL: testWebPage,
	}

	gotBody, gotUrls, gotErr := page.ParsePage(ctx, testWebPage)
		
	wantBody := "Яндекс"

	
			if  assert.Nil(t, gotErr){
					assert.Equal(t,wantBody, gotBody, "they should be equal") 
			}
	
			fmt.Println("url is ", gotUrls)
}

func TestScan(t *testing.T){

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	NCrawler := crawler.NewCrawler{
		MaxDepth: 4,
		Res:  make(chan crawler.CrawlResult),
		Visited: make(map[string]bool),
		Mu: sync.RWMutex{},
	}

	go NCrawler.Scan(ctx, "http://yandex.com/", NCrawler.MaxDepth)
	 result := NCrawler.Res
	want := 4
	 assert.Equal(t, want, result)

	 


}

	

