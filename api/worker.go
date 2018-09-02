package api

import (
	"context"
	"fmt"
	"github.com/jvdanker/mug/store"
	"sync"
	"time"
)

type WorkType int

type WorkItem struct {
	Type WorkType
	Url  store.Url
}

const (
	UpdateReference WorkType = iota
	UpdateCurrent
)

type Worker struct {
	c chan WorkItem
}

func NewWorker() Worker {
	var work = make(chan WorkItem, 100)

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

			_, thumb, err := CreateScreenshot(item.Url)
			if err != nil {
				panic(err)
			}

			switch w.Type {
			case UpdateReference:
				item.Reference = thumb
			case UpdateCurrent:
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
