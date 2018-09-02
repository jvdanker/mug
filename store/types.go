package store

type StatusType int

const (
	SUCCESS StatusType = iota
	WARNING
	FAIL
)

type Url struct {
	Id        int        `json:"id"`
	Url       string     `json:"url"`
	Reference string     `json:"reference"`
	Current   string     `json:"current"`
	Overlay   string     `json:"overlay"`
	Results   string     `json:"results"`
	Status    StatusType `json:"status"`
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

type HandlerError struct {
	Message string
	Code    int
}

func (h HandlerError) Error() string {
	return h.Message
}
