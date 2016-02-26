package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/fduxiao/w3crawl"
)

func crawlSave(url string, depth int, filename string) {
	var processor w3crawl.Processor
	pp := w3crawl.PrintProcessor{}
	pp.IfContinue = true
	sp := NewSearchProcessor()
	pp.Another = &sp
	processor = pp

	w3crawl.StartCrawl(url, depth, w3crawl.WebFetcher{}, processor)
	sp.BuildRank(0.85, 0.001)

	var r []PageInfo
	for _, one := range sp.AllPage {
		r = append(r, one)
	}
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	f.Write(b)
}

func openServe(filename string, host string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var pgInfoList []PageInfo
	json.Unmarshal(b, &pgInfoList)
	fmt.Println("start net service")

	handler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Path[1:]
		ch := seg.Cut(query, true)
		var keywords []string
		for i := range ch {
			keywords = append(keywords, i)
		}
		display, err := json.Marshal(keywords)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "%s", display)
	}
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(host, nil)
	if err != nil {
		panic(err)
	}
}

func main() {

	sf := flag.String("s", "", "<file path> the file to save")
	lf := flag.String("l", "", "<file path> the file to load")

	// parse the args
	flag.Parse()
	args := flag.Args()
	host := args[0]
	if *sf != "" {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}
		crawlSave(host, n, *sf)
	}
	if *lf != "" {
		openServe(*lf, host)
	}
}
