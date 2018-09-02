package api

import (
	"encoding/base64"
	"github.com/jvdanker/mug/store"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

type Api interface {
	List() ([]store.Url, error)
	ScanAll(t string)
	SubmitScanRequest(id int) error
	Init(id int) (interface{}, error)
	PDiff(id int) (interface{}, error)
	GetReferenceScreenshot(id int) (interface{}, error)
	GetScanScreenshot(id int) (interface{}, error)
	AddUrl(url string) (interface{}, error)
	DeleteUrl(id int) (interface{}, error)
}

type MugApi struct {
	Api
	worker Worker
}

func NewApi(worker Worker) MugApi {
	return MugApi{
		worker: worker,
	}
}

func (a MugApi) List() ([]store.Url, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	return fs.List(), nil
}

func (a MugApi) ScanAll(t string) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	type Response struct {
		Ids []int `json:"ids"`
	}

	var response Response

	for _, item := range fs.List() {
		switch t {
		case "current":
			response.Ids = append(response.Ids, item.Id)
			a.worker.c <- store.WorkItem{Type: store.Current, Url: item}
		case "reference":
			response.Ids = append(response.Ids, item.Id)
			a.worker.c <- store.WorkItem{Type: store.Reference, Url: item}
		default:
			panic("Unsupported type " + t)
		}
	}

	return response, nil
}

func (a MugApi) SubmitScanRequest(id int) error {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return err
	}

	item, err := fs.Get(id)
	if err != nil {
		return store.HandlerError{"", http.StatusNotFound}
	}

	a.worker.c <- store.WorkItem{Type: store.Current, Url: *item}

	return nil
}

func (a MugApi) Init(id int) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	item, err := fs.Get(id)
	if err != nil {
		return nil, err
	}

	_, data, err := CreateScreenshot(item.Url)
	if err != nil {
		return nil, err
	}

	item.Reference = data
	err = fs.Update(*item)
	if err != nil {
		return nil, err
	}

	fs.Close()

	return item, nil
}

func (a MugApi) PDiff(id int) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	output := ""
	status := true

	for _, item := range fs.List() {
		if item.Id == id {
			if item.Reference == "" || item.Current == "" {
				return nil, store.HandlerError{"Missing reference or current image", http.StatusInternalServerError}
			}

			str1 := item.Reference[len("data::image/png;base64,"):]
			b1, err := base64.StdEncoding.DecodeString(str1)
			if err != nil {
				return nil, err
			}

			err = ioutil.WriteFile("i1.png", b1, 0644)
			if err != nil {
				return nil, err
			}

			str2 := item.Current[len("data::image/png;base64,"):]
			b2, err := base64.StdEncoding.DecodeString(str2)
			if err != nil {
				return nil, err
			}

			err = ioutil.WriteFile("i2.png", b2, 0644)
			if err != nil {
				return nil, err
			}

			dir, err := os.Getwd()
			if err != nil {
				return nil, err
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
				status = cmd.ProcessState.Success()
				output = string(temp)
			} else {
				status = cmd.ProcessState.Success()
				output = string(temp)
			}

			break
		}
	}

	type Response struct {
		Output string `json:"output"`
		Status bool   `json:"status"`
	}

	response := Response{
		Output: output,
		Status: status,
	}

	return response, nil
}

func (a MugApi) GetReferenceScreenshot(id int) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	item, err := fs.Get(id)
	if err != nil || item.Reference == "" {
		return nil, store.HandlerError{"", http.StatusNotFound}
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{
		Data: item.Reference,
	}

	return response, nil
}

func (a MugApi) GetScanScreenshot(id int) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	item, err := fs.Get(id)
	if err != nil || item.Current == "" {
		return nil, store.HandlerError{"", http.StatusNotFound}
	}

	type ScreenshotResponse struct {
		Data string `json:"data"`
	}

	response := ScreenshotResponse{
		Data: item.Current,
	}

	return response, nil
}

func (a MugApi) AddUrl(url string) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	max := 0
	for _, item := range fs.List() {
		if item.Id > max {
			max = item.Id
		}
	}

	u := store.Url{
		Url: url,
		Id:  max + 1,
	}

	err = fs.Add(u)
	if err != nil {
		return nil, err
	}

	fs.Close()

	a.worker.c <- store.WorkItem{Type: store.Reference, Url: u}

	type Response struct {
		Id int `json:"id"`
	}

	return Response{Id: max + 1}, nil
}

func (a MugApi) DeleteUrl(id int) (interface{}, error) {
	fs := store.NewFileStore()
	err := fs.Open()
	if err != nil {
		return nil, err
	}

	err = fs.Delete(id)
	if err != nil {
		return nil, err
	}

	fs.Close()

	return nil, nil
}
