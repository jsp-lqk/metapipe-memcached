package internal

type RawClient interface {
	Dispatch(r []byte) <-chan Response
}

type Client interface {
	Add(key string) (bool, error)
	Delete(key string) (bool, error)
	Get(key string) ([]byte, error)
	Replace(key string) (bool, error)
	Set(key string, value []byte, ttl int) (bool, error)
	Stale(key string) (bool, error)
}

type RequestType int

const (
	ARITHMETIC RequestType = iota
	DEBUG
	DELETE
	GET
	NOOP
	SET
)

type Request struct {
	responseChannel chan Response
}

type Response struct {
	Header []string
	Value []byte
	Error error
}

type ConnectionTarget struct {
	address    string
	port       int
	maxConcurrent int
}
