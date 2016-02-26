package main

import (
	"flag"
	"fmt"
	"strconv"
)

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

func main() {
	var processor Processor // used to process each url and body
	hasPrevProcessor := false

	f := flag.Bool("f", false, "use the file processor")
	p := flag.Bool("p", false, "use the print processor")

	// parse the args
	flag.Parse()
	args := flag.Args()
	host := args[0]
	n, err := strconv.Atoi(args[1])
	if err != nil {
		panic(err)
	}

	if *f {
		hasPrevProcessor = true
		processor = FileProcessor{}
	}
	if *p {
		pp := printProcessor{}
		pp.ifContinue = false
		if hasPrevProcessor {
			pp.ifContinue = true
			pp.another = processor
		}
		hasPrevProcessor = true
		processor = pp
	}

	// no processor is specified
	if !hasPrevProcessor {
		processor = printProcessor{ifContinue: false}
	}

	// true part starts here
	ch := make(chan int, crawlSize)
	num := 1 // a go routine is running at the beginning
	go Crawl(host, n, WebFetcher{}, processor, ch)

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

type printProcessor struct {
	ifContinue bool
	another    Processor
}

func (pp printProcessor) Process(url, body string) (err error) {
	// fmt.Printf("found: %s %q\n", url, body)
	fmt.Printf("found: %s\n", url)
	err = nil
	if pp.ifContinue {
		err = pp.another.Process(url, body)
	}
	return
}
