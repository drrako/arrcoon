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

type Radarr struct {
	appDir        string
	torrentClient clients.TorrentClient
	restClient    *resty.Client
	index         Index
}

type RadarrApiResponse struct {
	Current string `json:"current"`
}

type RadarrMoviesResponse struct {
	Id int `json:"id"`
}

type RadarrMoviesHistoryResponse struct {
	MovieId    int       `json:"movieId"`
	DownloadId string    `json:"downloadId"`
	Date       time.Time `json:"date"`
	EventType  string    `json:"eventType"`
}

func NewRadarr(appDir string, host string, token string, torrentClient clients.TorrentClient) *Radarr {
	return &Radarr{
		appDir:        appDir,
		torrentClient: torrentClient,
		restClient:    resty.New().SetBaseURL(host).SetHeader("X-Api-Key", token),
		index:         *NewIndex("radarr", appDir),
	}
}

func (r *Radarr) HandleEvent(event string) {
	switch event {
	case "Test":
		log.Debug("Handling Test event")
		r.testApi()
		r.index.dropIndex()
		r.buildIndex()
	case "Grab":
		grabbedMovieId := os.Getenv("radarr_movie_id")
		downloadId := os.Getenv("radarr_download_id")
		movieTitle := os.Getenv("radarr_movie_title")
		log.WithFields(log.Fields{
			"radarr_movie_id":    grabbedMovieId,
			"radarr_download_id": downloadId,
			"radarr_movie_title": movieTitle,
		}).Debug("Handling Grab event")
		movieId, err := strconv.Atoi(grabbedMovieId)
		if err != nil {
			log.WithError(err).Error("Failed to convert radarr_movie_id to int")
			return
		}
		if isValidTorrentHash(downloadId) {
			r.updateIndexFile(movieId, downloadId)
		}
	case "Download":
		downloadedMovieId := os.Getenv("radarr_movie_id")
		// Log the event
		log.WithFields(log.Fields{
			"radarr_movie_id": downloadedMovieId,
		}).Debug("Handling Download event")
		movieId, err := strconv.Atoi(downloadedMovieId)
		if err != nil {
			log.WithError(err).Error("Failed to convert radarr_movie_id to int")
			return
		}
		r.removeOutdatedTorrents(movieId)
	case "MovieDelete":
		removedMovieId := os.Getenv("radarr_movie_id")
		movieId, err := strconv.Atoi(removedMovieId)
		if err != nil {
			log.WithError(err).Error("Failed to convert radarr_movie_id to int")
		}
		log.WithFields(log.Fields{
			"radarr_movie_id": removedMovieId,
		}).Debug("Handling MovieDelete event")
		r.removeAllDownloads(movieId)
	default:
		log.WithField("event", event).Info("Ignoring Radarr event type")
	}
}

func (r *Radarr) testApi() {
	log.Info("Testing Radarr API")
	var apiResponse RadarrApiResponse
	_, err := r.restClient.R().SetResult(&apiResponse).Get("api")
	if err != nil {
		log.WithError(err).Error("Couldn't connect to Radarr API")
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"Current API Version": apiResponse.Current,
	}).Info("Succesfully connected to Radarr")
}

func (r *Radarr) getMovies() []int {
	params := map[string]string{
		"excludeLocalCovers": "true",
	}
	var movies []RadarrMoviesResponse
	_, err := r.restClient.R().SetQueryParams(params).SetResult(&movies).Get("api/v3/movie")
	if err != nil {
		log.WithError(err).Error("Error making request")
		return []int{}
	}
	moviesIds := make([]int, len(movies))
	for i, movie := range movies {
		moviesIds[i] = movie.Id
	}
	log.WithFields(log.Fields{
		"Movies Ids": moviesIds,
	}).Info()
	return moviesIds
}

