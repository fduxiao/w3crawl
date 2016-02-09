package main

import (
	"fmt"
)

var pageVisited = map[string]bool{}

const (
	// NEWCRAWLLING means a new go routine to crawl is starting
	NEWCRAWLLING = iota
	// STOPCRAWLLING means a previous goroutine to crawl has ended crawlling
	STOPCRAWLLING = iota
)

// Fetcher defines how a page is fetched
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Processor processes each fetched process
type Processor interface {
	// Process read the content and process it
	Process(url, body string) (err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, processor Processor, calc chan int) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	defer func() {
		// inform a goroutine terminates
		calc <- STOPCRAWLLING
	}()

	if depth <= 0 {
		return
	}

	visited, ok := pageVisited[url]
	// the page has been visited
	if ok && visited {
		return
	}

	pageVisited[url] = true

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		err := processor.Process(url, body)
		if err != nil {
			fmt.Println(err)
		}
	}()
	for _, u := range urls {
		// inform a new goroutine
		calc <- NEWCRAWLLING
		go Crawl(u, depth-1, fetcher, processor, calc)
	}
	return
}

func main() {
	ch := make(chan int, 200)
	num := 1 // a go routine is running at the beginning
	// go Crawl("http://golang.org/", 4, fetcher, printProcessor{}, ch)
	go Crawl("http://w3school.com.cn/", 2, WebFetcher{}, FileProcessor{}, ch)
	for {
		select {
		case i := <-ch:
			switch i {
			case NEWCRAWLLING:
				num++
			case STOPCRAWLLING:
				num--
			}
		default:
			// exit when no go routine exists
			if num == 0 {
				return
			}
		}
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

type printProcessor struct{}

func (pp printProcessor) Process(url, body string) (err error) {
	// fmt.Printf("found: %s %q\n", url, body)
	fmt.Printf("found: %s\n", url)
	err = nil
	return
}
