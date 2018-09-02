package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jvdanker/mug/api"
	"github.com/jvdanker/mug/store"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
)

type Decorator func(http.HandlerFunc) http.HandlerFunc
type JsonHandler func(*http.Request) (interface{}, error)

var stop = make(chan os.Signal, 1)
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
var worker = api.NewWorker()
var a = api.NewApi(worker)

func main() {
	// start chrome
	// /opt/google/chrome/chrome --remote-debugging-port=9222 --user-data-dir=remote-profile

	signal.Notify(stop, os.Interrupt)

	h := &http.Server{Addr: ":8080", Handler: nil}
	Handle("/shutdown", handleShutdown)
	Handle("/list", handleListRequests)
	Handle("/init/", handleInitRequests)
	Handle("/pdiff/", handlePDiffRequest)
	Handle("/scan", handleScanAllRequests)
	Handle("/screenshot/reference/get/", handleGetReferenceScreenshot)
	Handle("/screenshot/scan/", handleGetScanScreenshot)
	Handle("/url/add", handleAddUrl)
	Handle("/url/delete/", handleDeleteUrl)
	Handle("/url/scan/", handleScanRequests)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(h *http.Server, cancel context.CancelFunc) {
		fmt.Println("Press ctrl+c to interrupt...")

		<-stop

		fmt.Println("Shutting down...")
		h.Shutdown(context.Background())
		cancel()
		logger.Println("Server gracefully stopped")
	}(h, cancel)

	var wg sync.WaitGroup
	wg.Add(1)

	//go startChrome()
	go worker.Worker(ctx, wg)

	logger.Printf("Listening on http://0.0.0.0:8080\n")
	if err := h.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}

	wg.Wait()
}

func Handle(pattern string, handler JsonHandler) {
	http.HandleFunc(pattern, Decorate(
		handler,
		WithJsonHandler(),
		WithLogger(logger),
		WithCors()))
}

func Decorate(h JsonHandler, decorators ...Decorator) http.HandlerFunc {
	var handler = func(w http.ResponseWriter, r *http.Request) {
		data, err := h(r)
		if err != nil {
			he, ok := err.(store.HandlerError)
			if ok {
				http.Error(w, err.Error(), he.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if data == nil {
			var s struct{}
			data = s
		}

		j, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}

	for _, d := range decorators {
		handler = d(handler)
	}

	return handler
}

func WithJsonHandler() Decorator {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			h.ServeHTTP(w, r)
		})
	}
}

func WithLogger(l *log.Logger) Decorator {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Println(r.Method, r.URL.Path)
			//req, _ := httputil.DumpRequest(r, true)
			//fmt.Printf("%s\n", string(req))

			h.ServeHTTP(w, r)
		})
	}
}

func WithCors() Decorator {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

			if r.Method == "OPTIONS" {
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// *********************************************************************************

func handleShutdown(r *http.Request) (interface{}, error) {
	stop <- os.Interrupt

	return nil, nil
}

func handleListRequests(r *http.Request) (interface{}, error) {
	l, err := a.List()
	return l, err
}

func handleScanAllRequests(r *http.Request) (interface{}, error) {
	var t struct {
		Type string `json:"type"`
	}
	err := parseBody(r, &t)

	l, err := a.ScanAll(t.Type)
	return l, err
}

func handleScanRequests(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/url/scan/"):])
	if err != nil {
		return nil, err
	}

	err = a.SubmitScanRequest(id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func handleInitRequests(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/init/"):])
	if err != nil {
		return nil, err
	}

	_, err = a.Init(id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func handlePDiffRequest(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/pdiff/"):])
	if err != nil {
		return nil, err
	}

	resp, err := a.PDiff(id)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func handleGetReferenceScreenshot(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/reference/get/"):])
	if err != nil {
		return nil, err
	}

	resp, err := a.GetReferenceScreenshot(id)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func handleGetScanScreenshot(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/scan/"):])
	if err != nil {
		return nil, err
	}

	resp, err := a.GetScanScreenshot(id)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func handleAddUrl(r *http.Request) (interface{}, error) {
	var t struct {
		Url string `json:"url"`
	}

	err := parseBody(r, &t)
	if err != nil {
		return nil, err
	}

	resp, err := a.AddUrl(t.Url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func handleDeleteUrl(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/url/delete/"):])
	if err != nil {
		return nil, err
	}

	_, err = a.DeleteUrl(id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// *********************************************************************************

func parseBody(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return err
	}

	return nil
}

func startChrome() {
	switch runtime.GOOS {
	case "linux":
		path := "/opt/google/chrome/chrome"
		cmd := exec.Command(path,
			"--remote-debugging-port=9222",
			"--disable-extensions",
			"--disable-default-apps",
			"--disable-sync",
			"--hide-scrollbars",
			"--incognito",
			"--window-size=800,600",
			"--user-data-dir=remote-profile")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return
		} else {
			fmt.Println(string(output))
		}
	case "darwin":
		path := "open"
		cmd := exec.Command(path,
			"-n",
			"-a",
			"Google Chrome",
			"--args",
			"--remote-debugging-port=9222",
			"--disable-extensions",
			"--disable-default-apps",
			"--disable-sync",
			"--hide-scrollbars",
			"--incognito",
			"--window-size=800,600",
			"--user-data-dir=/tmp/Chrome Alt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return
		} else {
			fmt.Println(string(output))
		}
	default:
		panic("Unsupported operating system " + runtime.GOOS)
	}
}
