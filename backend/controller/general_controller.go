package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"notifier/tools"
)

type ApiInfoResult struct {
	Version  string `json:"version_info"`
	TimeZone string `json:"time_zone"`
}

type GeneralController struct {
	log *log.Logger
}

func NewGeneralController(l *log.Logger) *GeneralController {
	return &GeneralController{
		log: l,
	}
}

func (s *GeneralController) Add() {
	http.HandleFunc("/notifier/api/general/info", s.HandleInfo)
}

// @Summary      Get info about API
// @Description  Returns information about the API version and other info
// @Tags	     General
// @Success      200  {object} ApiInfoResult
// @Failure      500  {object} string
// @Router       /notifier/api/general/info [get]
func (s *GeneralController) HandleInfo(w http.ResponseWriter, r *http.Request) {
	s.log.Printf("Returning API info")

	resp := ApiInfoResult{
		Version:  "0.9.0",
		TimeZone: tools.ClientTZ().String(),
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
