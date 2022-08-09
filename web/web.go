package web

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"text/template"

	"github.com/d-rep/netmon/storage"
)

var (
	//go:embed index.html
	indexTemplate string
)

type TemplateData struct {
	Title string
	Calls []*storage.Call
}

func GetIndex(db *storage.Storage) http.HandlerFunc {
	t, err := template.New("tmpl").Parse(indexTemplate)
	if err != nil {
		log.Fatal(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		calls, err := db.GetRecentCalls(10)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		templateData := TemplateData{
			Title: "Network Monitor",
			Calls: calls,
		}

		err = t.Execute(w, templateData)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func GetStatus(db *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		calls, err := db.GetRecentCalls(10)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(calls)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	}
}

func Serve(port string, db *storage.Storage) error {
	http.Handle("/", GetIndex(db))
	http.Handle("/status", GetStatus(db))
	return http.ListenAndServe("localhost:"+port, nil)
}
