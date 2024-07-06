package metapipe_client

type Client interface {
	Add(key string) (bool, error)
	Delete(key string) (bool, error)
	Get(key string) ([]byte, error)
	Replace(key string) (bool, error)
	Set(key string, value []byte, ttl int) (bool, error)
	Stale(key string) (bool, error)
	Touch(key string, ttl int) (bool, error)
}
