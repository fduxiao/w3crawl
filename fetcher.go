package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	urlUtil "net/url"
	"regexp"
)

// WebFetcher fetches from the internet
type WebFetcher struct{}

var pageVisited = map[string]bool{}
var fetchSize = 60

type fetchReq struct {
	url string
	ch  chan []byte
}

var linkExp, _ = regexp.Compile(`<a(.*?)href="(.*?)"(.*?)>(.*?)</a>`)
var fetchCh = make(chan fetchReq, fetchSize)

// Fetch urls from internet
func (wf WebFetcher) Fetch(url string) (body string, urls []string, err error) {
	rawURL, err := urlUtil.Parse(url)
	if err != nil {
		return
	}
	host := rawURL.Host
	scheme := rawURL.Scheme
	rawURL.Fragment = "" // remove #..
	url = rawURL.String()
	fr := fetchReq{url: url, ch: make(chan []byte)}
	fetchCh <- fr
	digits, ok := <-fr.ch
	if !ok {
		err = errors.New("CHAN_FAILED")
	}
	body = string(digits)
	matched := linkExp.FindAllStringSubmatch(body, -1)

	for _, one := range matched {
		newURL := one[2]
		u, err := urlUtil.Parse(newURL)
		if err != nil {
			break
		}
		if u.Host == "" {
			newURL = host + newURL
		}

		if u.Scheme == "" {
			newURL = scheme + "://" + newURL
		}
		urls = append(urls, newURL)
	}
	return
}

func serveFetch() {
	ch := make(chan int, fetchSize)
	for fr := range fetchCh {
		visited, ok := pageVisited[fr.url]
		// the page has been visited
		if ok && visited {
			close(fr.ch)
			continue
		}
		pageVisited[fr.url] = true
		procfr := func(f fetchReq) {
			defer func() {
				<-ch
				close(f.ch)
			}()
			resp, err := http.Get(f.url)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			f.ch <- buf
		}
		ch <- 0
		go procfr(fr)
	}
}

func init() {
	go serveFetch()
}
