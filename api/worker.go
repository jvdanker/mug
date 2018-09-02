package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/jvdanker/mug/lib"
	"github.com/jvdanker/mug/store"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"sync"
	"time"
)

type Worker struct {
	c chan store.WorkItem
}

func NewWorker() Worker {
	var work = make(chan store.WorkItem, 100)

	return Worker{
		c: work,
	}
}

func (w Worker) Worker(ctx context.Context, wg sync.WaitGroup) {
	fmt.Println("Listening for work...")
loop:
	for {
		select {
		case w := <-w.c:
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
