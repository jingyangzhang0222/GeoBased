package main

import (
	"gopkg.in/olivere/elastic.v3"
	"fmt"
	"reflect"
	"net/http"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	TYPE_USER = "user"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func checkUser(username, password string) bool {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		fmt.Printf("ES is not set up %v\n", err)
		return false
	}

	termQuery := elastic.NewTermQuery("username", username)
	queryResult, err := es_client.Search().
		Index(INDEX).
		Query(termQuery).
		Pretty(true).
		Do()
	if err != nil {
		fmt.Printf("ES query failed %v\n", err)
		return false
	}

	var tyu User
	for _, item := range queryResult.Each(reflect.TypeOf(tyu)) {
		u := item.(User)
		return u.Password == password && u.Username == username
	}

	return false
}

func addUser(username, password string) bool {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		fmt.Printf("ES is not set up %v\n", err)
		return false
	}

	user := &User {
		Username: username,
		Password: password,
	}

	termQuery := elastic.NewTermQuery("username", username)
	queryResult, err := es_client.Search().
		Index(INDEX).
		Query(termQuery).
		Pretty(true).
		Do()
	if err != nil {
		fmt.Printf("ES query failed %v\n", err)
		return false
	}

	if queryResult.TotalHits() > 0 {
		fmt.Printf("This user already exists! Username: %s\n", username)
		return false
	}

	// Save it to index
	_, err = es_client.Index().
		Index(INDEX).
		Type(TYPE_USER).
		Id(username).
		BodyJson(user).
		Refresh(true).
		Do()
	if err != nil {
		fmt.Printf("ES save failed! %v\n", err)
		return false
	}

	return true
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one signup request")

	decoder := json.NewDecoder(r.Body)
	var u User
	if err := decoder.Decode(&u); err != nil {
		panic(err)
		return
	}

	if u.Username != "" && u.Password != "" {
		if addUser(u.Username, u.Password) {
			fmt.Println("User added successfully.")
			w.Write([]byte("User added successfully."))
		} else {
			fmt.Println("Fail to add User")
			http.Error(w, "Fail to add User", http.StatusInternalServerError)
		}
	} else {
		fmt.Println("Empty password or username.")
		http.Error(w, "Empty password or username", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one login request")

	decoder := json.NewDecoder(r.Body)
	var u User
	if err := decoder.Decode(&u); err != nil {
		panic(err)
		return
	}

	if checkUser(u.Username, u.Password) {// Make sure user credential is correct.
		// Create a new token object to store.
		token := jwt.New(jwt.SigningMethodHS256)

		// Convert it into a map for lookup
		claims := token.Claims.(jwt.MapClaims)

		/* Set token claims */
		// Store username and expiration into it.
		claims["username"] = u.Username
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		/* Sign the token with our secret */
		tokenString, _ := token.SignedString(mySigningKey)

		/* Finally, write the token to the browser window */
		w.Write([]byte(tokenString))
	} else {
		// http 403 error
		fmt.Println("Invalid password or username.")
		http.Error(w, "Invalid password or username", http.StatusForbidden)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}