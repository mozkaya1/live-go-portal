package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/mozkaya1/live-go-portal/internal/api"
)

type server struct {
	subscriberMessageBuffer int
	mux                     http.ServeMux
	subscribersMu           sync.Mutex
	subscribers             map[*subscriber]struct{}
}

type subscriber struct {
	msgs chan []byte
}

func NewServer() *server {
	s := &server{
		subscriberMessageBuffer: 10,
		subscribers:             make(map[*subscriber]struct{}),
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)
	return s
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := s.subscribe(r.Context(), w, r)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *server) addSubscriber(subscriber *subscriber) {
	s.subscribersMu.Lock()
	s.subscribers[subscriber] = struct{}{}
	s.subscribersMu.Unlock()
	fmt.Println("Added subscriber", subscriber)
}

// Remove subscriber function to delete unused connection, release buffer ...
func (s *server) removeSubscriber(subscriber *subscriber) {
	s.subscribersMu.Lock()
	delete(s.subscribers, subscriber)
	s.subscribersMu.Unlock()
	fmt.Println("Removed subscriber", subscriber)
}

func (s *server) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var c *websocket.Conn
	subscriber := &subscriber{
		msgs: make(chan []byte, s.subscriberMessageBuffer),
	}
	s.addSubscriber(subscriber)
	defer s.removeSubscriber(subscriber)

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (cs *server) publishMsg(msg []byte) {
	cs.subscribersMu.Lock()
	defer cs.subscribersMu.Unlock()
	for s := range cs.subscribers {
		select {
		case s.msgs <- msg:
			// Message sent successfully
		default:
			// Channel is full, remove the subscriber
			delete(cs.subscribers, s)
			close(s.msgs)
			fmt.Println("Removed subscriber due to full channel", s)
		}
	}
}

// Added for optimize with no-fetching data if there are no subscriber ...
func (s *server) hasSubscribers() bool {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()
	return len(s.subscribers) > 0
}

// Helper function to get change class
func getChangeClass(priceChange string) string {
	priceChange = strings.TrimSpace(priceChange)
	if strings.HasPrefix(priceChange, "-") || strings.HasPrefix(priceChange, "%-") {
		return "change-negative"
	}

	// // Remove + sign if present and check if it's a positive number
	// clean := strings.TrimPrefix(priceChange, "+")
	// if clean != priceChange || (len(clean) > 0 && clean != "0" && clean != "0.00" && clean != "0.0") {
	// 	// It had a + sign or is a non-zero number
	// 	return "change-positive"
	// }

	return "change-positive"
}

