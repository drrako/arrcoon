package clients

type Clients struct {
	Type     string `yaml:"type"`
	Rtorrent struct {
		Host string `yaml:"host"`
	} `yaml:"rtorrent"`
}

type TorrentClient interface {
	Test() bool
	RemoveTorrents(hashes []string)
}

var Instances = map[string]func(clientsConfig Clients) TorrentClient{
	"rtorrent": NewRtorrentClient,
}
