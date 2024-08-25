package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	// Networking
	"github.com/mileusna/useragent"

	// Logging
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	// Packages
	"github.com/ftp27/go-universal-redirect/pkg/analytics"
	"github.com/ftp27/go-universal-redirect/pkg/database"
)

var (
	port          string
	defaultLink   string
	platformLinks map[string]string

	meta    *database.Metadata
	tracker *analytics.InfluxAnalytics
)

func main() {
	setuoLog()
	setupConfig()
	setupMeta()
	setupAnalytics()

	startServer()
}

func setupMeta() {
	host, isExists := os.LookupEnv("REDIS_URL")
	if !isExists {
		log.Fatal().Str("env", "REDIS_URL").Msg("Not found")
		return
	}
	lifetimeStr, isExists := os.LookupEnv("META_LIFETIME")
	lifetime := time.Hour
	if isExists {
		intLifetime, err := strconv.Atoi(lifetimeStr)
		if err != nil {
			log.Fatal().Err(err).Str("env", "META_LIFETIME").Msg("Failed to parse")
			return
		}
		lifetime = time.Duration(intLifetime) * time.Second
	}

	metadata, err := database.NewMetadata(host, lifetime)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create metadata")
		return
	}
	meta = metadata
}

func setupAnalytics() {
	id, isExists := os.LookupEnv("INFLUX_DATABASE")
	if isExists {
		tracker = analytics.NewInfluxAnalytics(id)
	}
}

func setuoLog() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func setupConfig() {
	serverPort, isExists := os.LookupEnv("PORT")
	if !isExists {
		port = ":8080"
	} else {
		port = ":" + serverPort
	}
	link, isExists := os.LookupEnv("LINK_DEFAULT")
	if !isExists {
		log.Fatal().Str("env", "LINK_DEFAULT").Msg("Not found")
		return
	}
	links := make(map[string]string)
	links["iOS"] = os.Getenv("LINK_APPSTORE")
	links["Android"] = os.Getenv("LINK_GOOGLEPLAY")
	platformLinks = links
	defaultLink = link
}

func startServer() {
	log.Info().Str("addr", port).Msg("Starting server")
	http.HandleFunc("/", linkHandler)
	http.HandleFunc("/meta", metaHandler)

	http.ListenAndServe(port, nil)
}

func linkHandler(w http.ResponseWriter, r *http.Request) {
	platform := getOS(r)
	link := getLink(platform)
	ip := getIP(r)

	metaData := r.URL.Query().Get("meta")
	if metaData != "" {
		err := meta.Set(ip, metaData)
		if err != nil {
			log.Error().Err(err).Str("ip", ip).Msg("Failed to set metadata")
		}
	}
	if tracker != nil {
		err := tracker.LogClick(metaData, platform, context.Background())
		if err != nil {
			log.Error().Err(err).Str("ip", ip).Msg("Failed to log click")
		}
	}

	http.Redirect(w, r, link, http.StatusMovedPermanently)
}

func metaHandler(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	metaData, _ := meta.Get(ip)
	if metaData != "" && tracker != nil {
		platform := getOS(r)
		err := tracker.LogInstall(metaData, platform, context.Background())
		if err != nil {
			log.Error().Err(err).Str("ip", ip).Msg("Failed to log install")
		}
	}
	w.Write([]byte(metaData))
}

func getOS(r *http.Request) string {
	s := r.Header.Get("User-Agent")
	if s == "" {
		return "Unknown"
	}
	return useragent.Parse(s).OS
}

func getLink(platform string) string {
	link := platformLinks[platform]
	if link != "" {
		return link
	} else {
		return defaultLink
	}
}

func getIP(r *http.Request) string {
	cloudflareIP := r.Header.Get("CF-Connecting-IP")
	if cloudflareIP != "" {
		return cloudflareIP
	}
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	return ""
}
