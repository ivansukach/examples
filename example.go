package main

import (
	"fmt"
	"sync"
)

var counter = 0

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

var searchResults = make(map[string]string)
var mutex sync.Mutex

func NewSearchResult(key string, value string) {
	mutex.Lock()
	searchResults[key] = value
	mutex.Unlock()
}
func PrintResults(results map[string]string) {
	fmt.Println("Printing Results")
	mutex.Lock()
	for key, value := range results {
		fmt.Println("Key:", key, "Value:", value)
	}
	mutex.Unlock()
}
func PrintExitStatus(index int) {
	fmt.Printf("exit from goroutine %d \n", index)
}
func PrintWaiting(index int) {
	fmt.Printf("goroutine %d is waiting \n", index)
}
func Crawl(url string, depth int, fetcher Fetcher, wg *sync.WaitGroup, index int) {
	defer wg.Done()
	defer PrintExitStatus(index)
	if depth <= 0 {
		return
	}
	mutex.Lock()
	_, ok1 := searchResults[url]
	mutex.Unlock()
	var body string
	var urls []string
	var err error
	if !ok1 {
		body, urls, err = fetcher.Fetch(url)
		NewSearchResult(url, body)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("found: %s %q\n", url, body)
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(urls))
	for _, u := range urls {
		mutex.Lock()
		_, ok2 := searchResults[u]
		mutex.Unlock()
		if !ok2 {
			go Crawl(u, depth-1, fetcher, &waitGroup, counter) //parallelize a web crawler
			counter++
		} else {
			fmt.Println("This record " + u + " exists")
			waitGroup.Done()
		}
	}
	PrintWaiting(index)
	waitGroup.Wait()
	return
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go Crawl("http://golang.org/", 4, fetcher, &wg, counter)
	counter++
	wg.Wait()
	fmt.Println("Hooray, finally I waited for result of all goroutines")
	PrintResults(searchResults)
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
