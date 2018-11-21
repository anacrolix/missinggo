package conntrack

type Protocol = string

type Endpoint = string

type Entry struct {
	Protocol
	LocalAddr  Endpoint
	RemoteAddr Endpoint
}
