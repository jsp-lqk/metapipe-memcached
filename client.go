package client

type MutationResult int

const (
	Success MutationResult = iota
	Error
	Exists
	NotFound
	NotStored
)

type EntryInfo struct {
	TimeToLive int
	LastAccess int
	CasId int
	Fetched bool
	SlabClassId int
	Size int
}

type Client interface {
	Add(key string) (MutationResult, error)
	Delete(key string) (MutationResult, error)
	Get(key string) ([]byte, error)
	GetMany(keys []string) (map[string][]byte, error)
	Info(key string) (EntryInfo, error)
	Replace(key string) (MutationResult, error)
	Set(key string, value []byte, ttl int) (MutationResult, error)
	Stale(key string) (MutationResult, error)
	Touch(key string, ttl int) (MutationResult, error)
	Shutdown()
}

type MetapipeClient struct {

}

func NewMetapipeClient() {
	
}
