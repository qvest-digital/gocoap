package main

import (
	"bytes"
	"log"
	"time"

	coap "github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

func getPath(opts message.Options) string {
	path, err := opts.Path()
	if err != nil {
		log.Printf("cannot get path: %v", err)
		return ""
	}
	return path
}

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("ClientAddress %v, %v\n", w.Client().RemoteAddr(), r.String())
		next.ServeCOAP(w, r)
	})
}

func handleHello(w mux.ResponseWriter, r *mux.Message) {
	err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("hello world")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func handleDelete(w mux.ResponseWriter, r *mux.Message) {
	err := w.SetResponse(codes.Deleted, message.TextPlain, bytes.NewReader([]byte("deleted")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func handleTime(w mux.ResponseWriter, r *mux.Message) {
	log.Printf("Got message path=%v: %+v from %v", getPath(r.Options), r, w.Client().RemoteAddr())

	subded := time.Now()
	err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte(subded.Format(time.RFC3339))))
	if err != nil {
		log.Printf("Error on transmitter: %v", err)
	}
}

func handleWrite(w mux.ResponseWriter, r *mux.Message) {
	customResp := message.Message{
		Code:    codes.Content,
		Token:   r.Token,
		Context: r.Context,
		Options: make(message.Options, 0, 16),
		Body:    bytes.NewReader([]byte("oldval")),
	}
	if (r.Code == codes.PUT) {
		customResp = message.Message{
			Code:    codes.Changed,
			Token:   r.Token,
			Context: r.Context,
			Options: make(message.Options, 0, 16),
			Body:    bytes.NewReader([]byte("changed")),
		}
	}
	optsBuf := make([]byte, 32)
	opts, used, err := customResp.Options.SetContentFormat(optsBuf, message.TextPlain)
	if err == message.ErrTooSmall {
		optsBuf = append(optsBuf, make([]byte, used)...)
		opts, used, err = customResp.Options.SetContentFormat(optsBuf, message.TextPlain)
	}
	if err != nil {
		log.Printf("cannot set options to response: %v", err)
		return
	}
	optsBuf = optsBuf[:used]
	customResp.Options = opts

	err = w.Client().WriteMessage(&customResp)
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func main() {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/Hello", mux.HandlerFunc(handleHello))
	r.Handle("/subpath/Another", mux.HandlerFunc(handleHello))
	r.Handle("/removeme!", mux.HandlerFunc(handleDelete))
	r.Handle("/writeme!", mux.HandlerFunc(handleWrite))
	r.Handle("/time", mux.HandlerFunc(handleTime))
	log.Printf("start Server")
	log.Fatal(coap.ListenAndServe("udp", ":5683", r))
}