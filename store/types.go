package store

type Url struct {
	Id        int    `json:"id"`
	Url       string `json:"url"`
	Reference string `json:"reference"`
	Current   string `json:"current"`
	Overlay   string `json:"overlay"`
}

type Store interface {
	Open() error
	Close()

	List() []Url
	Get(id int) (*Url, error)
	Update(url Url) error
	Add(url Url) error
	Delete(id int) error
}

type WorkType int
type WorkItem struct {
	Type WorkType
	Url  Url
}

const (
	Reference WorkType = iota
	Current
)

type HandlerError struct {
	Message string
	Code    int
}

func (h HandlerError) Error() string {
	return h.Message
}
