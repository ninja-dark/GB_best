package main

import (
	"best_go/internal/crawler"
	"best_go/internal/model"
	"best_go/internal/request"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	log "github.com/sirupsen/logrus"
)




type Config struct {
	MaxDepth   uint64
	MaxResults int
	MaxErrors  int
	URL        string
	Timeout    int
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	standardFields := log.Fields{
		"host": "srv42",
	}

	hlog := log.WithFields(standardFields)

	cfg := &Config{
		MaxDepth:   3,
		MaxResults: 10,
		MaxErrors:  500,
		URL:        "https://telegram.com",
		Timeout:    10,
	}

	r  := request.NewRequester(time.Duration(cfg.Timeout))
	cr := crawler.NewCrawler(r, cfg.MaxDepth)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	go cr.Scan(ctx, cfg.URL, cfg.MaxDepth)
	go processResult(ctx, cancel, cr, *cfg)

	sigCh := make(chan os.Signal) // Создаем канал для приема сигналов
	signal.Notify(sigCh, syscall.SIGINT)

	sigChDepth := make(chan os.Signal) // Создаем канал для прима сингнала
	signal.Notify(sigChDepth, syscall.SIGUSR1)

	for {
		select {
		case <-ctx.Done(): // Если всё завершили - выходим
			return
		case <-sigCh:
			hlog.WithFields(log.Fields{"uid": 100500}).Debug("context closed")
			cancel() // Если пришёл сигнал SigInt - завершаем контекст
		case <-sigChDepth:
			cr.AddMaxDepth()
			hlog.WithFields(log.Fields{"uid": 100500}).Info("Max depth increased")
		}

	}

}

func processResult(ctx context.Context, cancel func(), cr model.Crawler, cfg Config) {
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
