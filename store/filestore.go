package store

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sync"
)

type FileStore struct {
	data []Url
}

var lock sync.Mutex

func NewFileStore() FileStore {
	return FileStore{}
}

func (s *FileStore) Open() error {
	lock.Lock()
	defer lock.Unlock()

	f, err := os.Open("data.json")
	if err != nil {
		return err
	}
	defer f.Close()

	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	json.Unmarshal(byteValue, &s.data)

	return nil
}

func (s *FileStore) Close() {
	lock.Lock()
	defer lock.Unlock()

	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		panic(err)
	}

	outfile, err := os.Create("data.json")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	outfile.Write(b)
}

func (s *FileStore) List() []Url {
	lock.Lock()
	defer lock.Unlock()

	return s.data
}

func (s *FileStore) Get(id int) (*Url, error) {
	lock.Lock()
	defer lock.Unlock()

	i := s.indexOf(id)
	if i != -1 {
		return &s.data[i], nil
	}

	return nil, errors.New("Not found")
}

func (s *FileStore) Add(url Url) error {
	lock.Lock()
	defer lock.Unlock()

	s.data = append(s.data, url)

	return nil
}

func (s *FileStore) Delete(id int) error {
	lock.Lock()
	defer lock.Unlock()

	for i, item := range s.data {
		if item.Id == id {
			s.data = append(s.data[:i], s.data[i+1:]...)
			break
		}
	}

	return nil
}

func (s *FileStore) Update(url Url) error {
	lock.Lock()
	defer lock.Unlock()

	i := s.indexOf(url.Id)
	if i != -1 {
		s.data[i] = url
	}

	return errors.New("Not found")
}

func (s *FileStore) indexOf(id int) int {
	for i, item := range s.data {
		if item.Id == id {
			return i
		}
	}

	return -1
}
