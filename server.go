package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Products []Product

type productHandler struct {
	sync.Mutex
	products Products
}

func (p *productHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p.get(w, r)
	case "POST":
		p.post(w, r)
	case "PUT", "PATCH":
		p.put(w, r)
	case "DELETE":
		p.delete(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "invalid method")
	}
}

func (p *productHandler) get(w http.ResponseWriter, r *http.Request) {
	defer p.Unlock()
	p.Lock()
	id, err := IDFromURL(r)
	if err != nil {
		respondWithJSON(w, http.StatusOK, p.products)
		return
	}
	if id >= len(p.products) || id < 0 {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	respondWithJSON(w, http.StatusOK, p.products[id])
}

func (p *productHandler) post(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		respondWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json' required")
		return
	}
	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer p.Unlock()
	p.Lock()
	p.products = append(p.products, product)
	respondWithJSON(w, http.StatusCreated, product)
}

func (p *productHandler) put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := IDFromURL(r)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		respondWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json' required")
		return
	}
	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer p.Unlock()
	p.Lock()
	if id >= len(p.products) || id < 0 {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	if product.Name != "" {
		p.products[id].Name = product.Name
	}
	if product.Price != 0 {
		p.products[id].Price = product.Price
	}
	respondWithJSON(w, http.StatusOK, p.products[id])
}

func (p *productHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := IDFromURL(r)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	defer p.Unlock()
	p.Lock()
	if id >= len(p.products) || id < 0 {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	if id < len(p.products)-1 {
		p.products[len(p.products)-1], p.products[id] = p.products[id], p.products[len(p.products)-1]
	}
	p.products = p.products[:len(p.products)-1]
	respondWithJSON(w, http.StatusNoContent, "")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response, _ := json.Marshal(data)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func IDFromURL(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		return 0, errors.New("not found")
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, errors.New("not found")
	}
	return id, nil
}

func newProductHandler() *productHandler {
	return &productHandler{
		products: Products{
			{
				Name:  "Shoes",
				Price: 25.00,
			},
			{
				Name:  "Webcam",
				Price: 50.00,
			},
			{
				Name:  "Mic",
				Price: 20.00,
			},
		},
	}
}

func main() {
	port := ":8080"
	p := newProductHandler()
	http.Handle("/products", p)
	http.Handle("/products/", p)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World\n")
	})
	log.Fatal(http.ListenAndServe(port, nil))
}