func main() {
	refreshTime := os.Getenv("refreshTime")
	refreshTimeInt, err := strconv.Atoi(refreshTime)
	if err != nil {
		log.Println(err)
		refreshTimeInt = 20 // setting 20 secs refresh time - default
	}
	fmt.Println("refresh", refreshTimeInt)
	fmt.Println("Starting monitor server on port 8080")
	fmt.Println("open browser at http://localhost:8080")
	s := NewServer()

	go func(srv *server) {

		// Created ticker to check every 1 sec for subscriber
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {

			<-ticker.C
			if !srv.hasSubscribers() {
				continue
			}

			coin, err := api.GetWeatherApi()
			if err != nil {
				log.Println(err, "GetWeatherApi Error")
			}
			primeCurrency, err := api.GetRealCurrencyApi()

			if err != nil {
				log.Println(err, "GetRealCurrencyApi Error")
			}
			// Prime Asset CSS change
			dolarChangeClass := getChangeClass(primeCurrency.PrimeAssets["Dolar"].Change)
			euroChangeClass := getChangeClass(primeCurrency.PrimeAssets["Euro"].Change)
			goldChangeClass := getChangeClass(primeCurrency.PrimeAssets["ONS Altın"].Change)
			euroDolarChangeClass := getChangeClass(primeCurrency.PrimeAssets["Euro Dolar"].Change)

			// Crypto-Weather CSS change
			btcChangeClass := getChangeClass(coin.Crypto.Asset["BTCUSDT"].PriceChangePercent)
			ethChangeClass := getChangeClass(coin.Crypto.Asset["ETHUSDT"].PriceChangePercent)
			atomChangeClass := getChangeClass(coin.Crypto.Asset["ATOMUSDT"].PriceChangePercent)
			linkChangeClass := getChangeClass(coin.Crypto.Asset["LINKUSDT"].PriceChangePercent)
			filChangeClass := getChangeClass(coin.Crypto.Asset["FILUSDT"].PriceChangePercent)
			xlmChangeClass := getChangeClass(coin.Crypto.Asset["XLMUSDT"].PriceChangePercent)
			thetaChangeClass := getChangeClass(coin.Crypto.Asset["THETAUSDT"].PriceChangePercent)
			bnbChangeClass := getChangeClass(coin.Crypto.Asset["BNBUSDT"].PriceChangePercent)
			xrpChangeClass := getChangeClass(coin.Crypto.Asset["XRPUSDT"].PriceChangePercent)
			iotaChangeClass := getChangeClass(coin.Crypto.Asset["IOTAUSDT"].PriceChangePercent)

			timeStamp := time.Now().Format("2006-01-02 15:04:05")
			// up := exec.Command("who", "-b")
			// uptime, err := up.Output()

			timeLayout := "03:04 PM"
			parsedSunsetTime, err := time.Parse(timeLayout, coin.WeatherBucket.Sunset)
			if err != nil {
				log.Println("Time Parse Error")
			}

			now := time.Now()
			sunsetModifiedTime := time.Date(now.Year(), now.Month(), now.Day(), parsedSunsetTime.Hour(), parsedSunsetTime.Minute(), 0, 0, now.Location())
			iftarDuration := sunsetModifiedTime.Sub(now)
			if iftarDuration < 0 {
				iftarDuration += 24 * time.Hour
			}

			iftarDurationLeft := iftarDuration.Round(100 * time.Millisecond).String()

			msg := []byte(`
      <div hx-swap-oob="innerHTML:#update-timestamp">
      <i style="color: green" class="fa fa-circle"></i> ` + timeStamp + `
      </div>
      <div hx-swap-oob="innerHTML:#temp"> ` + coin.WeatherBucket.Temp + " / " + coin.WeatherBucket.FeelsLikeC + " °C" + `</div>
      <div hx-swap-oob="innerHTML:#weather"> ` + coin.WeatherBucket.WeatherDesc + `</div>
      <div hx-swap-oob="innerHTML:#location"> ` + coin.WeatherBucket.Location + `</div>
      <div hx-swap-oob="innerHTML:#sunrise"> ` + coin.WeatherBucket.Sunrise + `</div>
      <div hx-swap-oob="innerHTML:#sunset"> ` + coin.WeatherBucket.Sunset + `</div>
      <div hx-swap-oob="innerHTML:#iftar"> ` + iftarDurationLeft + `</div>
      <div hx-swap-oob="innerHTML:#dolar-price"> ` + primeCurrency.PrimeAssets["Dolar"].Price + `</div>
      <div hx-swap-oob="outerHTML:#dolar-change" class="crypto-change ` + dolarChangeClass + `">` + primeCurrency.PrimeAssets["Dolar"].Change + `</div>

      <div hx-swap-oob="innerHTML:#euro-price"> ` + primeCurrency.PrimeAssets["Euro"].Price + `</div>
      <div hx-swap-oob="outerHTML:#euro-change" class="crypto-change ` + euroChangeClass + `">` + primeCurrency.PrimeAssets["Euro"].Change + `</div>

      <div hx-swap-oob="innerHTML:#gold-price"> ` + primeCurrency.PrimeAssets["ONS Altın"].Price + `</div>
      <div hx-swap-oob="outerHTML:#gold-change" class="crypto-change ` + goldChangeClass + `">` + primeCurrency.PrimeAssets["ONS Altın"].Change + `</div>
      
      <div hx-swap-oob="innerHTML:#euro-dolar-price"> ` + primeCurrency.PrimeAssets["Euro Dolar"].Price + `</div>
      <div hx-swap-oob="outerHTML:#euro-dolar-change" class="crypto-change ` + euroDolarChangeClass + `">` + primeCurrency.PrimeAssets["Euro Dolar"].Change + `</div>


      <div hx-swap-oob="innerHTML:#btc-price"> ` + strings.TrimRight(coin.Crypto.Asset["BTCUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#btc-change" class="crypto-change ` + btcChangeClass + `">` + coin.Crypto.Asset["BTCUSDT"].PriceChangePercent + `%</div>
      <div hx-swap-oob="innerHTML:#eth-price"> ` + strings.TrimRight(coin.Crypto.Asset["ETHUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#eth-change" class="crypto-change ` + ethChangeClass + `">` + coin.Crypto.Asset["ETHUSDT"].PriceChangePercent + `%</div>


      <div hx-swap-oob="innerHTML:#atom-price"> ` + strings.TrimRight(coin.Crypto.Asset["ATOMUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#atom-change" class="crypto-change ` + atomChangeClass + `">` + coin.Crypto.Asset["ATOMUSDT"].PriceChangePercent + `%</div>


      <div hx-swap-oob="innerHTML:#link-price"> ` + strings.TrimRight(coin.Crypto.Asset["LINKUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#link-change" class="crypto-change ` + linkChangeClass + `">` + coin.Crypto.Asset["LINKUSDT"].PriceChangePercent + `%</div>


      <div hx-swap-oob="innerHTML:#fil-price"> ` + strings.TrimRight(coin.Crypto.Asset["FILUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#fil-change" class="crypto-change ` + filChangeClass + `">` + coin.Crypto.Asset["FILUSDT"].PriceChangePercent + `%</div>
			
      <div hx-swap-oob="innerHTML:#xlm-price"> ` + strings.TrimRight(coin.Crypto.Asset["XLMUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#xlm-change" class="crypto-change ` + xlmChangeClass + `">` + coin.Crypto.Asset["XLMUSDT"].PriceChangePercent + `%</div>
      
			
      <div hx-swap-oob="innerHTML:#theta-price"> ` + strings.TrimRight(coin.Crypto.Asset["THETAUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#theta-change" class="crypto-change ` + thetaChangeClass + `">` + coin.Crypto.Asset["THETAUSDT"].PriceChangePercent + `%</div>

			
      <div hx-swap-oob="innerHTML:#bnb-price"> ` + strings.TrimRight(coin.Crypto.Asset["BNBUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#bnb-change" class="crypto-change ` + bnbChangeClass + `">` + coin.Crypto.Asset["BNBUSDT"].PriceChangePercent + `%</div>

			
      <div hx-swap-oob="innerHTML:#xrp-price"> ` + strings.TrimRight(coin.Crypto.Asset["XRPUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#xrp-change" class="crypto-change ` + xrpChangeClass + `">` + coin.Crypto.Asset["XRPUSDT"].PriceChangePercent + `%</div>

			
      <div hx-swap-oob="innerHTML:#iota-price"> ` + strings.TrimRight(coin.Crypto.Asset["IOTAUSDT"].LastPrice, "0") + `</div>
      <div hx-swap-oob="outerHTML:#iota-change" class="crypto-change ` + iotaChangeClass + `">` + coin.Crypto.Asset["IOTAUSDT"].PriceChangePercent + `%</div>

      <div hx-swap-oob="innerHTML:#market-update-time">` + timeStamp + `</div>


      `)
			srv.publishMsg(msg)

			// Refreshing DATA inverval
			// fmt.Println(string(msg))
			time.Sleep(time.Duration(refreshTimeInt) * time.Second)
		}
	}(s)

	err = http.ListenAndServe(":8080", &s.mux)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
