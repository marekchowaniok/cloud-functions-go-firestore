package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
)

// ArticlesType represents the articles collection in the database
type ArticlesType []map[string]interface{}

// ArticleFieldsType defines the structure of the fields in an article from the articles collection.
type ArticleFieldsType struct {
	ID          string  `firestore:"id"`
	Name        string  `firestore:"name"`
	Price       float64 `firestore:"price"`
	Type        string  `firestore:"type"`
	Year        string  `firestore:"year"`
	Image       string  `firestore:"image"`
	Description string  `firestore:"description"`
	Slug        string  `firestore:"slug"`
}

// DeleteType represents the body expected structure of a delete http call
type DeleteType struct {
	ID string `json:"id"`
}

// ArticleAPI is an HTTP Cloud Function with a request parameter.
func ArticleAPI(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: "***YOUR-PROJECT-ID***"}

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Printf("error initializing app: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Firestore init: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch method := r.Method; method {
	case http.MethodGet:
		getArticles(ctx, client, w)
	case http.MethodPost:
		setArticle(ctx, client, w, r)
	case http.MethodDelete:
		deleteArticle(ctx, client, w, r)
	case http.MethodPut:
		updateArticle(ctx, client, w, r)
	default:
		http.Error(w, "UNSUPPORTED METHOD", http.StatusNotFound)
	}

}

func getArticles(ctx context.Context, client *firestore.Client, w http.ResponseWriter) {
	var articles ArticlesType
	iter := client.Collection("articles").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Iteration over documents failed %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		articles = append(articles, doc.Data())
	}

	json.NewEncoder(w).Encode(articles)
}

func setArticle(ctx context.Context, client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var newArticle ArticleFieldsType

	err = json.Unmarshal(body, &newArticle)
	if err != nil {
		log.Printf("Unmarshalling json failed %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = client.Collection("articles").Doc(newArticle.ID).Create(ctx, &newArticle)
	if err != nil {
		log.Printf("Collection update failed %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteArticle(ctx context.Context, client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Reading request body failed %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var Body DeleteType

	err = json.Unmarshal(body, &Body)
	if err != nil {
		log.Printf("Unmarshalling json failed %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = client.Collection("articles").Doc(Body.ID).Delete(ctx)
	if err != nil {
		log.Printf("Document deletion failed %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updateArticle(ctx context.Context, client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Reading request body failed %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var Body ArticleFieldsType

	err = json.Unmarshal(body, &Body)
	if err != nil {
		log.Printf("Unmarshalling json failed %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = client.Collection("articles").Doc(Body.ID).Set(ctx, &Body)
	if err != nil {
		log.Printf("Document update failed %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
