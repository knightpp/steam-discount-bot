package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	t "steam-discount/types"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SteamResponse struct {
	Success bool `json:"success"`
	Data    struct {
		PriceOverview SteamPriceOverview `json:"price_overview"`
	} `json:"data"`
}
type SteamPriceOverview struct {
	Currency         string `json:"currency"`
	Initial          int    `json:"initial"`
	Final            int    `json:"final"`
	DiscountPercent  int    `json:"discount_percent"`
	InitialFormatted string `json:"initial_formatted"`
	FinalFormatted   string `json:"final_formatted"`
}

// {"297130":{"success":true,"data":{"price_overview":{"currency":"UAH","initial":22900,"final":2200,"discount_percent":100,"initial_formatted":"229₴","final_formatted":"Free"}}}}
func requestPriceOverview(appIds []t.GameId) (map[t.GameId]SteamResponse, error) {
	log.WithField("app_ids", appIds).Trace()
	if len(appIds) > 100 {
		return nil, fmt.Errorf("allowed length of appIds is up to 100")
	}
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString("https://store.steampowered.com/api/appdetails?appids=")
	for _, appId := range appIds {
		fmt.Fprintf(&urlBuilder, "%s,", appId)
	}
	urlBuilder.WriteString("&filters=price_overview")
	log.WithField("url", urlBuilder.String()).Trace()
	resp, err := http.Get(urlBuilder.String())
	if err != nil {
		return nil, fmt.Errorf("requestPriceOverview http get returned: %w", err)
	}
	defer resp.Body.Close()
	sr := make(map[t.GameId]SteamResponse)
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return nil, fmt.Errorf("json failed to decode: %w", err)
	}
	log.WithField("map", sr).Trace("ended decoding json response")
	return sr, nil
}
