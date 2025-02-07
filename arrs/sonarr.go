package arrs

import (
	"arrcoon/clients"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Sonarr struct {
	appDir        string
	torrentClient clients.TorrentClient
	restClient    *resty.Client
	index         Index
}

type SonarrSeriesResponse struct {
	Id int `json:"id"`
}

type SonarrApiResponse struct {
	Current string `json:"current"`
}

type SonarrSeriesEpisodeHistoryResponse struct {
	EpisodeId  int       `json:"episodeId"`
	DownloadId string    `json:"downloadId"`
	Date       time.Time `json:"date"`
	EventType  string    `json:"eventType"`
}

func NewSonarr(appDir string, host string, token string, torrentClient clients.TorrentClient) *Sonarr {
	return &Sonarr{
		appDir:        appDir,
		torrentClient: torrentClient,
		restClient:    resty.New().SetBaseURL(host).SetHeader("X-Api-Key", token),
		index:         *NewIndex("sonarr", appDir),
	}
}
func (s *Sonarr) HandleEvent(event string) {
	switch event {
	case "Test":
		log.Debug("Handling Test event")
		if !s.testApi() {
			os.Exit(1)
		}
		s.index.dropIndex()
		s.buildIndex()
	case "Grab":
		grabbedSeriesId := os.Getenv("sonarr_series_id")
		downloadId := os.Getenv("sonarr_download_id")
		seriesTitle := os.Getenv("sonarr_series_title")
		log.WithFields(log.Fields{
			"sonarr_series_id":    grabbedSeriesId,
			"sonarr_download_id":  downloadId,
			"sonarr_series_title": seriesTitle,
		}).Debug("Handling Grab event")
		seriesId, err := strconv.Atoi(grabbedSeriesId)
		if err != nil {
			log.WithError(err).Error("Failed to convert grabbedSeriesId to int")
			return
		}
		if isValidTorrentHash(downloadId) {
			s.updateIndexFile(seriesId, downloadId)
		}
	case "Download":
		downloadedSeriesId := os.Getenv("sonarr_series_id")
		// Log the event
		log.WithFields(log.Fields{
			"sonarr_series_id": downloadedSeriesId,
		}).Debug("Handling Download event")
		seriesId, err := strconv.Atoi(downloadedSeriesId)
		if err != nil {
			log.WithError(err).Error("Failed to convert sonarr_series_id to int")
			return
		}
		s.removeOutdatedTorrents(seriesId)
	case "SeriesDelete":
		removedSeriesId := os.Getenv("sonarr_series_id")
		seriesId, err := strconv.Atoi(removedSeriesId)
		if err != nil {
			log.WithError(err).Error("Failed to convert sonarr_series_id to int")
			return
		}
		log.WithFields(log.Fields{
			"sonarr_series_id": removedSeriesId,
		}).Debug("Handling SeriesDelete event")
		s.removeAllDownloads(seriesId)
	default:
		log.WithFields(log.Fields{"Event": event}).Debug("Ignoring Sonarr event type")
	}
}

func (s *Sonarr) testApi() bool {
	log.Info("Testing Sonarr accessibility")
	var apiResponse SonarrApiResponse
	_, err := s.restClient.R().SetResult(&apiResponse).Get("api")
	if err != nil {
		log.WithError(err).Error("Couldn't connect to Sonarr API")
		return false
	}
	log.WithFields(log.Fields{
		"Current API Version": apiResponse.Current,
	}).Info("Succesfully connected to Sonarr")
	return true
}

func (s *Sonarr) getSeriesIds() []int {
	params := map[string]string{
		"includeSeasonImages": "false",
	}
	var series []SonarrSeriesResponse
	_, err := s.restClient.R().SetQueryParams(params).SetResult(&series).Get("api/v3/series")
	if err != nil {
		log.WithError(err).Error("Error making request")
		return []int{}
	}
	seriesIds := make([]int, len(series))
	for i, series := range series {
		seriesIds[i] = series.Id
	}
	log.WithFields(log.Fields{
		"Series Ids": seriesIds,
	}).Info()
	return seriesIds
}

func (s *Sonarr) removeOutdatedTorrents(seriesId int) {
	seriesHistory := s.getSeriesHistory(seriesId)

	// Filter in elements where event type is downloadFolderImported and valid torrent hash
	downloadFolderImportedHistory := make([]SonarrSeriesEpisodeHistoryResponse, 0)
	for _, history := range seriesHistory {
		if history.EventType == "downloadFolderImported" && isValidTorrentHash(history.DownloadId) {
			downloadFolderImportedHistory = append(downloadFolderImportedHistory, history)
		}
	}

	// History entries to episode id
	historyMap := make(map[int][]SonarrSeriesEpisodeHistoryResponse)

	for _, history := range downloadFolderImportedHistory {
		if _, ok := historyMap[history.EpisodeId]; !ok {
			historyMap[history.EpisodeId] = []SonarrSeriesEpisodeHistoryResponse{history}
		} else {
			historyMap[history.EpisodeId] = append(historyMap[history.EpisodeId], history)
		}
	}

	// Sort the history entries for each episode ID in descending order by date
	for episodeId, histories := range historyMap {
		sort.Slice(histories, func(i, j int) bool {
			return histories[i].Date.After(histories[j].Date)
		})
		historyMap[episodeId] = histories
	}

	log.WithFields(log.Fields{
		"Download Folder Imported History": historyMap,
	}).Trace()

	var uniqueRelevantValuesMap = make(map[string]bool)
	var uniqueOutdatedValuesMap = make(map[string]bool)

	// Get the first history entry for each episode ID
	for _, histories := range historyMap {
		if len(histories) > 0 {
			uniqueRelevantValuesMap[histories[0].DownloadId] = true
		}
		if len(histories) > 1 {
			for _, history := range histories[1:] {
				uniqueOutdatedValuesMap[history.DownloadId] = true
			}
		}
	}

	// Remove the relevant values from the outdated values
	for key := range uniqueRelevantValuesMap {
		delete(uniqueOutdatedValuesMap, key)
	}

	var oudatedHashValues []string
	for key := range uniqueOutdatedValuesMap {
		oudatedHashValues = append(oudatedHashValues, key)
	}

	log.WithFields(log.Fields{
		"Outdated Hash Values": oudatedHashValues,
	}).Debug()

	s.torrentClient.RemoveTorrents(oudatedHashValues)
}

func (s *Sonarr) getSeriesHistory(seriesId int) []SonarrSeriesEpisodeHistoryResponse {
	params := map[string]string{
		"seriesId":       strconv.Itoa(seriesId),
		"includeSeries":  "false",
		"includeEpisode": "false",
	}
	var seriesHistory []SonarrSeriesEpisodeHistoryResponse
	_, err := s.restClient.R().SetQueryParams(params).SetResult(&seriesHistory).Get("api/v3/history/series")
	if err != nil {
		log.WithError(err).Error("Error making request")
		return []SonarrSeriesEpisodeHistoryResponse{}
	}
	return seriesHistory
}

func (s *Sonarr) getDeduplicatedDownloadIds(seriesId int, downloadIds []string) []string {
	seriesHistory := s.getSeriesHistory(seriesId)
	uniqueRequestedDownloadsMap := make(map[string]struct{})
	var uniqueRequestedDownloadIds []string
	for _, history := range seriesHistory {
		if _, ok := uniqueRequestedDownloadsMap[history.DownloadId]; !ok {
			uniqueRequestedDownloadsMap[history.DownloadId] = struct{}{}
			uniqueRequestedDownloadIds = append(uniqueRequestedDownloadIds, history.DownloadId)
		}
	}
	for _, downloadId := range downloadIds {
		if _, ok := uniqueRequestedDownloadsMap[downloadId]; !ok {
			uniqueRequestedDownloadsMap[downloadId] = struct{}{}
			uniqueRequestedDownloadIds = append(uniqueRequestedDownloadIds, downloadId)
		}
	}

	var validTorrentHashDownloadIds []string
	for _, downloadId := range uniqueRequestedDownloadIds {
		if isValidTorrentHash(downloadId) {
			validTorrentHashDownloadIds = append(validTorrentHashDownloadIds, downloadId)
		}
	}

	log.WithFields(log.Fields{
		"Series Id": seriesId,
		"Hashes":    validTorrentHashDownloadIds,
	}).Debug("Deduplicated download ids")

	return validTorrentHashDownloadIds
}

func (s *Sonarr) removeAllDownloads(seriesId int) {
	indexFile := s.index.readIndexFile(sonarrIndexFileName(seriesId))
	if len(indexFile.Hashes) > 0 {
		s.torrentClient.RemoveTorrents(indexFile.Hashes)
	}
	s.index.removeIndexFile(sonarrIndexFileName(seriesId))
}

func (s *Sonarr) buildIndex() {
	log.Info("Building sonarr series index...")
	seriesIds := s.getSeriesIds()
	var indexedSeriesCounter int
	for _, seriesId := range seriesIds {
		hashes := s.getDeduplicatedDownloadIds(seriesId, nil)
		indexFile := &IndexFile{
			Hashes: hashes,
		}
		if s.index.saveIndexFile("series_"+strconv.Itoa(seriesId), *indexFile) {
			indexedSeriesCounter++
		} else {
			os.Exit(1)
		}
	}
	log.WithFields(log.Fields{
		"Indexed Series": indexedSeriesCounter,
	}).Info("Sonarr index built")
}

func (s *Sonarr) updateIndexFile(seriesId int, downloadId string) {
	hashes := s.getDeduplicatedDownloadIds(seriesId, []string{downloadId})
	indexFile := &IndexFile{
		Hashes: hashes,
	}
	s.index.saveIndexFile(sonarrIndexFileName(seriesId), *indexFile)
}

func sonarrIndexFileName(seriesId int) string {
	return "series_" + strconv.Itoa(seriesId)
}
