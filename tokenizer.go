package main

import (
	"bytes"
	"container/heap"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

const (
	// tags
	LinkTag    = "a"
	Header1Tag = "h1"
	ImageTag   = "img"
	SpanTag    = "span"

	// attributes
	Href   = "href"
	Class  = "class"
	Source = "src"
)

func GetAllProductsFromUrl(url string) [][]string {
	urlPool := make(PriorityQueue, 1)
	urlPool[0] = &Item{
		Value:    url,
		Priority: 1,
		Index:    0,
	}
	heap.Init(&urlPool)

	visited := NewSet[string]()
	products := [][]string{{"Url", "Image", "Name", "Price"}}

	for len(urlPool) > 0 {
		urlItem := heap.Pop(&urlPool).(*Item)
		visited.Add(urlItem.Value)

		resp, err := http.Get(urlItem.Value)
		if err != nil {
			log.Fatal("Error with request: ", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error with reading body: ", err)
		}

		urls := getUrlsFromHtml(bytes.NewReader(body))
		product := getProductFromHtml(bytes.NewReader(body), urlItem.Value)

		for _, url := range urls {
			if !visited.Contains(url) && !urlPool.ContainsValue(url) {
				r, _ := regexp.Compile(`^https://scrapeme\.live/shop/page/\d+/?$`)
				if r.MatchString(url) {
					heap.Push(&urlPool, &Item{Value: url, Priority: 2})
				} else {
					heap.Push(&urlPool, &Item{Value: url, Priority: 1})
				}
			}
		}
		if product != nil {
			products = append(products, product)
		}
	}
	return products
}

func getUrlsFromHtml(body io.Reader) []string {
	t := html.NewTokenizer(body)
	urls := make([]string, 0)
	for {
		tt := t.Next()
		switch tt {
		case html.ErrorToken:
			return urls
		case html.StartTagToken:
			token := t.Token()
			if token.Data == LinkTag {
				for _, attr := range token.Attr {
					if attr.Key == Href {
						if url := attr.Val; strings.Contains(url, "https://scrapeme.live/shop") {
							urls = append(urls, url)
						}
					}
				}
			}
		}
	}
}

func getProductFromHtml(body io.Reader, currUrl string) []string {
	t := html.NewTokenizer(body)
	var image, name, price string
	priceDepth := 0

	for {
		if image != "" && name != "" && price != "" {
			return []string{currUrl, image, name, price}
		}

		tt := t.Next()
		switch tt {
		case html.ErrorToken:
			return nil

		case html.StartTagToken, html.SelfClosingTagToken:
			token := t.Token()
			switch token.Data {
			case Header1Tag:
				for _, attr := range token.Attr {
					if attr.Key == Class {
						if class := attr.Val; strings.Contains(class, "product_title") {
							if nextTT := t.Next(); nextTT == html.TextToken {
								name = t.Token().Data
							}
						}
					}
				}

			case ImageTag:
				attributes := make(map[string]string)
				for _, attr := range token.Attr {
					attributes[attr.Key] = attr.Val
				}
				if strings.Contains(attributes[Class], "wp-post-image") {
					image = attributes[Source]
				}

			case SpanTag:
				for _, attr := range token.Attr {
					if class := attr.Val; class == "woocommerce-Price-currencySymbol" {
						priceDepth += 1
						if priceDepth == 2 {
							if nextTT := t.Next(); nextTT == html.TextToken {
								price = t.Token().Data
								t.Next()
							}
							if nextTT := t.Next(); nextTT == html.TextToken {
								price = price + t.Token().Data
							}
						}
					}
				}
			}
		}
	}
}