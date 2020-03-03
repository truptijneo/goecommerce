package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// PonzuProduct struct ...
type PonzuProduct struct {
	UUID string 		`json : "uuid"`
	ID 	int				`json : "id"`
	Slug string			`json : "slug"`
	Timestamp int64 	`json : "timestamp"`
	Updated int64 		`json : "updated"`
	Name string 		`json : "name"`
	Price float32 		`json : "price"`
	Description string 	`json : "description"`
	Image string 		`json : "image"`
}

// PonzuProductResponse struct ...
type PonzuProductResponse struct {
	Data []PonzuProduct `json:"data"`
}

// HugoProduct struct ...
type HugoProduct struct {
	ID string 			`json : "id"`
	Title string 		`json : "title"`
	Date time.Time 		`json : "date"`
	LastModification time.Time 	`json : "lastmod"`
	Description	string 	
	Price float32		`json : "price"`
	Image string 		`json : "image"`
	Stock int 			`json : "stock"`
}

// SnipcartProductResponse struct
type SnipcartProductResponse struct {
	Items []SnipcartProduct `json : "items"`
}

// SnipcartProduct struct
type SnipcartProduct struct {
	Stock int `json : "stock"`
}

func (dest *HugoProduct) mapPonzuProduct(src PonzuProduct, ponzuHostURL string, client *http.Client) {
	dest.ID = src.Slug
	dest.Title = src.Name
	dest.Price = src.Price
	dest.Description = src.Description
	dest.Image = ponzuHostURL + src.Image
	dest.Date = time.Unix(src.Timestamp/1000, 0)
	dest.LastModification = time.Unix(src.Updated/1000, 0)

	// Fetch stock from snipcart api
	var url = "https://app.snipcart.com/api/products?userDefinedId=" + dest.ID
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	//var apiKey = base64.StdEncoding.EncodeToString([]byte(os.Getenv("SNIPCART_PRIVATE_API_KEY")))
	var apiKey = base64.StdEncoding.EncodeToString([]byte(os.Getenv("SNIPCART_PRIVATE_API_KEY")))

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Authorization", "Basic" + apiKey)
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var products SnipcartProductResponse
	if err = json.Unmarshal(body, &products); err != nil {
		log.Fatal(err)
	}
	dest.Stock = products.Items[0].Stock
}

func main(){
	ponzuHostURL, ok := os.LookupEnv("PONZU_HOST_URL")

	if !ok || ponzuHostURL == "" {

	}

	var productsEndpoint = ponzuHostURL + "api/contents?type=Product"

	// Fetch products 
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true}, // this line should be removed in production
	}

	client := &http.Client{Transport : tr}

	request, err := http.NewRequest(http.MethodGet, productsEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Parse response json
	var products PonzuProductResponse
	if err = json.Unmarshal(body, &products); err != nil {
		log.Fatal(err)
	}

	// clear content/product directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.RemoveAll(dir + "/content/product"); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(dir + "/content/product", 0777); err != nil {
		log.Fatal(err)
	}

	// write product content files
	for _, sourceProduct := range products.Data {
		var destProduct = HugoProduct{}
		destProduct.mapPonzuProduct(sourceProduct, ponzuHostURL, client)
		entryJSON, err := json.MarshalIndent(destProduct, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(dir + "/content/product/" + destProduct.ID + ".md")
		if err != nil {
			log.Fatal(err)
		}
		writer := bufio.NewWriter(file)
		writer.WriteString(string(entryJSON) + "\n")
		writer.WriteString("\n")
		writer.WriteString(destProduct.Description)
		writer.Flush()
		file.Close()
	}



}

/*
run ponzu run --dev-https
https://github.com/snipcart/ponzu-hugo-snipcart/blob/master/myshop.com/themes/beautifulhugo/layouts/index.html
*/