func (r *Radarr) getMovieHistory(movieId int) []RadarrMoviesHistoryResponse {
	params := map[string]string{
		"movieId":      strconv.Itoa(movieId),
		"includeMovie": "false",
	}
	var moviesHistory []RadarrMoviesHistoryResponse
	_, err := r.restClient.R().SetQueryParams(params).SetResult(&moviesHistory).Get("api/v3/history/movie")
	if err != nil {
		log.WithError(err).Error("Error making request")
		return []RadarrMoviesHistoryResponse{}
	}
	return moviesHistory
}

func (r *Radarr) getDeduplicatedDownloadIds(movies int, downloadIds []string) []string {
	moviesHistory := r.getMovieHistory(movies)
	uniqueRequestedDownloadsMap := make(map[string]struct{})
	var uniqueRequestedDownloadIds []string
	for _, history := range moviesHistory {
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

	// filter values that match isValidTorrentHash function
	var validTorrentHashDownloadIds []string
	for _, downloadId := range uniqueRequestedDownloadIds {
		if isValidTorrentHash(downloadId) {
			validTorrentHashDownloadIds = append(validTorrentHashDownloadIds, downloadId)
		}
	}
	log.WithFields(log.Fields{
		"Movies Id": movies,
		"Hashes":    validTorrentHashDownloadIds,
	}).Debug("Deduplicated download ids")

	return validTorrentHashDownloadIds
}

func (r *Radarr) buildIndex() {
	log.Info("Building radarr mvies index...")
	moviesIds := r.getMovies()
	var indexedMoviesCounter int
	for _, moviesId := range moviesIds {
		hashes := r.getDeduplicatedDownloadIds(moviesId, nil)
		indexFile := &IndexFile{
			Hashes: hashes,
		}
		if r.index.saveIndexFile("movie_"+strconv.Itoa(moviesId), *indexFile) {
			indexedMoviesCounter++
		} else {
			os.Exit(1)
		}
	}
	log.WithFields(log.Fields{
		"Indexed Movies": indexedMoviesCounter,
	}).Info("Radarr index built")
}

func (r *Radarr) removeOutdatedTorrents(movieId int) {
	movieHistory := r.getMovieHistory(movieId)

	// Filter in elements where event type is downloadFolderImported and valid torrent hash
	downloadFolderImportedHistory := make([]RadarrMoviesHistoryResponse, 0)
	for _, history := range movieHistory {
		if history.EventType == "downloadFolderImported" && isValidTorrentHash(history.DownloadId) {
			downloadFolderImportedHistory = append(downloadFolderImportedHistory, history)
		}
	}

	// Order imported history by date descending
	sort.Slice(downloadFolderImportedHistory, func(i, j int) bool {
		return downloadFolderImportedHistory[i].Date.After(downloadFolderImportedHistory[j].Date)
	})

	log.WithFields(log.Fields{
		"Download Folder Imported History": downloadFolderImportedHistory,
	}).Trace()

	if len(downloadFolderImportedHistory) < 2 {
		log.Debug("Nothing to remove, ignore")
		return
	}

	outdatedHashValues := make([]string, len(downloadFolderImportedHistory)-1)
	for i := 1; i < len(downloadFolderImportedHistory); i++ {
		outdatedHashValues[i-1] = downloadFolderImportedHistory[i].DownloadId
	}
	log.WithFields(log.Fields{
		"Outdated Hash Values": outdatedHashValues,
	}).Debug()

	r.torrentClient.RemoveTorrents(outdatedHashValues)
}

func (r *Radarr) updateIndexFile(movieId int, downloadId string) {
	hashes := r.getDeduplicatedDownloadIds(movieId, []string{downloadId})
	indexFile := &IndexFile{
		Hashes: hashes,
	}
	r.index.saveIndexFile(radarrIndexFileName(movieId), *indexFile)
}

func (r *Radarr) removeAllDownloads(movieId int) {
	indexFile := r.index.readIndexFile(radarrIndexFileName(movieId))
	if len(indexFile.Hashes) > 0 {
		r.torrentClient.RemoveTorrents(indexFile.Hashes)
	}
	r.index.removeIndexFile(radarrIndexFileName(movieId))
}

func radarrIndexFileName(movieId int) string {
	return "movie_" + strconv.Itoa(movieId)
}
