package arrs

import (
	"encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type IndexFile struct {
	Hashes []string `json:"hashes"`
}

type Index struct {
	name string
	path string
}

func NewIndex(name string, path string) *Index {
	return &Index{
		name: name,
		path: path,
	}
}

func (i *Index) saveIndexFile(name string, indexFile IndexFile) bool {
	indexFilePath := filepath.Join(i.indexPath(), name+".json")
	// Ensure the sonarr indexes directory exists
	logDir := filepath.Dir(indexFilePath)
	err := os.MkdirAll(logDir, os.ModePerm)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"File Path": indexFilePath,
		}).Error("Failed to create a directory for index file")
		return false
	}

	jsonBytes, err := json.Marshal(indexFile)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"File Path": indexFilePath,
		}).Error("Error marshaling JSON")
		return false
	}

	err = os.WriteFile(indexFilePath, jsonBytes, 0644)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"File Path": indexFilePath,
		}).Error("Error writing to file")
		return false
	}
	return true
}

func (i *Index) readIndexFile(name string) IndexFile {
	indexFilePath := filepath.Join(i.indexPath(), name+".json")
	jsonBytes, err := os.ReadFile(indexFilePath)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"File Path": indexFilePath,
		}).Error("Error reading file")
		return IndexFile{}
	}
	var indexFile IndexFile
	err = json.Unmarshal(jsonBytes, &indexFile)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"File Path": indexFilePath,
		}).Error("Error unmarshaling JSON")
		return IndexFile{}
	}
	return indexFile
}

func (i *Index) removeIndexFile(name string) {
	indexFilePath := filepath.Join(i.indexPath(), name+".json")
	err := os.Remove(indexFilePath)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"File Path": indexFilePath,
		}).Error("Error removing index file")
	}
}

func (i *Index) indexPath() string {
	return filepath.Join(i.path, ".index", i.name)
}

func (i *Index) dropIndex() {
	indexPath := i.indexPath()
	err := os.RemoveAll(indexPath)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Index Path": indexPath,
		}).Error("Error removing index directory")
		return
	}
	log.WithFields(log.Fields{
		"Index Path": indexPath,
	}).Info("Index dropped")
}
