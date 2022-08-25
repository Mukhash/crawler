package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type M struct {
	Store         map[int]map[string]struct{}
	FailedCounter uint64
	mu            sync.RWMutex
}

func (m *M) Add(depth int, u string) {
	m.mu.Lock()
	if m.Store[depth] == nil {
		m.Store[depth] = map[string]struct{}{}
	}
	m.Store[depth][u] = struct{}{}
	m.mu.Unlock()
}

func (m *M) AddFailed() {
	atomic.AddUint64(&m.FailedCounter, 1)
}

func (m *M) Seen(depth int, u string) bool {
	m.mu.RLock()
	for d := 1; d <= depth; d++ {
		if _, ok := m.Store[d][u]; ok {
			m.mu.RUnlock()
			return true
		}
	}
	m.mu.RUnlock()
	return false
}

func (m *M) String() string {
	sb := strings.Builder{}
	m.mu.RLock()
	cnt := 0
	for i := 1; i < maxDepth; i++ {
		if m.Store[i] != nil {
			cnt += len(m.Store[i])
			for u := range m.Store[i] {
				temp := fmt.Sprintf("%d %s\n", i, u)
				sb.WriteString(temp)
			}
		} else {
			break
		}
	}
	m.mu.RUnlock()
	return fmt.Sprintf("%d (failed %d)\n%s", cnt, m.FailedCounter, sb.String())
}

func NewM() *M {
	return &M{
		Store: map[int]map[string]struct{}{},
		mu:    sync.RWMutex{},
	}
}

var (
	n         = flag.Int("n", 6, "max concurrent requests number")
	root      = flag.String("root", "https://en.wikipedia.org/wiki/Money_Heist", "initial page")
	d         = flag.Int("d", 3, "max recursion depth")
	userAgent = flag.String("user-agent", "", "user agent")
	maxDepth  = 0
)

func main() {
	flag.Parse()
	maxDepth = *d
	var (
		wg    = sync.WaitGroup{}
		sem   = make(chan struct{}, *n)
		store = NewM()
	)
	sem <- struct{}{}
	wg.Add(1)
	go crawler(sem, &wg, store, *root, 1)
	<-sem

	wg.Wait()
	fmt.Print(store)
}

func crawler(sem chan struct{}, wg *sync.WaitGroup, store *M, u string, depth int) {
	defer wg.Done()
	var (
		rootUrl *url.URL
		host    string
		scheme  = "https://"
	)

	rootUrl, err := url.Parse(u)
	if err != nil {
		return
	}
	host = rootUrl.Host

	req, err := http.NewRequest("GET", rootUrl.String(), nil)
	if err != nil {

		return
	}
	req.Header.Set("User-Agent", *userAgent)

	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	if err != nil {
		store.AddFailed()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return
	}

	if !store.Seen(depth, rootUrl.String()) {
		store.Add(depth, rootUrl.String())
	}

	if depth >= maxDepth {
		return
	}

	z := html.NewTokenizer(resp.Body)
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			store.AddFailed()
			if z.Err() != nil {
				return
			}
		}
		if tokenType != html.StartTagToken {
			continue
		}

		tn, _ := z.TagName()
		if string(tn) != "a" {
			continue
		}

		k, v, more := z.TagAttr()
		if string(k) == "href" {
			tempUrl, err := url.Parse(string(v))
			if err != nil {
				continue
			}
			vStr := fmt.Sprintf("%s%s%s", scheme, host, tempUrl.Path)

			if !store.Seen(depth, vStr) {
				sem <- struct{}{}
				wg.Add(1)
				go crawler(sem, wg, store, vStr, depth+1)
				<-sem
			}
		} else {
			for more {
				k, v, more = z.TagAttr()
				if string(k) == "href" {
					tempUrl, err := url.Parse(string(v))
					if err != nil {
						continue
					}
					vStr := fmt.Sprintf("%s%s%s", scheme, host, tempUrl.Path)

					if !store.Seen(depth, vStr) {
						sem <- struct{}{}
						wg.Add(1)
						go crawler(sem, wg, store, vStr, depth+1)
						<-sem
					}
				}
			}

		}
	}
}
