package main

import (
	"io/ioutil"
	"net/http"
	urlUtil "net/url"
	"os"
	"path"
	"regexp"
)

var resourceVisited = map[string]bool{}

var resourceExp, _ = regexp.Compile(`<link(.*?)href="(.*?)"(.*?)>(.*?)</link>`)

// FileProcessor write every thing to file
type FileProcessor struct{}

func getResource(url string) (err error) {
	visited, ok := resourceVisited[url]
	if ok && visited {
		return nil
	}
	resourceVisited[url] = true
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	digits, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	Process(url, string(digits))
	return
}

// Process everyting into the file
func Process(url, body string) (err error) {
	ru, err := urlUtil.Parse(url)
	host := ru.Host
	fullPath := ru.Path
	dir, filename := path.Split(fullPath)
	if filename == "" || filename == "/" || filename == "." {
		filename = "index.html"
	}
	dir = "./output/" + host + dir
	fullPath = dir + filename
	if visited, ok := resourceVisited[fullPath]; ok && visited {
		return nil
	}
	resourceVisited[fullPath] = true
	os.MkdirAll(dir, os.ModePerm)

	ioutil.WriteFile(fullPath, ([]byte)(body), os.ModePerm)

	return
}

// Process :
func (fp FileProcessor) Process(url, body string) (err error) {
	ru, err := urlUtil.Parse(url)
	host := ru.Host
	scheme := ru.Scheme
	err = Process(url, body)
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
	return
}
