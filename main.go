package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/couchbase/gocb"
	"github.com/gorilla/mux"
	hashids "github.com/speps/go-hashids"
)

type MyUrl struct {
	ID       string `json:"id,omitempty"`
	LongUrl  string `json:"longUrl,omitempty"`
	ShortUrl string `json:"shortUrl,omitempty"`
}

var bucket *gocb.Bucket
var bucketName string

func ExpandEndpoint(w http.ResponseWriter, req *http.Request) {

}

func CreateEndpoint(w http.ResponseWriter, req *http.Request) {
	var url MyUrl
	err := json.NewDecoder(req.Body).Decode(&url)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	n1qlParams := []interface{}{}
	n1qlParams = append(n1qlParams, url.LongUrl)
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "`WHERE longUrl = $1`")
	rows, err := bucket.ExecuteN1qlQuery(query, n1qlParams)

	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}

	var row MyUrl
	rows.One(&row)
	if row == (MyUrl{}) {
		hd := hashids.NewData()
		h, _ := hashids.NewWithData(hd)
		now := time.Now()
		url.ID, _ = h.Encode([]int{int(now.Unix())})
		url.ShortUrl = "http://localhost:12345/" + url.ID
		bucket.Insert(url.ID, url, 0)
	} else {
		url = row
	}
	json.NewEncoder(w).Encode(url)
}

func RootEndpoint(w http.ResponseWriter, req *http.Request) {

}

func main() {
	router := mux.NewRouter()
	cluster, err := gocb.Connect("http://127.0.0.1")
	if err != nil {
		fmt.Println(err)
	}

	auth := &gocb.PasswordAuthenticator{
		Username: "Post",
		Password: "postpost",
	}

	cluster.Authenticate(auth)

	bucket, err = cluster.OpenBucket("example", "")
	if err != nil {
		fmt.Println(err)
	}
	router.HandleFunc("/create", CreateEndpoint).Methods("POST")
	router.HandleFunc("/expand/", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/id{}", RootEndpoint).Methods("GET")
	log.Println("Server is running on: localhost:12345..")
	log.Fatal(http.ListenAndServe(":12345", router))

}
