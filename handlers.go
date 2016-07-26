package main

import (
	"fmt"
	"log"
	"net/http"
)

type PubSub struct {
	Listeners map[string]map[string]*chan string
}

func (ps *PubSub) AddSubscriber(addr, topic string, comet chan string) {
	if ps.Listeners[topic] == nil {
		ps.Listeners[topic] = make(map[string]*chan string)
	}
	ps.Listeners[topic][addr] = &comet //append(ps.Listeners[topic][ss], &comet)
}

func (ps *PubSub) Publish(topic string, body string) error {
	lm := ps.Listeners[topic]
	for w := range lm {
		wrt := lm[w] // k //[v]
		*wrt <- body
	}
	return nil
}

func (ps *PubSub) RemoveSubscriber(addr, topic string) {
	delete(ps.Listeners[topic], addr)
}

type HTTPResources struct {
	serverMux *http.ServeMux
	ps        *PubSub
}

func NewHTTPResources() *HTTPResources {
	sm := http.NewServeMux()
	ps := PubSub{}
	ps.Listeners = make(map[string]map[string]*chan string)
	hr := HTTPResources{ps: &ps, serverMux: sm}
	hr.serverMux.HandleFunc("/api/v1/pub/", hr.PubHandler)
	hr.serverMux.HandleFunc("/api/v1/sub/", hr.SubHandler)
	return &hr
}

func (hr *HTTPResources) PubHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	topic := r.URL.Path[len("/api/v1/pub/"):]

	switch r.Method {
	case "POST":
		var oBody []string
		var err error

		if len(topic) < 1 {
			http.Error(w, "Invalid parameter range ", http.StatusBadRequest)
			return
		}
		if err = r.ParseForm(); err != nil {
			http.Error(w, "Error parsing parameters", http.StatusInternalServerError)
			return
		}
		if oBody = r.Form["body"]; oBody == nil {
			http.Error(w, "No body", http.StatusBadRequest)
			return
		}

		err = hr.ps.Publish(topic, oBody[0])
		if err != nil {
			http.Error(w, "Publish Error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Message published")
		return
		break
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

func (hr *HTTPResources) SubHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	topic := r.URL.Path[len("/api/v1/sub/"):]

	switch r.Method {
	case "GET":
		var comet = make(chan string)
		var timeout = make(chan bool)

		if len(topic) < 1 {
			http.Error(w, "Invalid parameter range ", http.StatusBadRequest)
			return
		}
		log.Println("Connected: ", r.RemoteAddr)

		hr.ps.AddSubscriber(r.RemoteAddr, topic, comet)
		f, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Oops", http.StatusInternalServerError)
			return
		}
		notify := w.(http.CloseNotifier).CloseNotify()

		for {
			select {
			case msg := <-comet:
				fmt.Fprintf(w, "%s\n", msg)
				f.Flush()
				break
			case stop := <-timeout:
				if stop {
					return
					break
				}
			case ok := <-notify:
				if ok {
					log.Println("Disconnecting ", r.RemoteAddr)
					hr.ps.RemoveSubscriber(r.RemoteAddr, topic)
					return
				}
				break
			}

		}
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}
