package main

import (
	"io/ioutil"
	"net/http"
	urlUtil "net/url"
	"regexp"
)

// WebFetcher fetches from the internet
type WebFetcher struct{}

var linkExp, _ = regexp.Compile(`<a(.*?)href="(.*?)"(.*?)>(.*?)</a>`)

// Fetch urls from internet
func (wf WebFetcher) Fetch(url string) (body string, urls []string, err error) {
	rawURL, err := urlUtil.Parse(url)
	if err != nil {
		return
	}
	host := rawURL.Host
	scheme := rawURL.Scheme
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	digits, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
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
