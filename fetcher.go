package w3crawl

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	urlUtil "net/url"
	"os/exec"
	"path"
	"regexp"
	"runtime"
)

// WebFetcher fetches from the internet
type WebFetcher struct{}

// BrowserFetcher fetches from the internet with phantomjs
type BrowserFetcher struct{}

var pageVisited = map[string]bool{}
var fetchSize = 10

const (
	network   = iota
	phantomjs = iota
)

type fetchReq struct {
	url string
	ch  chan []byte
	t   int
}

var linkExp, _ = regexp.Compile(`<a(.*?)href="(.*?)"(.*?)>(.*?)</a>`)
var fetchCh = make(chan fetchReq, fetchSize)

// GetLinks gets all links from a page
func GetLinks(body, host, scheme string) (urls []string) {
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
	fr := fetchReq{url: url, ch: make(chan []byte), t: network}
	fetchCh <- fr
	digits, ok := <-fr.ch
	if !ok {
		err = errors.New("CHAN_FAILED")
	}
	body = string(digits)
	urls = GetLinks(body, host, scheme)
	return
}

// Fetch urls from internet
func (wf BrowserFetcher) Fetch(url string) (body string, urls []string, err error) {
	rawURL, err := urlUtil.Parse(url)
	if err != nil {
		return
	}
	host := rawURL.Host
	scheme := rawURL.Scheme
	rawURL.Fragment = "" // remove #..
	url = rawURL.String()
	fr := fetchReq{url: url, ch: make(chan []byte), t: phantomjs}
	fetchCh <- fr
	digits, ok := <-fr.ch
	if !ok {
		err = errors.New("CHAN_FAILED")
	}
	body = string(digits)
	urls = GetLinks(body, host, scheme)
	return
}

func serveFetch() {
	ch := make(chan int, fetchSize)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	ppath := path.Dir(filename)

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
			switch f.t {
			case network:
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
			case phantomjs:
				buf, err := exec.Command("phantomjs", ppath+"/browser.js", f.url).Output()
				if err != nil {
					log.Fatal(err)
				}
				f.ch <- buf
			}
		}
		ch <- 0
		go procfr(fr)
	}
}

func init() {
	go serveFetch()
}
