package clients

import (
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/kolo/xmlrpc"
	log "github.com/sirupsen/logrus"
)

type RtorrentClient struct {
	xmlrpcClient *xmlrpc.Client
}

func NewRtorrentClient(config ClientConfig) TorrentClient {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	transport := &http.Transport{
		Proxy: defaultTransport.Proxy,
		DialContext: (&net.Dialer{
			Timeout:   4 * time.Second,
			KeepAlive: 4 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     defaultTransport.ForceAttemptHTTP2,
		MaxIdleConns:          defaultTransport.MaxIdleConns,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
	}
	xmlrpcClient, err := xmlrpc.NewClient(config["host"].(string), transport)
	if err != nil {
		log.Error("Coulnd't initialize rtorrent client")
		return nil
	}
	return &RtorrentClient{xmlrpcClient: xmlrpcClient}
}

func (rc RtorrentClient) Test() bool {
	log.Info("Testing Rtorrent accessibility")
	var listedMethods []string
	err := rc.xmlrpcClient.Call("system.listMethods", []any{}, &listedMethods)
	if err != nil {
		log.WithError(err).Error("Couldn't connect to rTorrent")
		return false
	}
	if slices.Contains(listedMethods, "d.erase") {
		log.Info("Succesfully connected to rTorrent client")
		return true
	}
	log.Error("Unsupported rTorrent instance")
	return false
}

func (rc RtorrentClient) RemoveTorrents(hashes []string) {
	if len(hashes) == 0 {
		return
	}
	log.WithFields(log.Fields{
		"Hashes": hashes,
	}).Info("Requesting torrent files removal")
	for _, hash := range hashes {
		for attempt := 1; attempt <= 3; attempt++ {
			var response any
			deleteParams := []map[string]any{
				{
					"methodName": "d.custom5.set",
					"params":     []any{hash, "1"},
				},
				{
					"methodName": "d.delete_tied",
					"params":     []any{hash},
				},
				{
					"methodName": "d.erase",
					"params":     []any{hash},
				},
			}
			err := rc.xmlrpcClient.Call("system.multicall", deleteParams, &response)

			if err != nil {
				log.WithError(err).Error("Couldn't run erase call")
				time.Sleep(time.Second)
				continue
			}

			responseSlice, ok := response.([]any)
			if !ok {
				log.WithFields(log.Fields{
					"Error responses": response,
					"Hash":            hash,
				}).Error("Unknown response type")
				time.Sleep(time.Second)
				continue
			}

			var successResponses []interface{}
			var errorResponses []interface{}

			for _, responseMap := range responseSlice {
				if responseMapSlice, ok := responseMap.([]interface{}); ok {
					successResponse, ok := responseMapSlice[0].(string)
					if ok {
						successResponses = append(successResponses, successResponse)
					}
				} else if responseMapMap, ok := responseMap.(map[string]interface{}); ok {
					errorResponses = append(errorResponses, responseMapMap)
				}
			}

			if len(errorResponses) > 0 {
				log.WithFields(log.Fields{
					"Attempt":         attempt,
					"Error responses": errorResponses,
					"Hash":            hash,
				}).Debug("Couldn't remove torrents")
				time.Sleep(time.Second)
				continue
			}

			if len(successResponses) > 0 {
				log.WithFields(log.Fields{
					"Hash":    hash,
					"Attempt": attempt,
				}).Info("Torrent has been removed")
				break
			}
		}
	}
}
