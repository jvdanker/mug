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
	NewUrl
	UpdateDiff
)

type NotificationType int
type NotificationItem struct {
	Type NotificationType
	Id   int
	Data interface{}
}

const (
	ReferenceUpdated NotificationType = iota
	CurrentUpdated
	DiffUpdated
)

type Worker struct {
	c chan WorkItem
	u chan NotificationItem
}

func NewWorker() Worker {
	var work = make(chan WorkItem, 100)
	var updates = make(chan NotificationItem, 100)

	return Worker{
		c: work,
		u: updates,
	}
}

func (w Worker) Worker(ctx context.Context, wg sync.WaitGroup) {
	fmt.Println("Listening for work...")
loop:
	for {
		select {
		case work := <-w.c:
			time.Sleep(1 * time.Second)
			fmt.Println("work received %v", work)

			fs := store.NewFileStore()
			err := fs.Open()
			if err != nil {
				panic(err)
			}

			item, err := fs.Get(work.Url.Id)
			if err != nil {
				panic(err)
			}

			if work.Type == UpdateDiff {
				a := NewApi(w)
				resp, err := a.PDiff(work.Url.Id)
				if err != nil {
					panic(err)
				}

				w.u <- NotificationItem{Type: DiffUpdated, Id: work.Url.Id, Data: resp}

			} else {
				_, thumb, err := CreateScreenshot(item.Url)
				if err != nil {
					panic(err)
				}

				switch work.Type {
				case NewUrl:
					item.Reference = thumb
					w.c <- WorkItem{Type: UpdateCurrent, Url: *item}
					w.u <- NotificationItem{Type: ReferenceUpdated, Id: work.Url.Id, Data: *item}
				case UpdateReference:
					item.Reference = thumb
					w.u <- NotificationItem{Type: ReferenceUpdated, Id: work.Url.Id, Data: *item}
				case UpdateCurrent:
					item.Current = thumb
					w.c <- WorkItem{Type: UpdateDiff, Url: *item}
					w.u <- NotificationItem{Type: CurrentUpdated, Id: work.Url.Id, Data: *item}
				}
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
