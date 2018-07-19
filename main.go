package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strconv"
)

const (
	DISTANCE = "200km"
)

// 1.3.1 Encode json object
// 1.3.2 Add one method handlerPost() after main() to handle Post.
// 1.3.3 Replace main function to call handlerPost when started.
// 1.3.6 Add another handler for search (called it handlerSearch), the request has a url pattern like
// http://localhost:8080/search?lat=10.0&lon=20.0. Parse it and then print out the lat and lon.
// 1.3.7 return a fake JSON object
type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Post struct {
	// `json:"user"` is for the json parsing of this User field.
	// Otherwise, by default it's 'User'.
	User     string   `json:"user"`
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func main() {
	fmt.Println("Back-end Service Started Based on GoLang ")

	// http handler Func mapping, /post -> handlerPost
	http.HandleFunc("/post", handlerPost)

	//1.3.6
	http.HandleFunc("/search", handlerSearch)

	// if err, throw a fatal message, show log message
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlerPost(w http.ResponseWriter, r *http.Request) {
	// Parse from body of request to get a json object.
	fmt.Println("Reveived a Post Request")
	var p Post

	//decoder := json.NewDecoder(r.Body)

	// &p: a pointer
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		panic(err)
		return
	}

	fmt.Fprintf(w, "Post Received: %s\n", p.Message)
}

// 1.3.6
func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Reveived a Search Request")
	/* 1.3.6
	//  http://localhost:8080/search?lat=10.0&lon=20.0


	// to get request para from url:
	// lat := r.URL.Query().Get("lat")

	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	fmt.Fprintf(w, "Search Received, Latitude: %s, Lontitude: %s\n", lat, lon)
	*/

	/*1.3.7*/
	// _; we expect no invalid url from front-end
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

	// default range: 200km
	ran := DISTANCE
	if rangeValue := r.URL.Query().Get("range"); rangeValue != "" {
		ran = rangeValue + "km"
	}
	fmt.Println("Range is:", ran)

	// fake
	p := &Post{
		User:"1111",
		Message:"一生必去的100个地方",
		Location:Location{
			Lat:lat,
			Lon:lon,
		},
	}

	jsonObject, err := json.Marshal(p)
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write(jsonObject)
}
