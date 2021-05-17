package main

import (
	"encoding/xml"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"time"
)

type Category struct {
	Id   uint32 `xml:"id,attr"`
	Name string `xml:"name"`
}

type Offers struct {
	Offers []*Product `xml:"offers"`
}

type Shop struct {
	XMLName    xml.Name   `xml:"shop"`
	Name       string     `xml:"name"`
	Company    string     `xml:"company"`
	Url        string     `xml:"url"`
	Offers     Offers     `xml:"offers"`
	Categories []Category `xml:"categories>category"`
}

type YmlCatalog struct {
	XMLName xml.Name `xml:"yml_catalog"`
	Date    string   `xml:"date,attr"`
	Shop    Shop     `xml:"shop"`
}

func ProductsToYml(products []*Product, path string) {
	offers := Offers{Offers: products}
	//categories := Categories{getCategories(products)}
	yc := YmlCatalog{
		Date: time.Now().String(),
		Shop: Shop{
			Name:       "ОБИ",
			Company:    "ОБИ",
			Url:        "https://www.obi.ru/",
			Offers:     offers,
			Categories: getCategories(products),
		},
	}
	output, err := xml.MarshalIndent(yc, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	output = []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + string(output))
	fullFilePath := path + "Obi.xml"
	err1 := ioutil.WriteFile(fullFilePath, output, 0644)
	if err != nil {
		log.Fatal(err1)
	}
}

func getCategories(products []*Product) (categories []Category) {
	for _, p := range products {
		categories = appendIfMissing(categories, Category{
			Id:   MakeCategoryId(p.CategoryName),
			Name: p.CategoryName,
		})
	}
	return categories
}

func appendIfMissing(slice []Category, i Category) []Category {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func MakeCategoryId(categoryName string) (id uint32) {
	return hash(categoryName)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
