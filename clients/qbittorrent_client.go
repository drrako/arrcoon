package clients

import (
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/autobrr/go-qbittorrent"
)

type QBittorentClient struct {
	qbittorrentClient *qbittorrent.Client
}

func NewQbittorrentClient(config ClientConfig) TorrentClient {
	endpoint, err := url.Parse(config["host"].(string))
	if err != nil {
		log.Fatal(err)
	}
	password, _ := endpoint.User.Password()
	client := qbittorrent.NewClient(qbittorrent.Config{
		Host:     endpoint.Scheme + "://" + endpoint.Host,
		Username: endpoint.User.Username(),
		Password: password,
	})
	err = client.Login()
	if err != nil {
		log.Fatal(err)
	}
	return &QBittorentClient{qbittorrentClient: client}
}

func (qbc QBittorentClient) Test() bool {
	version, err := qbc.qbittorrentClient.GetWebAPIVersion()
	if err != nil {
		log.WithError(err).Error("Couldn't connect to qbittorrent")
		return false
	}
	log.WithFields(log.Fields{
		"Version": version,
	}).Info("Succesfully connected to qbittorrent")
	return true
}

func (qbc QBittorentClient) RemoveTorrents(hashes []string) {
	err := qbc.qbittorrentClient.DeleteTorrents(hashes, true)
	if err != nil {
		log.WithError(err).Error("Error while removing qbittorrent torrents")
	} else {
		log.WithFields(log.Fields{
			"Hashes": hashes,
		}).Info("Successfully removed qbittorrent torrents")
	}
}
