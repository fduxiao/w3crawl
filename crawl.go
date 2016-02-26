package w3crawl

import "fmt"

const (
	// NEWCRAWLLING means a new go routine to crawl is starting
	NEWCRAWLLING = iota
	// STOPCRAWLLING means a previous goroutine to crawl has ended crawlling
	STOPCRAWLLING = iota
)

var crawlSize = 200

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

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		if err.Error() != "CHAN_FAILED" {
			fmt.Println(err)
		}
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

// StartCrawl starts the crawlling goroutine
func StartCrawl(url string, depth int, fetcher Fetcher, processor Processor) {
	// true part starts here
	ch := make(chan int, crawlSize)
	num := 1 // a go routine is running at the beginning
	go Crawl(url, depth, WebFetcher{}, processor, ch)

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
