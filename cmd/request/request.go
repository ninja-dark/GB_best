package request

import (
	
	"context"
	"net/http"
	"time"
	log "github.com/sirupsen/logrus"
	"best_go/cmd/model"
	"best_go/cmd/page"
)

type requester struct{
	timeout time.Duration
}

func NewRequester(timeout time.Duration) requester{
	return requester{timeout: timeout}
}

func (r requester) Get(ctx context.Context, url string) (model.Page, error){
	log.SetFormatter(&log.JSONFormatter{})
	standardFields := log.Fields{
		"host": "srv42",
	}
	hlog := log.WithFields(standardFields)
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		cl := &http.Client{
			Timeout: r.timeout,
		}
		req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if err != nil {
			hlog.WithFields(log.Fields{"uid": 101345}).Error("Error by sending request")

			return nil, err
		}
		body, err := cl.Do(req)
		if err != nil {
			hlog.WithFields(log.Fields{"uid": 101345}).Error("get new page error")

			return nil, err
		}
		defer body.Body.Close()
		page, err := page.NewPage(body.Body)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
}