package handler

import (
	"encoding/json"
	"github.com/jvdanker/mug/api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type HttpHandlers struct {
	stop   chan<- os.Signal
	a      api.Api
	worker api.Worker
}

func NewHandlers(stop chan<- os.Signal, worker api.Worker) HttpHandlers {
	var a = api.NewApi(worker)

	return HttpHandlers{
		stop:   stop,
		a:      a,
		worker: worker,
	}
}

func (h HttpHandlers) AddHandler(pattern string, handler JsonHandler) {
	var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc(pattern, Decorate(
		handler,
		WithJsonHandler(),
		WithLogger(logger),
		WithCors()))
}

func (h HttpHandlers) HandleShutdown(r *http.Request) (interface{}, error) {
	h.stop <- os.Interrupt

	return nil, nil
}

func (h HttpHandlers) HandleListRequests(r *http.Request) (interface{}, error) {
	l, err := h.a.List()
	return l, err
}

func (h HttpHandlers) HandleScanAllRequests(r *http.Request) (interface{}, error) {
	var t struct {
		Type string `json:"type"`
	}
	err := parseBody(r, &t)

	l, err := h.a.ScanAll(t.Type)
	return l, err
}

func (h HttpHandlers) HandleScanRequests(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/url/scan/"):])
	if err != nil {
		return nil, err
	}

	err = h.a.SubmitScanRequest(id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h HttpHandlers) HandleInitRequests(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/init/"):])
	if err != nil {
		return nil, err
	}

	_, err = h.a.Init(id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h HttpHandlers) HandlePDiffRequest(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/pdiff/"):])
	if err != nil {
		return nil, err
	}

	resp, err := h.a.PDiff(id)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (h HttpHandlers) HandleGetReferenceScreenshot(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/reference/get/"):])
	if err != nil {
		return nil, err
	}

	resp, err := h.a.GetReferenceScreenshot(id)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h HttpHandlers) HandleGetScanScreenshot(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/screenshot/scan/"):])
	if err != nil {
		return nil, err
	}

	resp, err := h.a.GetScanScreenshot(id)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h HttpHandlers) HandleAddUrl(r *http.Request) (interface{}, error) {
	var t struct {
		Url string `json:"url"`
	}

	err := parseBody(r, &t)
	if err != nil {
		return nil, err
	}

	resp, err := h.a.AddUrl(t.Url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h HttpHandlers) HandleDeleteUrl(r *http.Request) (interface{}, error) {
	id, err := strconv.Atoi(r.URL.Path[len("/url/delete/"):])
	if err != nil {
		return nil, err
	}

	_, err = h.a.DeleteUrl(id)
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
