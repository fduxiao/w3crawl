package main

import (
	"net/url"
	"regexp"

	"github.com/dcadenas/pagerank"
	"github.com/fduxiao/w3crawl"
	"github.com/wangbin/jiebago"
)

const urlRreqSize = 100

var urlIndex = map[string]int{}

type urlIndexStruct struct {
	u string
	i chan int
}

var urlIndexCh = make(chan urlIndexStruct, urlRreqSize)

func getURLIndex(u string) int {
	var uis urlIndexStruct
	uis.u = u
	uis.i = make(chan int)
	urlIndexCh <- uis
	return <-uis.i
}

func servGetURLIndex() {
	query := func(u string) int {
		i, ok := urlIndex[u]
		// exists
		if ok {
			return i
		}
		// do not contain
		n := len(urlIndex)
		urlIndex[u] = n
		return n
	}
	for {
		select {
		case one := <-urlIndexCh:
			i := query(one.u)
			one.i <- i
		}
	}
}

// used to seg
var seg jiebago.Segmenter

// used to pagerank
var graph = pagerank.New()
var graphCh = make(chan [2]int, urlRreqSize)

func linkGraph(from, to int) {
	d := [2]int{from, to}
	graphCh <- d
}
func servGraph() {
	for {
		select {
		case one := <-graphCh:
			graph.Link(one[0], one[1])
		}
	}
}

func init() {
	go servGetURLIndex()
	go servGraph()
	seg.LoadDictionary("dict.txt")
}

// PageInfo :
type PageInfo struct {
	URL   string   `json:"url"`
	Words []string `json:"words"`
	Rank  float64  `json:"rank"`
}

// SearchProcessor process each page for searching
type SearchProcessor struct {
	AllPage map[int]PageInfo
}

// NewSearchProcessor :
func NewSearchProcessor() (sp SearchProcessor) {
	sp.AllPage = make(map[int]PageInfo)
	return
}

var newlinePattern = regexp.MustCompile("\r|\n|\t")
var commentPattern = regexp.MustCompile("<!--(.*?)--!?>")
var tagPattern = regexp.MustCompile("</?(.*?)/?>")
var blankPattern = regexp.MustCompile(" +|\t+")
var jsPattern = regexp.MustCompile("<script(.*?)></script>")
var stylePattern = regexp.MustCompile("<style(.*?)></style>")

func getWords(body string) (words []string) {
	// remove all newline and tab
	body = newlinePattern.ReplaceAllString(body, "")
	// remove all comments that contains tag
	body = commentPattern.ReplaceAllString(body, "")
	body = jsPattern.ReplaceAllString(body, "")
	body = stylePattern.ReplaceAllString(body, "")
	temp1 := tagPattern.Split(body, -1)
	temp2 := []string{}

	for _, one := range temp1 {
		one = blankPattern.ReplaceAllString(one, " ")
		if one == "" {
			continue
		}
		// blanks
		if one == " " {
			continue
		}
		for _, i := range blankPattern.Split(one, -1) {
			temp2 = append(temp2, i)
		}
	}
	for _, s := range temp2 {
		ch := seg.Cut(s, true)
		for one := range ch {
			words = append(words, one)
		}
	}
	return
}

// Process get words, add links
func (sp *SearchProcessor) Process(u, body string) (err error) {
	pgInfo := PageInfo{}
	pgInfo.URL = u
	ru, err := url.Parse(u)
	if err != nil {
		return
	}
	urls := w3crawl.GetLinks(body, ru.Host, ru.Scheme)
	id := getURLIndex(u)
	for _, one := range urls {
		newID := getURLIndex(one)
		graph.Link(id, newID)
	}
	pgInfo.Words = getWords(body)
	sp.AllPage[id] = pgInfo
	return
}

// BuildRank gets the rank for each index
func (sp *SearchProcessor) BuildRank(probabilityOfFollowingALink, tolerance float64) (err error) {
	graph.Rank(probabilityOfFollowingALink, tolerance, func(identifier int, rank float64) {
		pgInfo := sp.AllPage[identifier]
		if len(pgInfo.Words) == 0 {
			delete(sp.AllPage, identifier)
			return
		}
		pgInfo.Rank = rank
		sp.AllPage[identifier] = pgInfo
	})
	return
}
