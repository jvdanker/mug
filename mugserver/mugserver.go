package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jvdanker/mug/lib"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type Url struct {
	Id        int    `json:"id"`
	Url       string `json:"url"`
	Reference string `json:"reference"`
	Current   string `json:"current"`
	Overlay   string `json:"overlay"`
}

var stop = make(chan os.Signal, 1)

func main() {
	signal.Notify(stop, os.Interrupt)
	logger := log.New(os.Stdout, "", 0)

	h := &http.Server{Addr: ":8080", Handler: nil}
	http.HandleFunc("/shutdown", handleShutdown)
	http.HandleFunc("/list", handleListRequests)
	http.HandleFunc("/scan/", handleScanRequests)
	http.HandleFunc("/init/", handleInitRequests)
	http.HandleFunc("/merge/", handleMergeRequests)
	http.HandleFunc("/screenshot/get/", handleGetScreenshot)
	http.HandleFunc("/url/add", handleAddUrl)

	go func(h *http.Server) {
		fmt.Println("Press ctrl+c to interrupt...")

		<-stop

		fmt.Println("Shutting down...")
		h.Shutdown(context.Background())
		logger.Println("Server gracefully stopped")
	}(h)

	logger.Printf("Listening on http://0.0.0.0:8080\n")
	if err := h.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}

func handleShutdown(w http.ResponseWriter, r *http.Request) {
	stop <- os.Interrupt
}

func handleListRequests(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	d, err := read()
	if err != nil {
		panic(err)
	}

	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleScanRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/scan/"):])
	fmt.Println(id)

	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := read()
	if err != nil {
		panic(err)
	}

	for i, item := range data {
		if item.Id == id {
			b, err := lib.Run(5*time.Second, item.Url)

			img, _, _ := image.Decode(bytes.NewReader(b))
			image2 := resize.Resize(100, 0, img, resize.NearestNeighbor)

			buf := new(bytes.Buffer)
			err = png.Encode(buf, image2)
			if err != nil {
				panic(err)
			}
			b2 := buf.Bytes()

			data[i].Current = "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2)
			break
		}
	}
	fmt.Println("done")

	saveData(data)

	j, err := json.MarshalIndent("", "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleInitRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/init/"):])
	fmt.Println(id)

	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := read()
	if err != nil {
		panic(err)
	}

	for i, item := range data {
		if item.Id == id {
			b, err := lib.Run(5*time.Second, item.Url)

			img, _, _ := image.Decode(bytes.NewReader(b))
			image2 := resize.Resize(100, 0, img, resize.NearestNeighbor)

			buf := new(bytes.Buffer)
			err = png.Encode(buf, image2)
			if err != nil {
				panic(err)
			}
			b2 := buf.Bytes()

			data[i].Reference = "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2)
			break
		}
	}
	fmt.Println("done")

	saveData(data)

	j, err := json.MarshalIndent("", "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleMergeRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/merge/"):])
	fmt.Println(id)

	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := read()
	if err != nil {
		panic(err)
	}

	for i, item := range data {
		if item.Id == id {
			img1, err := decodeImage(item.Reference)
			if err != nil {
				panic(err)
			}

			img2, err := decodeImage(item.Current)
			if err != nil {
				panic(err)
			}

			b1 := img1.Bounds()
			img3 := image.NewRGBA(b1)

			pixList1 := img1.(*image.RGBA).Pix
			pixList2 := img2.(*image.RGBA).Pix
			for i := 0; i < len(pixList1); i += 4 {
				a1 := float32(pixList1[i+3]) / float32(255)
				a2 := float32(pixList2[i+3]) / 255

				img3.Pix[i] = uint8((float32(pixList1[i]) * a1) + (float32(pixList2[i]) * a2))
				img3.Pix[i+1] = uint8((float32(pixList1[i+1]) * a1) + (float32(pixList2[i+1]) * a2))
				img3.Pix[i+2] = uint8((float32(pixList1[i+2]) * a1) + (float32(pixList2[i+2]) * a2))
				img3.Pix[i+3] = uint8((float32(pixList1[i+3]) * a1) + (float32(pixList2[i+3]) * a2))
			}

			buf := new(bytes.Buffer)
			err = png.Encode(buf, img3)
			if err != nil {
				panic(err)
			}
			b2 := buf.Bytes()

			outfile, err := os.Create("image.png")
			if err != nil {
				panic(err)
			}
			png.Encode(outfile, img3)
			outfile.Close()

			data[i].Overlay = "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2)

			break
		}
	}
	fmt.Println("done")

	saveData(data)

	j, err := json.MarshalIndent("", "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func decodeImage(data string) (image.Image, error) {
	str := data[len("data::image/png;base64,"):]
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(bytes.NewReader(b))

	return img, err
}

func handleGetScreenshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/get/"):])
	fmt.Println(id)

	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{}

	data, err := read()
	if err != nil {
		panic(err)
	}

	for _, item := range data {
		if item.Id == id {
			response.Data = item.Current
		}
	}

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleAddUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)

	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(body))
	var t struct {
		Url string `json:"url"`
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}
	log.Println(t.Url)

	data, err := read()
	if err != nil {
		panic(err)
	}

	max := 0
	for _, item := range data {
		if item.Id > max {
			max = item.Id
		}
	}

	data = append(data, Url{
		Url: t.Url,
		Id:  max + 1,
	})

	saveData(data)

	j, err := json.MarshalIndent("", "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func saveData(data []Url) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}

	outfile, err := os.Create("data.json")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	outfile.Write(b)
}

func read() ([]Url, error) {
	d := []Url{}

	f, err := os.Open("data.json")
	if err != nil {
		return d, nil
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	json.Unmarshal(byteValue, &d)

	return d, nil
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
