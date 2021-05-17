package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var obiBaseUrl = "https://www.obi.ru"
var products []*Product

func main() {
	path := ("")
	fmt.Println(path)
	collector := GetCollector()
	responsesCounter := 0

	collector.OnError(func(cr *colly.Response, err error) {
		fmt.Println("Response error: ", cr.Request.URL.String())
		fmt.Println(cr.StatusCode)
	})

	collector.OnResponse(func(rese *colly.Response) {
		responsesCounter++
		if responsesCounter > 10 {
			ProductsToYml(products, path)
			responsesCounter = 0
		}
	})

	// Categories
	collector.OnHTML("ul#First-Level > li:nth-child(-n+6) li.span4 > a", func(he *colly.HTMLElement) {
		if he.Request.URL.String() != "https://www.obi.ru" {
			return
		}
		AddVisit(collector, he.Attr("href"))
	})

	// Category pages check and add to visit
	collector.OnHTML("button.pagination-bar__link > span", func(he *colly.HTMLElement) {
		page := he.Request.URL.Query().Get("page")
		if len(page) > 0 {
			return
		}

		pagesCountString := he.Text
		pagesCount, err := strconv.Atoi(pagesCountString)
		if err != nil {
			failOnError(err, "Failed convert pages count to int")
		}
		if pagesCount > 1 {
			for i := 2; i <= pagesCount; i++ {
				newUrl := he.Request.URL.String() + "?page=" + strconv.Itoa(i)
				AddVisit(collector, newUrl)
			}
		}
	})

	// Sub categories
	collector.OnHTML("div.categoryitem > *[wt_name='assortment_tile']", func(he *colly.HTMLElement) {
		AddVisit(collector, he.Attr("href"))
	})

	// Category page products
	collector.OnHTML("ul.products-wp > li.product > a.product-wrapper", func(he *colly.HTMLElement) {

		noProduct := strings.HasPrefix(he.Attr("href"), "https://")
		if noProduct {
			return
		}

		productUrl := obiBaseUrl + he.Attr("href")
		productBaseUrl := productUrl[:len(obiBaseUrl)]

		if productBaseUrl != obiBaseUrl{
			productUrl = obiBaseUrl + he.Attr("href")
		}

		AddVisit(collector, productUrl)
	})

	// Product
	collector.OnHTML("body", func(ce *colly.HTMLElement) {
		dom := ce.DOM
		article := dom.Find("p.article-number").Text()
		if len(article) > 0 {
			article = article[19:]
			product := &Product{}
			product.SetAvailable(dom)

			if !product.Available {
				return
			}

			product.Id = article
			product.Url = ce.Request.URL.String()
			product.Article = article
			product.Brand = dom.Find("form.order-details div.logo img").AttrOr("title", "")
			product.Name = dom.Find("h1.h2").Text()
			product.Price = dom.Find("*[data-ui-name='ads.price.strong']").Text()
			product.OldPrice = dom.Find("#AB_radio_wrapper del:nth-child(2)").Text()
			product.SetCategory(dom)
			product.SetAvailableCount(dom)
			product.SetDescription(dom)
			product.SetImages(dom)
			products = append(products, product)
		}
	})

	err := collector.Visit(obiBaseUrl)
	failOnError(err, "Base visit error.")

	collector.Wait()
	ProductsToYml(products, path)
}

func AddVisit(collector *colly.Collector, link string) {
	visited, err := collector.HasVisited(link)
	if err != nil {
		failOnError(err, "Check visited error")
	}

	if !visited {
		err := collector.Visit(link)
		if err != nil {
			failOnError(err, "Add to visit error !")
		}
	}
}

func GetCollector() *colly.Collector {
	collector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"),
		colly.MaxDepth(5),
		colly.AllowURLRevisit(),
		//colly.Async(true),
	)

	collector.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second, // timeout
		}).DialContext,
		MaxIdleConns:          1,                // Maximum number of idle connections
		IdleConnTimeout:       30 * time.Second, // Idle connection timeout
		TLSHandshakeTimeout:   30 * time.Second, // TLS handshake timeout
		ExpectContinueTimeout: 1 * time.Second,
	})

	err := collector.Limit(&colly.LimitRule{
		DomainGlob:  "*httpbin.*",
		Parallelism: 2,
		Delay:       30 * time.Second,
	})

	if err != nil {
		failOnError(err, "make collector error")
	}

	return collector
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Println("ERROR !!!")
		log.Println(msg+": ", err)
	}
}