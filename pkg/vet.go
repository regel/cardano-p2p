package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dchest/blake2b"
	"github.com/gorilla/websocket"
	"github.com/regel/cardano-p2p/log"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

var mutex sync.Mutex

type Query struct {
	MethodName  string    `json:"methodname"`
	ServiceName string    `json:"servicename"`
	QueryType   string    `json:"type"`
	Version     string    `json:"version"`
	QueryArgs   QueryArgs `json:"args"`
}

type QueryArgs struct {
	Query interface{} `json:"query"`
}

type PoolIdsResponse struct {
	Result []string `json:"result"`
}

type BlockHeightResponse struct {
	Result *int64 `json:"result"`
}

type PoolRelay struct {
	Port     int     `json:"port"`
	Ip4      *string `json:"ipv4"`
	Ip6      *string `json:"ipv6"`
	HostName *string `json:"hostname"`
}

type PoolMetadata struct {
	Url  string `json:"url"`
	Hash string `json:"hash"`
}

type PoolParameters struct {
	Id            string       `json:"id"`
	Vrf           string       `json:"vrf"`
	Pledge        uint64       `json:"pledge"`
	Cost          uint64       `json:"cost"`
	Margin        string       `json:"margin"`
	RewardAccount string       `json:"rewardAccount"`
	Owners        []string     `json:"owners"`
	Relays        []PoolRelay  `json:"relays"`
	Metadata      PoolMetadata `json:"metadata"`
}

type PoolParametersResponse struct {
	Result map[string]PoolParameters `json:"result"`
}

func buildPoolIdsQuery() Query {
	args := QueryArgs{
		Query: "poolIds",
	}
	query := Query{
		MethodName:  WebsocketMethodName,
		ServiceName: WebsocketServiceName,
		QueryType:   WebsocketQueryType,
		Version:     WebsocketVersion,
		QueryArgs:   args,
	}
	return query
}

func buildblockHeightQuery() Query {
	args := QueryArgs{
		Query: "blockHeight",
	}
	query := Query{
		MethodName:  WebsocketMethodName,
		ServiceName: WebsocketServiceName,
		QueryType:   WebsocketQueryType,
		Version:     WebsocketVersion,
		QueryArgs:   args,
	}
	return query
}

func buildPoolParametersQuery(poolId string) Query {
	var poolIds = map[string][]string{
		"poolParameters": []string{poolId},
	}
	args := QueryArgs{
		Query: poolIds,
	}
	query := Query{
		MethodName:  WebsocketMethodName,
		ServiceName: WebsocketServiceName,
		QueryType:   WebsocketQueryType,
		Version:     WebsocketVersion,
		QueryArgs:   args,
	}
	return query
}

func GetBlockHeight(url string) (*int64, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %q: %v\n", url, err)
	}
	defer ws.Close()

	msg := buildblockHeightQuery()
	data, _ := json.Marshal(msg)
	ws.WriteMessage(websocket.TextMessage, data)
	_, message, err := ws.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
			return nil, fmt.Errorf("unexpected read error %v\n", err)
		}
	}
	var blockHeight BlockHeightResponse
	err = json.Unmarshal(message, &blockHeight)
	if err != nil {
		return nil, err
	}
	return blockHeight.Result, nil
}

func getPoolParameters(url string, poolId string) (*PoolParameters, error) {
	msg := buildPoolParametersQuery(poolId)
	data, _ := json.Marshal(msg)
	// ensures that no more than one goroutine calls the write methods on the same ws
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %q: %v\n", url, err)
	}
	defer ws.Close()
	ws.WriteMessage(websocket.TextMessage, data)
	_, message, err := ws.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
			return nil, fmt.Errorf("Unexpected read error %v\n", err)
		}
	}
	var response PoolParametersResponse
	err = json.Unmarshal(message, &response)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error: %v\n", err)
	}
	poolParameters := response.Result[poolId]
	return &poolParameters, nil
}

func vetPool(url string, client *http.Client, poolId string) (*PoolParameters, error) {
	// ensures that no more than one goroutine calls the write methods on the same ws
	mutex.Lock()
	poolParameters, err := getPoolParameters(url, poolId)
	mutex.Unlock()
	if err != nil {
		return nil, err
	}
	if len(poolParameters.Relays) == 0 {
		return nil, fmt.Errorf("No relays")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestMaxWaitTime)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, poolParameters.Metadata.Url, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create request: %v", err)
	}
	req.Close = true
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if e, ok := err.(net.Error); ok && e.Timeout() {
		return nil, fmt.Errorf("Get '%s' timeout", poolParameters.Metadata.Url)
	} else if err != nil {
		return nil, fmt.Errorf("Cannot do request: %v", err)
	}
	buf, err := ioutil.ReadAll(io.LimitReader(resp.Body, MaxMetadataLen))
	if err != nil {
		return nil, fmt.Errorf("Cannot get all response body at url '%s': %v", poolParameters.Metadata.Url, err)
	}
	sum := blake2b.Sum256(buf)
	if fmt.Sprintf("%x", sum) != poolParameters.Metadata.Hash {
		return nil, fmt.Errorf("invalid hash expected '%s' was '%s'", poolParameters.Metadata.Hash, fmt.Sprintf("%x", sum))
	}
	log.Infof("Verified hash for pool '%s' at url '%s'", poolId, poolParameters.Metadata.Url)
	return poolParameters, nil
}

func getPoolIds(url string) ([]string, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %q: %v\n", url, err)
	}
	defer ws.Close()

	msg := buildPoolIdsQuery()
	data, _ := json.Marshal(msg)
	ws.WriteMessage(websocket.TextMessage, data)
	_, message, err := ws.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
			return nil, fmt.Errorf("unexpected read error %v\n", err)
		}
	}
	var poolIds PoolIdsResponse
	err = json.Unmarshal(message, &poolIds)
	if err != nil {
		return nil, err
	}
	return poolIds.Result, nil
}

func VetPools(url string) ([]*PoolParameters, error) {
	var wg sync.WaitGroup
	var ch = make(chan string, MaxWorkers)

	poolIds, err := getPoolIds(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool ids: %v\n", err)
	}

	poolChan := make(chan *PoolParameters)
	wg.Add(MaxWorkers)
	for i := 0; i < MaxWorkers; i++ {
		go func() {
			var netTransport = &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			}
			var cli = &http.Client{
				Transport: netTransport,
			}
			for {
				pool, ok := <-ch
				if !ok { // the channel has been closed.
					wg.Done()
					return
				}
				parameters, err := vetPool(url, cli, pool)
				if err != nil {
					log.Errorf("Error fetching pool '%s' data: %v", pool, err)
					continue
				}
				poolChan <- parameters
			}
		}()
	}
	pools := make([]*PoolParameters, 0)
	go func() {
		for parameters := range poolChan {
			pools = append(pools, parameters)
		}
	}()
	for _, poolId := range poolIds {
		ch <- poolId
	}
	close(ch)
	wg.Wait()
	close(poolChan)
	return pools, nil
}
