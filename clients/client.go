package clients

type ClientConfig map[string]interface{}

type TorrentClient interface {
	Test() bool
	RemoveTorrents(hashes []string)
}

var Instances = map[string]func(clientConfig ClientConfig) TorrentClient{
	"rtorrent":     NewRtorrentClient,
	"transmission": NewTransmissionClient,
	"qbittorrent":  NewQbittorrentClient,
}
