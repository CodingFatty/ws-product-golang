package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"github.com/didip/tollbooth"
)

type counters struct {
	sync.Mutex
	view  int
	click int
}

var (
	c = counters{}

	content = []string{"sports", "entertainment", "business", "education"}

	// time layout
	timeLayout = "2006-01-02 15:04"
	// counter for each click
	Counters = make(map[string]counters)
	// store storing in memory that get updated every 5 seconds
	Store = make(map[string]counters)
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	c.Lock()
	c.view++
	c.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	c.Lock()
	c.click++
	c.Unlock()

	// counter for each click
	t := time.Now()
	var currentTime = data + ":" + t.Format(timeLayout)
	Counters[currentTime] = c
	fmt.Println(Counters)

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
}

func isAllowed() bool {
	return true
}

func uploadCounters(currentCounter map[string]counters, s chan map[string]counters) {
	// copying value to Store and delete key/value from Counters
	for k, v := range currentCounter {
		// v.Lock()
		Store[k] = v
		// v.Unlock()
		delete(currentCounter, k)
	}
	s <- currentCounter
}

func main() {
	// initial a new multiplxer for limiting request at middleware handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.HandleFunc("/view/", viewHandler)
	mux.HandleFunc("/stats/", statsHandler)

	// repeat uploading every 5 secs
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for  {
			select {
			case <-ticker.C:
				s := make(chan map[string]counters)
				go uploadCounters(Counters,s)
				Counters := <- s
				fmt.Println("uploaded", Counters)
				fmt.Println(Store)
			}
		}
	}()
	// limit 1 request per second by default with no initial number of token
	log.Fatal(http.ListenAndServe(":8080", tollbooth.LimitHandler(tollbooth.NewLimiter(1, nil), mux)))
}
