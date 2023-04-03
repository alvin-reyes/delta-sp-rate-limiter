package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type StorageProvider struct {
	ID         uint   `gorm:"primarykey"`
	Miner      string `gorm:"unique"`
	UploadSize int    `gorm:"default:0"`
	Uploads    []int  `gorm:"type:integer[]"`
	Limits     []int  `gorm:"type:integer[]"`
}

var storageProviders = make(map[string]*StorageProvider)
var storageMutex = &sync.Mutex{}
var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("miner_limits.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	db.AutoMigrate(&StorageProvider{})
}

func recordUploadLimit(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	address := r.URL.Query().Get("miner")
	uploadLimit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Invalid upload limit", http.StatusBadRequest)
		return
	}

	// Get or create storage provider
	storageMutex.Lock()
	provider, ok := storageProviders[address]
	if !ok {
		// Check if provider exists in database
		var dbProvider StorageProvider
		if err := db.First(&dbProvider, "miner = ?", address).Error; err == nil {
			provider = &dbProvider
			storageProviders[address] = provider
		} else {
			provider = &StorageProvider{
				Miner:   address,
				Uploads: make([]int, 24),
				Limits:  make([]int, 24),
			}
			storageProviders[address] = provider
			db.Create(provider)
		}
	}
	storageMutex.Unlock()

	// Set the hourly upload limit
	hour := time.Now().Hour()
	provider.Limits[hour] = uploadLimit
	db.Save(provider)

	// Return the updated provider data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

func recordUploadSize(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	address := r.URL.Query().Get("miner")
	uploadSize, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil {
		http.Error(w, "Invalid upload size", http.StatusBadRequest)
		return
	}

	// Get or create storage provider
	storageMutex.Lock()
	provider, ok := storageProviders[address]
	if !ok {
		// Check if provider exists in database
		var dbProvider StorageProvider
		if err := db.First(&dbProvider, "miner = ?", address).Error; err == nil {
			provider = &dbProvider
			storageProviders[address] = provider
		} else {
			provider = &StorageProvider{
				Miner:   address,
				Uploads: make([]int, 24),
				Limits:  make([]int, 24),
			}
			storageProviders[address] = provider
			db.Create(provider)
		}
	}
	storageMutex.Unlock()

	// Add the upload size to the hourly total
	hour := time.Now().Hour()
	provider.Uploads[hour] += uploadSize
	db.Save(provider)

	// Return the updated provider data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

func checkUploadLimit(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	address := r.URL.Query().Get("address")
	// Get the storage provider
	storageMutex.Lock()
	provider, ok := storageProviders[address]
	if !ok {
		// Check if provider exists in database
		var dbProvider StorageProvider
		if err := db.First(&dbProvider, "address = ?", address).Error; err == nil {
			provider = &dbProvider
			storageProviders[address] = provider
		} else {
			http.Error(w, "Storage provider not found", http.StatusNotFound)
			storageMutex.Unlock()
			return
		}
	}
	storageMutex.Unlock()

	// Check if the storage provider is within limits
	hour := time.Now().Hour()
	uploadsThisHour := provider.Uploads[hour]
	limitThisHour := provider.Limits[hour]
	if uploadsThisHour > limitThisHour {
		http.Error(w, "Upload limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Return the provider data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

func main() {
	initDB()
	http.HandleFunc("/record-upload-limit", recordUploadLimit)
	http.HandleFunc("/record-upload-size", recordUploadSize)
	http.HandleFunc("/check-upload-limit", checkUploadLimit)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
