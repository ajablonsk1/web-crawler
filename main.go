package main

func main() {
	products := GetAllProductsFromUrl("https://scrapeme.live/shop")
	SaveResultsToFile(products)
}
