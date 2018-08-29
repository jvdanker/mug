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
	"net/http/httputil"
	"os"
	"os/exec"
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
	// start chrome
	// /opt/google/chrome/chrome --remote-debugging-port=9222 --user-data-dir=remote-profile

	signal.Notify(stop, os.Interrupt)
	logger := log.New(os.Stdout, "", 0)

	h := &http.Server{Addr: ":8080", Handler: nil}
	http.HandleFunc("/shutdown", handleShutdown)
	http.HandleFunc("/list", handleListRequests)
	http.HandleFunc("/scan/", handleScanRequests)
	http.HandleFunc("/init/", handleInitRequests)
	http.HandleFunc("/merge/", handleMergeRequests)
	http.HandleFunc("/diff/", handleDiffRequest)
	http.HandleFunc("/pdiff/", handlePDiffRequest)
	http.HandleFunc("/screenshot/get/", handleGetScreenshot)
	http.HandleFunc("/screenshot/reference/get/", handleGetReferenceScreenshot)
	http.HandleFunc("/url/add", handleAddUrl)
	http.HandleFunc("/url/delete/", handleDeleteUrl)

	go func(h *http.Server) {
		fmt.Println("Press ctrl+c to interrupt...")

		<-stop

		fmt.Println("Shutting down...")
		h.Shutdown(context.Background())
		logger.Println("Server gracefully stopped")
	}(h)

	// start chrome
	go func() {
		cmd := exec.Command("/opt/google/chrome/chrome",
			"--remote-debugging-port=9222",
			"--disable-extensions",
			"--disable-default-apps",
			"--disable-sync",
			"--hide-scrollbars",
			"--incognito",
			//"--kiosk",
			"--window-size=800,600",
			"--user-data-dir=remote-profile")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return
		} else {
			fmt.Println(string(output))
		}
	}()

	logger.Printf("Listening on http://0.0.0.0:8080\n")
	if err := h.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}

func handleShutdown(w http.ResponseWriter, r *http.Request) {
	stop <- os.Interrupt
}

func handleListRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		setupResponse(w, r)
		return
	}

	d, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, err := json.Marshal(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setupResponse(w, r)
	w.Write(j)
}

func handleScanRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/scan/"):])
	fmt.Println(id)

	if r.Method == "OPTIONS" {
		setupResponse(w, r)
		return
	}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{}

	for i, item := range data {
		if item.Id == id {
			b, err := lib.Run(5*time.Second, item.Url)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			img, _, err := image.Decode(bytes.NewReader(b))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			image2 := resize.Resize(100, 0, img, resize.NearestNeighbor)

			buf := new(bytes.Buffer)
			err = png.Encode(buf, image2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b2 := buf.Bytes()

			data[i].Current = "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2)
			response = ScreenshotResponse{Data: data[i].Current}

			break
		}
	}
	fmt.Println("done")

	saveData(data)

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setupResponse(w, r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleInitRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/init/"):])
	fmt.Println(id)

	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, item := range data {
		if item.Id == id {
			b, err := lib.Run(5*time.Second, item.Url)

			img, _, err := image.Decode(bytes.NewReader(b))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			image2 := resize.Resize(100, 0, img, resize.NearestNeighbor)

			buf := new(bytes.Buffer)
			err = png.Encode(buf, image2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleMergeRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/merge/"):])
	fmt.Println(id)

	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, item := range data {
		if item.Id == id {
			img1, err := decodeImage(item.Reference)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			img2, err := decodeImage(item.Current)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b2 := buf.Bytes()

			outfile, err := os.Create("image.png")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleDiffRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/diff/"):])
	fmt.Println(id)

	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range data {
		if item.Id == id {
			str1 := item.Reference[len("data::image/png;base64,"):]
			b1, err := base64.StdEncoding.DecodeString(str1)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = ioutil.WriteFile("i1.png", b1, 0644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			str2 := item.Reference[len("data::image/png;base64,"):]
			b2, err := base64.StdEncoding.DecodeString(str2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = ioutil.WriteFile("i2.png", b2, 0644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// /opt/google/chrome/chrome --remote-debugging-port=9222 --user-data-dir=remote-profile
			cmd := exec.Command("/Users/juan/workspaces/go/src/github.com/jvdanker/mug/pdiff/perceptualdiff",
				"/Users/juan/workspaces/go/src/github.com/jvdanker/mug/i1.png",
				"/Users/juan/workspaces/go/src/github.com/jvdanker/mug/datai2.png",
				"-verbose")
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println(fmt.Sprint(err) + ": " + string(output))
				return
			} else {
				fmt.Println(string(output))
			}
			fmt.Println("state = ", cmd.ProcessState)
			fmt.Println("state = ", cmd.ProcessState.Success())

			break
		}
	}
	fmt.Println("done")

	saveData(data)

	j, err := json.MarshalIndent("", "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handlePDiffRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/pdiff/"):])
	fmt.Println(id)

	if r.Method == "OPTIONS" {
		setupResponse(w, r)
		return
	}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output := ""
	status := true

	for _, item := range data {
		if item.Id == id {
			if item.Reference == "" || item.Current == "" {
				http.Error(w, "Missing reference or current image", http.StatusInternalServerError)
				return
			}

			str1 := item.Reference[len("data::image/png;base64,"):]
			b1, err := base64.StdEncoding.DecodeString(str1)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = ioutil.WriteFile("i1.png", b1, 0644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			str2 := item.Current[len("data::image/png;base64,"):]
			b2, err := base64.StdEncoding.DecodeString(str2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = ioutil.WriteFile("i2.png", b2, 0644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			dir, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			cmd := exec.Command("docker",
				"run",
				"--rm",
				"-v",
				dir+":/images",
				"jvdanker/pdiff",
				"-verbose",
				"i1.png",
				"i2.png")

			temp, err := cmd.CombinedOutput()
			if err != nil {
				status =  cmd.ProcessState.Success()
				output = string(temp)
			} else {
				status =  cmd.ProcessState.Success()
				output = string(temp)
			}

			break
		}
	}
	fmt.Println("done")

	type Response struct {
		Output string `json:"output"`
		Status bool `json:"status"`
	}

	response := Response{
		Output: output,
		Status: status,
	}

	j, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setupResponse(w, r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleGetScreenshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/get/"):])
	fmt.Println(id)

	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	found := false
	for _, item := range data {
		if item.Id == id {
			response.Data = item.Current
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleGetReferenceScreenshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/reference/get/"):])
	fmt.Println(id)

	if r.Method == "OPTIONS" {
		setupResponse(w, r)
		return
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	index := -1
	var found Url
	for i, item := range data {
		if item.Id == id {
			response.Data = item.Reference
			index = i
			found = data[i]
			break
		}
	}

	if index != -1 && found.Reference == "" {
		_, thumb, err := createScreenshot(found.Url)
		if err != nil {
			panic(err)
		}

		data[index].Reference = thumb
		saveData(data)

		response.Data = thumb
	}

	if index == -1 || response.Data == "" {
		setupResponse(w, r)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setupResponse(w, r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleAddUrl(w http.ResponseWriter, r *http.Request) {
	req, _ := httputil.DumpRequest(r, true)
	fmt.Printf("%s\n", string(req))

	if r.Method == "OPTIONS" {
		setupResponse(w, r)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(string(body))
	var t struct {
		Url string `json:"url"`
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(t.Url)

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	max := 0
	for _, item := range data {
		if item.Id > max {
			max = item.Id
		}
	}

	//_, thumb, err := createScreenshot(t.Url)

	u := Url{
		Url: t.Url,
		Id:  max + 1,
		//Reference: thumb,
	}
	data = append(data, u)

	saveData(data)

	type Response struct {
		Id int `json:"id"`
	}

	response := Response{Id: max + 1}
	j, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setupResponse(w, r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleDeleteUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	id, err := strconv.Atoi(r.URL.Path[len("/url/delete/"):])
	fmt.Println(id)

	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{}

	data, err := read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, item := range data {
		if item.Id == id {
			data = append(data[:i], data[i+1:]...)
			break
		}
	}

	saveData(data)

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
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

func setupResponse(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Content-Type", "application/json")
}
