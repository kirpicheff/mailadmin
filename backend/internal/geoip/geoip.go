package geoip

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

var (
	db     *geoip2.Reader
	mu     sync.RWMutex
	dbPath = "GeoLite2-Country.mmdb"
	dbURL  = "https://raw.githubusercontent.com/P3TERX/GeoLite.mmdb/download/GeoLite2-Country.mmdb"
)

// InitGeoIP запускает фоновую загрузку базы, если её нет, и открывает её.
func InitGeoIP(dataDir string) {
	if dataDir != "" {
		dbPath = filepath.Join(dataDir, "GeoLite2-Country.mmdb")
	}

	// Попытка открыть базу, если она уже есть
	if _, err := os.Stat(dbPath); err == nil {
		openDB()
	} else {
		log.Println("[GeoIP] База не найдена. Начинается фоновая загрузка...")
		go func() {
			if err := downloadDB(); err != nil {
				log.Printf("[GeoIP] Ошибка загрузки базы: %v\n", err)
				return
			}
			log.Println("[GeoIP] База успешно скачана.")
			openDB()
		}()
	}
}

func downloadDB() error {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	tmpPath := dbPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(dbURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close()

	return os.Rename(tmpPath, dbPath)
}

func openDB() {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		db.Close()
	}

	var err error
	db, err = geoip2.Open(dbPath)
	if err != nil {
		log.Printf("[GeoIP] Ошибка открытия базы %s: %v\n", dbPath, err)
	} else {
		log.Println("[GeoIP] База GeoIP успешно загружена.")
	}
}

// GetCountryCode возвращает код страны (например, RU, US) или пустую строку
func GetCountryCode(ipStr string) string {
	mu.RLock()
	defer mu.RUnlock()

	if db == nil {
		return ""
	}

	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return ""
	}

	record, err := db.Country(ip)
	if err != nil {
		return ""
	}

	return record.Country.IsoCode
}
