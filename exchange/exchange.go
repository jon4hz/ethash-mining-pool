package exchange

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ei8ht187/ethash-mining-pool/storage"
	"github.com/ei8ht187/ethash-mining-pool/util"
)

type ExchangeProcessor struct {
	ExchangeConfig *ExchangeConfig
	backend        *storage.RedisClient
	rpc            *RestClient
}

type ExchangeConfig struct {
	Enabled         bool   `json:"enabled"`
	Name            string `json:"name"`
	Url             string `json:"url"`
	Timeout         string `json:"timeout"`
	RefreshInterval string `json:"refreshInterval"`
}

type RestClient struct {
	sync.RWMutex
	Url    string
	Name   string
	client *http.Client
}

type ExchangeReply []map[string]string

func NewRestClient(name, url, timeout string) *RestClient {
	restClient := &RestClient{Name: name, Url: url}
	timeoutIntv := util.MustParseDuration(timeout)
	restClient.client = &http.Client{
		Timeout: timeoutIntv,
	}
	return restClient
}

func (r *RestClient) GetData() (ExchangeReply, error) {
	Resp, err := r.doPost(r.Url, "ticker")
	if err != nil {
		return nil, err
	}

	// var reply map[string][]interface{}
	var data ExchangeReply
	err = json.Unmarshal(Resp, &data)
	return data, err
}

func StartExchangeProcessor(cfg *ExchangeConfig, backend *storage.RedisClient) *ExchangeProcessor {
	u := &ExchangeProcessor{ExchangeConfig: cfg, backend: backend}
	u.rpc = NewRestClient("ExchangeProcessor", cfg.Url, cfg.Timeout)
	return u
}

func (u *ExchangeProcessor) Start() {
	refreshIntv := util.MustParseDuration(u.ExchangeConfig.RefreshInterval)
	refreshTimer := time.NewTimer(refreshIntv)
	log.Printf("Set Exchange data refresh every %v", refreshIntv)

	u.fetchData()
	refreshTimer.Reset(refreshIntv)

	go func() {
		for range refreshTimer.C {
			u.fetchData()
			refreshTimer.Reset(refreshIntv)
		}
	}()
}

func (u *ExchangeProcessor) fetchData() {
	reply, err := u.rpc.GetData()
	if err != nil {
		log.Printf("Failed to fetch data from exchange %v", err)
		return
	}

	// log.Printf("Reply %v", reply)

	// Store the data into the Redis Store
	u.backend.StoreExchangeData(reply)

	if err != nil {
		log.Printf("Failed to Store the data to echange %v", err)
		return
	}
}

func (r *RestClient) doPost(url string, method string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	log.Println(req)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 { // OK
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)

		return bodyBytes, err2
	}

	return nil, err
}
