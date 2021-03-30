package kava

// Client implements services.Client interface for communcating with the kava chain
type Client struct {
}

// NewClient initialized a new Kava Client
func NewClient() (*Client, error) {
	return &Client{}, nil
}
