package w3crawl

import (
	"errors"
	"fmt"
	"io/ioutil"
	urlUtil "net/url"
	"os"
	"path"
	"regexp"
)

var resourceExp, _ = regexp.Compile(`<link(.*?)href="(.*?)"(.*?)/?>`)
var imageExp, _ = regexp.Compile(`<img(.*?)src="(.*?)"(.*?)/?>`)

func getResource(url string) (err error) {
	fr := fetchReq{url: url, ch: make(chan []byte)}
	fetchCh <- fr
	digits, ok := <-fr.ch
	if !ok {
		err = errors.New("CHAN_FAILED")
		return
	}
	SaveURLToFile(url, digits)
	return
}

// SaveURLToFile the content of a url to file
func SaveURLToFile(url string, digits []byte) (err error) {
	ru, err := urlUtil.Parse(url)
	host := ru.Host
	fullPath := ru.Path
	dir, filename := path.Split(fullPath)
	if filename == "" || filename == "/" || filename == "." {
		filename = "index.html"
	}
	dir = "./output/" + host + dir
	fullPath = dir + "/" + filename
	os.MkdirAll(dir, os.ModePerm)

	ioutil.WriteFile(fullPath, digits, os.ModePerm)

	return
}

// FileProcessor write every thing to file
type FileProcessor struct{}

// Process :
func (fp FileProcessor) Process(url, body string) (err error) {
	ru, err := urlUtil.Parse(url)
	host := ru.Host
	scheme := ru.Scheme
	err = SaveURLToFile(url, ([]byte)(body))
	if err != nil {
		return
	}
	matched := resourceExp.FindAllStringSubmatch(body, -1)
	for _, one := range matched {
		newURL := one[2]
		u, err := urlUtil.Parse(newURL)
		if err != nil {
			break
		}
		if u.Host == "" {
			newURL = scheme + "://" + host + newURL
			go getResource(newURL)
		}
	}
	matched = imageExp.FindAllStringSubmatch(body, -1)
	for _, one := range matched {
		newURL := one[2]
		u, err := urlUtil.Parse(newURL)
		if err != nil {
			break
		}
		if u.Host == "" {
			newURL = scheme + "://" + host + newURL
			go getResource(newURL)
		}
	}
	return
}

// PrintProcessor print the url and process to the next
type PrintProcessor struct {
	IfContinue bool
	Another    Processor
}

// Process process the data
func (pp PrintProcessor) Process(url, body string) (err error) {
	// fmt.Printf("found: %s %q\n", url, body)
	fmt.Printf("found: %s\n", url)
	err = nil
	if pp.IfContinue {
		err = pp.Another.Process(url, body)
	}
	return
}
