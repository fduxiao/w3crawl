package main

import (
	"flag"
	"strconv"

	"github.com/fduxiao/w3crawl"
)

func main() {
	var processor w3crawl.Processor // used to process each url and body
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
		processor = w3crawl.FileProcessor{}
	}
	if *p {
		pp := w3crawl.PrintProcessor{}
		pp.IfContinue = false
		if hasPrevProcessor {
			pp.IfContinue = true
			pp.Another = processor
		}
		hasPrevProcessor = true
		processor = pp
	}

	// no processor is specified
	if !hasPrevProcessor {
		processor = w3crawl.PrintProcessor{IfContinue: false}
	}

	w3crawl.StartCrawl(host, n, w3crawl.WebFetcher{}, processor)
}
