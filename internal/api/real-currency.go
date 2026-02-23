package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type MainTree struct {
	UpdateTime  string           `json:"time"`
	PrimeAssets map[string]Asset `json:"primeassets"`
	Others      map[string]Asset `json:"others"`
}

type Asset struct {
	Name   string `json:"name"`
	Price  string `json:"price"`
	Change string `json:"change"`
}

func GetRealCurrencyApi() (MainTree, error) {
	url := os.Getenv("API2_URL")
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return MainTree{}, err
	}
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return MainTree{}, err
	}

	var v MainTree
	err = json.Unmarshal(output, &v)
	if err != nil {
		log.Println(err)
		return MainTree{}, err
	}
	return v, nil

}
