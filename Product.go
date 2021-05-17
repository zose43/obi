package main

import (
	"encoding/xml"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type Image string

type Product struct {
	XMLName        xml.Name `xml:"offer"`
	Id             string   `xml:"id,attr"`
	Available      bool     `xml:"available,attr"`
	Url            string   `xml:"url"`
	AvailableCount string   `xml:"available_count"`
	Article        string   `xml:"article"`
	Brand          string   `xml:"brand,omitempty"`
	Name           string   `xml:"name"`
	Description    string   `xml:"description"`
	Price          string   `xml:"price"`
	OldPrice	   string	`xml:"old_price"`
	CategoryId     uint32   `xml:"categoryId"`
	CategoryName   string   `xml:"-"`
	Images         []Image  `xml:"images>image"`
}

func (p *Product) SetImages(dom *goquery.Selection) {
	dom.Find("img.ads-slider__image").Each(func(_ int, s *goquery.Selection) {
		imgLink := s.AttrOr("data-bigpic", "")
		if len(imgLink) <= 0 {
			return
		}
		cleanLink := imgLink[2:]
		p.Images = append(p.Images, Image("https://"+cleanLink))
	})
}

func (p *Product) SetDescription(dom *goquery.Selection) {
	dom.Find("div.description-text p").Each(func(i int, s *goquery.Selection) {
		if i >= 1 {
			p.Description += s.Text()
		}
	})
}

func (p *Product) SetAvailable(gs *goquery.Selection) {
	acs := gs.Find("*[data-ui-name='instore.adp.availability_message']").Text()
	if acs == "Нет в наличии" {
		p.Available = false
	} else {
		p.Available = true
	}
}

func (p *Product) SetAvailableCount(gs *goquery.Selection) {
	p.AvailableCount = "0"
	acs := gs.Find("*[data-ui-name='instore.adp.availability_message']").Text()
	endOfCountSubString := strings.Index(acs, " шт")
	if endOfCountSubString > 0 {
		p.AvailableCount = acs[:endOfCountSubString]
	}
}

func (p *Product) SetCategory(dom *goquery.Selection) {
	p.CategoryName = dom.Find("section.breadcrumb > ul > li:last-child a span").Text()
	p.CategoryId = MakeCategoryId(p.CategoryName)
}