/*
Copyright © 2021 Sebastien Leger

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package pkg

import (
	"context"
	"github.com/regel/cardano-p2p/pkg/probe"
	"gopkg.in/validator.v1"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"encoding/json"
	"github.com/regel/cardano-p2p/log"
	"github.com/regel/cardano-p2p/server"
)

type PushPayload struct {
	ResultCode string `json:"resultcode"`
	Date       string `json:"datetime"`
	ClientIp   string `json:"clientIp"`
	IpType     int    `json:"iptype"`
	Msg        string `json:"msg"`
}

type PullPayload struct {
	ResultCode string     `json:"resultcode"`
	Date       string     `json:"datetime"`
	ClientIp   string     `json:"clientIp"`
	IpType     int        `json:"iptype"`
	Msg        string     `json:"msg"`
	Producers  []Producer `json:"Producers"`
}

type Producer struct {
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	Valency int    `json:"valency"`
}

type FetchRequest struct {
	Magic     uint64 `validate:"min=0"`
	Max       int    `validate:"min=1,max=20"`
	IpVersion int    `validate:"min=4"`
}

func Push(config *server.ClientConfig, ch chan<- Producer) {
	rand.Seed(time.Now().UnixNano())
	push(config, ch)
	for {
		<-time.After(config.PeriodSeconds)
		rand.Seed(time.Now().UnixNano())
		push(config, ch)
	}
}

func push(config *server.ClientConfig, ch chan<- Producer) {
	pools, err := VetPools(config.Endpoint)
	if err != nil {
		log.Errorf("Could not get pool data: %v", err)
		return
	}
	rand.Shuffle(len(pools), func(i, j int) { pools[i], pools[j] = pools[j], pools[i] })
	for _, pool := range pools {
		for _, relay := range pool.Relays {
			if relay.Ip4 != nil || relay.Ip6 != nil {
				if relay.Ip4 != nil {
					addr := net.JoinHostPort(*relay.Ip4, strconv.Itoa(relay.Port))
					tcpProbe, err := probe.DoTCPProbe(addr, config.ProbeTimeout)
					if tcpProbe != probe.Success {
						log.Errorf("tcp probe to '%s' failed: %v", addr, err)
						continue
					}
					log.Infof("tcp probe to '%s' success", addr)
					ch <- Producer{
						Addr:    *relay.Ip4,
						Port:    relay.Port,
						Valency: 1,
					}
				} else if relay.Ip6 != nil {
					addr := net.JoinHostPort(*relay.Ip6, strconv.Itoa(relay.Port))
					tcpProbe, err := probe.DoTCPProbe(addr, config.ProbeTimeout)
					if tcpProbe != probe.Success {
						log.Errorf("tcp probe to '%s' failed: %v", addr, err)
						continue
					}
					log.Infof("tcp probe to '%s' success", addr)
					ch <- Producer{
						Addr:    *relay.Ip6,
						Port:    relay.Port,
						Valency: 1,
					}
				}
			} else if relay.HostName != nil {
				addr := net.JoinHostPort(*relay.HostName, strconv.Itoa(relay.Port))
				tcpProbe, err := probe.DoTCPProbe(addr, config.ProbeTimeout)
				if tcpProbe != probe.Success {
					log.Errorf("tcp probe to '%s' failed: %v", addr, err)
					continue
				}
				ips, err := net.LookupIP(*relay.HostName)
				val := len(ips)
				if err != nil {
					log.Errorf("dns lookup to '%s' failed: %v", addr, err)
					continue
				}
				log.Infof("tcp probe to '%s' success", addr)
				ch <- Producer{
					Addr:    *relay.HostName,
					Port:    relay.Port,
					Valency: val,
				}
			}
		}
	}
}

func writeFetch(config *server.ServerConfig, w http.ResponseWriter, t *FetchRequest, clientIp string, ch chan Producer, defaultPeer string) {
	p := make([]Producer, 0)
	for i := t.Max; i > 0; i-- {
		select {
		case x, ok := <-ch:
			if ok {
				p = append(p, x)
			} else {
				w.WriteHeader(400)
				return
			}
		case <-time.After(config.ReadTimeout):
			break
		default:
			break
			// No more values to read
		}
	}
	if len(p) == 0 && defaultPeer != "" {
		addr, portStr, err := net.SplitHostPort(defaultPeer)
		if err == nil {
			ips, err := net.LookupIP(addr)
			val := len(ips)
			if err != nil {
				val = 1
			}
			port, _ := strconv.Atoi(portStr)
			p = append(p, Producer{
				Addr:    addr,
				Port:    port,
				Valency: val,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	pull := PullPayload{
		ResultCode: "200",
		Date:       time.Now().Format("2006-01-02 15:04:05"),
		ClientIp:   clientIp,
		IpType:     t.IpVersion,
		Msg:        "welcome to the topology",
		Producers:  p,
	}
	_ = json.NewEncoder(w).Encode(pull)
}

func Serve(config *server.ServerConfig, ch chan Producer) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		return
	})
	mux.HandleFunc("/htopology/v1/", func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Infof("userip: %q is not IP:port", r.RemoteAddr)
			w.WriteHeader(400)
			return
		}
		clientIp := ip
		forward := r.Header.Get("X-Forwarded-For")
		if forward != "" {
			clientIp = forward
		}
		p := PushPayload{
			ResultCode: "203",
			Date:       time.Now().Format("2006-01-02 15:04:05"),
			ClientIp:   clientIp,
			IpType:     4,
			Msg:        "welcome to the topology",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(203)
		_ = json.NewEncoder(w).Encode(p)
	})
	mux.HandleFunc("/htopology/v1/fetch/", func(w http.ResponseWriter, r *http.Request) {
		var i int64
		var err error
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Infof("userip: %q is not IP:port", r.RemoteAddr)
			w.WriteHeader(400)
			return
		}
		clientIp := ip
		forward := r.Header.Get("X-Forwarded-For")
		if forward != "" {
			clientIp = forward
		}

		s, ok := r.URL.Query()["magic"]
		if !ok {
			w.WriteHeader(400)
			return
		}
		if i, err = strconv.ParseInt(s[0], 10, 64); err != nil {
			log.Infof("failed to parse magic: %v", err)
			w.WriteHeader(400)
			return
		}
		magic := uint64(i)
		max := config.MaxPeers
		if s, ok := r.URL.Query()["max"]; ok {
			if i, err = strconv.ParseInt(s[0], 10, 64); err != nil {
				log.Infof("failed to parse max: %v", err)
				w.WriteHeader(400)
				return
			}
			max = int(i)
		}
		ipv := 4
		if s, ok := r.URL.Query()["ipv"]; ok {
			if i, err = strconv.ParseInt(s[0], 10, 64); err != nil {
				log.Infof("failed to parse ipv: %v", err)
				w.WriteHeader(400)
				return
			}
			ipv = int(i)
		}

		t := &FetchRequest{
			Magic:     magic,
			Max:       max,
			IpVersion: ipv,
		}
		if ok, errs := validator.Validate(t); !ok {
			log.Infof("validation failed: %v", errs)
			w.WriteHeader(400)
			return
		}
		if t.Magic != config.NetworkMagic {
			w.WriteHeader(400)
			return
		}
		writeFetch(config, w, t, clientIp, ch, config.DefaultPeer)
	})

	httpListener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		panic(err)
	}
	httpServer := &http.Server{
		Addr:    config.ListenAddress,
		Handler: mux,
	}

	log.Infof("listening: %s", config.ListenAddress)
	go func() {
		if err := httpServer.Serve(httpListener); err != nil {
			log.Infof("%s", err)
		}
	}()

	// watch for termination signals
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)    // terminal
		signal.Notify(sigint, syscall.SIGTERM) // kubernetes
		sig := <-sigint
		log.Infof("shutdown signal: %s", sig)
		signal.Stop(sigint)

		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := httpServer.Shutdown(timeout); err != nil {
			log.Errorf("http shutdown: %v", err)
		}
		cancel()

		log.Infof("shutdown complete")
		os.Exit(0)
	}()
}
