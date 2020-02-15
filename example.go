package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup
var counter = 0

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}
type KeyValue struct {
	key   string
	value string
}

var firstStart = true
var channelsAreClosed = false

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, ch chan *KeyValue, ch1 chan map[string]string, index int) {
	fmt.Printf("Goroutine №%d started \n", index)
	kv := new(KeyValue)
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 && !channelsAreClosed {
		wg.Done()
		return
	}
	records, ok1 := <-ch1
	_, ok2 := records[url]
	var body string
	var urls []string
	var err error
	if ok1 && !ok2 {
		body, urls, err = fetcher.Fetch(url)
		kv.key = url
		kv.value = body
		if !channelsAreClosed {
			ch <- kv
		}
		if err != nil {
			fmt.Println("Error!!!")
			fmt.Println(err)
			wg.Done()
			return
		}
	}
	fmt.Printf("Value has been pushed in channel ch : %s in goroutine %d\n", kv.key, index)
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		_, ok2 = records[u]
		if !ok2 {
			wg.Add(1)
			go Crawl(u, depth-1, fetcher, ch, ch1, counter) //parallelize a web crawler
			counter++
		} else {
			fmt.Println("This record " + u + " exists")
		}
	}
	if !firstStart {
		defer wg.Done()
	}
	if firstStart {
		firstStart = false
		wg.Wait()
		channelsAreClosed = true
		fmt.Println("Goroutines just have finished their work")
		//defer close(ch)
		//defer close(ch1)
		close(ch)
		close(ch1)
	}
	return
}

func main() {
	ch := make(chan *KeyValue)
	var searchResults = make(map[string]string)
	ch1 := make(chan map[string]string)
	go Crawl("http://golang.org/", 4, fetcher, ch, ch1, counter)
	counter++
	if !channelsAreClosed {
		ch1 <- searchResults
	}
	for {
		fmt.Println("Channel is waiting")
		var r *KeyValue
		var ok bool
		if !channelsAreClosed {
			r, ok = <-ch
		} else {
			fmt.Println("Channel is closed")
			return
		}

		fmt.Printf("Data has been taken from channel ch\n")
		if ok && r != nil && r.key != "" {
			searchResults[r.key] = r.value
			if !channelsAreClosed /*&& wg.Wait()*/ {
				ch1 <- searchResults
			} else {
				fmt.Println("Channel is closed")
				break
			}
		} else {
			if !ok {
				fmt.Println("Channel is closed")
				break
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
