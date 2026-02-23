package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type WeatherBucket struct {
	Status           int    `json:"status"`
	UpdateTime       string `json:"updatetime"`
	Location         string `json:"location,omitempty"`
	Temp             string `json:"temp,omitempty"`
	WeatherDesc      string `json:"weatherDesc,omitempty"`
	Humidity         string `json:"humidity"`
	FeelsLikeC       string `json:"feelsLikeC,omitempty"`
	WindspeedKm      string `json:"windspeedKm"`
	AreaName         string `json:"areaName"`
	Latitude         string `json:"latitude"`
	Longitude        string `json:"longitude"`
	Country          string `json:"country"`
	Sunrise          string `json:"sunrise"`
	Sunset           string `json:"sunset"`
	MoonIllumination string `json:"moon_illumination"`
	MoonPhase        string `json:"moon_phase"`
	Moonrise         string `json:"moonrise"`
	Moonset          string `json:"moonset"`
}

type Currency struct {
	Status int                `json:"status"`
	Assets map[string]float64 `json:"assets"`
}

type CryptoAsset struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
}

type Crypto struct {
	Status int                    `json:"status"`
	Asset  map[string]CryptoAsset `json:"asset"`
}

type APIResponse struct {
	Time          string        `json:"time"`
	WeatherBucket WeatherBucket `json:"weatherbucket"`
	Currency      Currency      `json:"currency"`
	Crypto        Crypto        `json:"crypto"`
}

func GetWeatherApi() (APIResponse, error) {
	// change API url according to your requirement as location, asset, coin etc..
	// Details : https://github.com/mozkaya1/go-api#
	url := os.Getenv("API1_URL")
	resp, err := http.Get(url)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, err
	}
	var v APIResponse
	err = json.Unmarshal(r, &v)
	if err != nil {
		return APIResponse{}, err
	}
	return v, nil
}
