package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"notifier/repo"
	"notifier/tools"
	"time"
)

type ApiInfoResult struct {
	Version    string         `json:"version_info"`
	TimeZone   string         `json:"time_zone"`
	ClientTime time.Time      `json:"client_time"`
	Count      int            `json:"reminder_count"`
	Metrics    map[string]int `json:"metrics"`
	TokenTtl   int64          `json:"token_ttl"`
}

type GeneralController struct {
	log             *log.Logger
	dbl             repo.DBSerializer
	metricCollector *tools.MetricsCollector
}

func NewGeneralController(s repo.DBSerializer, l *log.Logger, m *tools.MetricsCollector) *GeneralController {
	return &GeneralController{
		log:             l,
		dbl:             s,
		metricCollector: m,
	}
}

func (s *GeneralController) AddHandlersWithAuth(authWrapper tools.AuthWrapperFunc) {
	http.HandleFunc("/notifier/api/general/info", authWrapper(s.HandleInfo))
}

func countReminders(dbl repo.DBSerializer) int {
	_, readRepo := dbl.RLock()
	defer func() { dbl.RUnlock() }()

	count, err := repo.CountEntries(readRepo)
	if err != nil {
		count = 0
	}

	return count
}

// @Summary      Get info about API
// @Description  Returns information about the API version and other info
// @Tags	     General
// @Success      200  {object} ApiInfoResult
// @Failure      500  {object} string
// @Router       /notifier/api/general/info [get]
// @Security     ApiKeyAuth
func (s *GeneralController) HandleInfo(w http.ResponseWriter, r *http.Request) {
	s.log.Printf("Returning API info")

	resp := ApiInfoResult{
		Version:    tools.VersionString,
		TimeZone:   tools.ClientTZ().String(),
		ClientTime: time.Now().UTC().In(tools.ClientTZ()),
		Count:      countReminders(s.dbl),
		Metrics:    s.metricCollector.GetMetrics(),
		TokenTtl:   tools.TokenTtl,
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		s.log.Printf("error serializing response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(data))
}
