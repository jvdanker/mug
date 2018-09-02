package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jvdanker/mug/api"
	"github.com/jvdanker/mug/lib"
	"github.com/jvdanker/mug/store"
	"github.com/nfnt/resize"
	"image"
	"image/png"
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
	"time"
)

type Decorator func(http.HandlerFunc) http.HandlerFunc
type JsonHandler func(*http.Request) (interface{}, error)

var stop = make(chan os.Signal, 1)
var work = make(chan store.WorkItem, 100)
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	// start chrome
	// /opt/google/chrome/chrome --remote-debugging-port=9222 --user-data-dir=remote-profile

	signal.Notify(stop, os.Interrupt)

	h := &http.Server{Addr: ":8080", Handler: nil}
	Handle("/shutdown", handleShutdown)
	Handle("/list", handleListRequests)
	Handle("/init/", handleInitRequests)
	Handle("/merge/", handleMergeRequests)
	Handle("/diff/", handleDiffRequest)
	Handle("/pdiff/", handlePDiffRequest)
	Handle("/scan", handleScanAllRequests)
	Handle("/screenshot/get/", handleGetScreenshot)
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
	go backgroundWorker(ctx, wg)

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

var a = api.NewApi(work)

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

func handleMergeRequests(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/merge/"):])
	if err != nil {
		return nil, err
	}

	fs := store.NewFileStore()
	err = fs.Open()
	if err != nil {
		return nil, err
	}

	for _, item := range fs.List() {
		if item.Id == id {
			img1, err := decodeImage(item.Reference)
			if err != nil {
				return nil, err
			}

			img2, err := decodeImage(item.Current)
			if err != nil {
				return nil, err
			}

			b1 := img1.Bounds()
			img3 := image.NewRGBA(b1)

			pixList1 := img1.(*image.RGBA).Pix
			pixList2 := img2.(*image.RGBA).Pix
			for i := 0; i < len(pixList1); i += 4 {
				a1 := float32(pixList1[i+3]) / float32(255)
				a2 := float32(pixList2[i+3]) / float32(255)

				img3.Pix[i] = uint8((float32(pixList1[i]) * a1) + (float32(pixList2[i]) * a2))
				img3.Pix[i+1] = uint8((float32(pixList1[i+1]) * a1) + (float32(pixList2[i+1]) * a2))
				img3.Pix[i+2] = uint8((float32(pixList1[i+2]) * a1) + (float32(pixList2[i+2]) * a2))
				img3.Pix[i+3] = uint8((float32(pixList1[i+3]) * a1) + (float32(pixList2[i+3]) * a2))
			}

			buf := new(bytes.Buffer)
			err = png.Encode(buf, img3)
			if err != nil {
				return nil, err
			}
			b2 := buf.Bytes()

			outfile, err := os.Create("image.png")
			if err != nil {
				return nil, err
			}
			png.Encode(outfile, img3)
			outfile.Close()

			item.Overlay = "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2)
			break
		}
	}

	fs.Close()

	return nil, nil
}

func handleDiffRequest(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/diff/"):])
	if err != nil {
		return nil, err
	}

	fs := store.NewFileStore()
	err = fs.Open()
	if err != nil {
		return nil, err
	}

	for _, item := range fs.List() {
		if item.Id == id {
			str1 := item.Reference[len("data::image/png;base64,"):]
			b1, err := base64.StdEncoding.DecodeString(str1)
			if err != nil {
				return nil, err
			}

			err = ioutil.WriteFile("i1.png", b1, 0644)
			if err != nil {
				return nil, err
			}

			str2 := item.Reference[len("data::image/png;base64,"):]
			b2, err := base64.StdEncoding.DecodeString(str2)
			if err != nil {
				return nil, err
			}

			err = ioutil.WriteFile("i2.png", b2, 0644)
			if err != nil {
				return nil, err
			}

			// /opt/google/chrome/chrome --remote-debugging-port=9222 --user-data-dir=remote-profile
			cmd := exec.Command("/Users/juan/workspaces/go/src/github.com/jvdanker/mug/pdiff/perceptualdiff",
				"/Users/juan/workspaces/go/src/github.com/jvdanker/mug/i1.png",
				"/Users/juan/workspaces/go/src/github.com/jvdanker/mug/datai2.png",
				"-verbose")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, err
			} else {
				fmt.Println(string(output))
			}
			fmt.Println("state = ", cmd.ProcessState)
			fmt.Println("state = ", cmd.ProcessState.Success())

			break
		}
	}

	fs.Close()

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

func handleGetScreenshot(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/get/"):])
	if err != nil {
		return nil, err
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{}

	fs := store.NewFileStore()
	err = fs.Open()
	if err != nil {
		return nil, err
	}

	found := false
	for _, item := range fs.List() {
		if item.Id == id {
			response.Data = item.Current
			found = true
			break
		}
	}

	if !found {
		return nil, store.HandlerError{"", http.StatusNotFound}
	}

	return response, nil
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

func createScreenshot(url string) (string, string, error) {
	b, err := lib.Run(5*time.Second, url)
	if err != nil {
		return "", "", err
	}

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return "", "", err
	}

	image2 := resize.Resize(100, 0, img, resize.NearestNeighbor)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, image2)
	if err != nil {
		return "", "", err
	}
	b2 := buf.Bytes()

	return "", "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2), nil
}

func decodeImage(data string) (image.Image, error) {
	str := data[len("data::image/png;base64,"):]
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	return img, err
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

func backgroundWorker(ctx context.Context, wg sync.WaitGroup) {
	fmt.Println("Listening for work...")
loop:
	for {
		select {
		case w := <-work:
			time.Sleep(1 * time.Second)
			fmt.Println("work received %v", w)

			fs := store.NewFileStore()
			err := fs.Open()
			if err != nil {
				panic(err)
			}

			item, err := fs.Get(w.Url.Id)
			if err != nil {
				panic(err)
			}

			_, thumb, err := createScreenshot(item.Url)
			if err != nil {
				panic(err)
			}

			switch w.Type {
			case store.Reference:
				item.Reference = thumb
			case store.Current:
				item.Current = thumb
			}

			fs.Close()

		case <-ctx.Done():
			fmt.Println("ctx done")
			break loop
		}
	}
	fmt.Println("Done listening for work...")
	wg.Done()
}
