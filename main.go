package main

import (
	"gopkg.in/olivere/elastic.v3"
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"github.com/pborman/uuid"
	"cloud.google.com/go/bigtable"
	"context"
)

const (
	INDEX    = "around"
	TYPE     = "post"
	DISTANCE = "200km"

	PROJECT_ID = "around-210816"
	BT_INSTANCE = "around-post"
	// Needs to update this URL if you deploy it to cloud.
	ES_URL = "http://35.196.231.235:9200"
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

	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Use the IndexExists service to check if a specified index exists.
	/*If not, create a new mapping. For other fields (user, message, etc.)
	no need to have mapping as they are default. For geo location (lat, lon),
	we need to tell ES that they are geo points instead of two float points
	such that ES will use Geo-indexing for them (K-D tree)
	*/
	exists, err := client.IndexExists(INDEX).Do()
	if err != nil {
		panic(err)
	}
	if !exists {
		// Create a new index.
		mapping := `{
                    "mappings":{
                           "post":{
                                  "properties":{
                                         "location":{
                                                "type":"geo_point"
                                         }
                                  }
                           }
                    }
             }
             `
		_, err := client.CreateIndex(INDEX).Body(mapping).Do()
		if err != nil {
			// Handle error
			panic(err)
		}
	}

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

	id := uuid.New()
	// Save to ES.
	saveToES(&p, id)

	ctx := context.Background()

	// you must update project name here
	bt_client, err := bigtable.NewClient(ctx, PROJECT_ID, BT_INSTANCE)
	if err != nil {
		panic(err)
		return
	}


	tbl := bt_client.Open("post")
	mut := bigtable.NewMutation()
	t := bigtable.Now()

	mut.Set("post", "user", t, []byte(p.User))
	mut.Set("post", "message", t, []byte(p.Message))
	mut.Set("location", "lat", t, []byte(strconv.FormatFloat(p.Location.Lat, 'f',-1, 64)))
	mut.Set("location", "lon", t, []byte(strconv.FormatFloat(p.Location.Lon, 'f',-1, 64)))

	err = tbl.Apply(ctx, id, mut)
	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("Post is saved to BigTable: %s\n", p.Message)
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

	/*1.3.7
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
	*/

	/*2.0*/
	fmt.Println("Received a Search Request")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

	// range is optional
	ran := DISTANCE
	if val := r.URL.Query().Get("range"); val != "" {
		ran = val + "km"
	}

	fmt.Printf("Search received: %f %f %s\n", lat, lon, ran)

	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Prepare a geo based query to find posts within a geo box.
	q := elastic.NewGeoDistanceQuery("location")
	q = q.Distance(ran).Lat(lat).Lon(lon)

	// Get the results based on Index (similar to dataset) and query (q that we just prepared). Pretty means to format the output.
	searchResult, err := client.Search().
		Index(INDEX).
		Query(q).
		Pretty(true).
		Do()
	if err != nil {
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d post\n", searchResult.TotalHits())

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization.
	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) { // instance of
		p := item.(Post) // p = (Post) item
		fmt.Printf("Post by %s: %s at lat %v and lon %v\n", p.User, p.Message, p.Location.Lat, p.Location.Lon)
		// TODO(student homework): Perform filtering based on keywords such as web spam etc.

		// Add the p to an array, equals ps.add(p) in java
		ps = append(ps, p)
	}

	// Convert the go object to a string
	js, err := json.Marshal(ps)
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Allow cross domain visit for javascript.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

// Save a post to ElasticSearch
func saveToES(p *Post, id string) {
	// Create a client
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Save it to index
	_, err = es_client.Index().
		Index(INDEX).
		Type(TYPE).
		Id(id).
		BodyJson(p).
		Refresh(true).
		Do()
	if err != nil {
		panic(err)
		return
	}
}

