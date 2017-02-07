package wordpress

/*
 * http://www.apache.org/licenses/LICENSE-2.0.txt
 *
 * Copyright 2017 OpsVision Solutions
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Page represents a WordPress page
type Page struct {
	Date          string  `json:"date"`
	DateGmt       string  `json:"date_gmt"`
	GUID          GUID    `json:"guid"`
	ID            int     `json:"id"`
	Link          string  `json:"link"`
	Modified      string  `json:"modified"`
	ModifiedGmt   string  `json:"modified_gmt"`
	Slug          string  `json:"slug"`
	Status        string  `json:"status"`
	Type          string  `json:"type"`
	Parent        int     `json:"parent"`
	Title         Title   `json:"title"`
	Content       Content `json:"content"`
	Author        int     `json:"author"`
	Excerpt       Excerpt `json:"excerpt"`
	FeaturedMedia int     `json:"featured_media"`
	CommentStatus string  `json:"comment_status"`
	PingStatus    string  `json:"ping_status"`
	MenuOrder     int     `json:"menu_order"`
	Meta          Metas   `json:"meta"`
	Template      string  `json:"template"`
}

// Pages is a collection of Page objects
type Pages []Page

// GetPageMetrics calculates the time in milliseconds to load the page and resources
func (p *Page) GetPageMetrics() Metric {
	var wg sync.WaitGroup
	var links []string
	var start time.Time
	var pLoad, rLoad, tLoad float64

	// Get the page's contents
	start = time.Now()
	content := p.getPageContents(p.Link)
	pLoad = time.Since(start).Seconds() * 1e3

	// Create a page query
	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		fmt.Println(err)
	}

	// Get Links
	p.getCSSLinks(&links, doc)
	p.getJSLinks(&links, doc)
	p.getIMGLinks(&links, doc)

	// Set the size of our WaitGroup
	wg.Add(len(links))

	// Process the Links concurrently
	start = time.Now()
	for _, l := range links {
		go func(l string) {
			defer wg.Done()
			getClient().Get(l)
		}(l)
	}

	// Wait for the threads to finish
	wg.Wait()

	// Calculate the resource and total load time
	rLoad = time.Since(start).Seconds() * 1e3
	tLoad = pLoad + rLoad

	// Store the metrics and return them
	return Metric{
		Page:         p.Slug,
		PageLoad:     pLoad,
		ResourceLoad: rLoad,
		TotalLoad:    tLoad,
	}
}

// GetPages retrieves an array of Page objects from a WordPress site
func GetPages(site string) (Pages, error) {
	var pages Pages
	var buff bytes.Buffer

	// Build the REST endpoint URL
	fmt.Fprintf(&buff, "%s/wp-json/wp/v2/pages", site)

	// Fetch the pages
	resp, err := getClient().Get(buff.String())
	if err != nil {
		return pages, err
	}
	defer resp.Body.Close()

	// Make sure we got a status OK
	if resp.StatusCode != http.StatusOK {
		return pages, fmt.Errorf("%s", resp.Status)
	}

	// Parse the JSON
	err = json.NewDecoder(resp.Body).Decode(&pages)
	if err != nil {
		return pages, err
	}

	return pages, nil
}

// getCSSLinks extracts links for external CSS stylesheets
func (p *Page) getCSSLinks(links *[]string, doc *goquery.Document) {
	// CSS links
	doc.Find("link").Each(func(index int, item *goquery.Selection) {
		rel, _ := item.Attr("rel")
		if rel == "stylesheet" {
			link, _ := item.Attr("href")
			*links = append(*links, link)
		}
	})
}

// getJSLinks extracts links for external JavaScript
func (p *Page) getJSLinks(links *[]string, doc *goquery.Document) {
	// Script Links
	doc.Find("script").Each(func(index int, item *goquery.Selection) {
		if _, ok := item.Attr("type"); ok {
			if src, ok := item.Attr("src"); ok {
				*links = append(*links, src)
			}
		}
	})
}

// getIMGLinks extracts links for IMG tags
func (p *Page) getIMGLinks(links *[]string, doc *goquery.Document) {
	// Images
	doc.Find("img").Each(func(index int, item *goquery.Selection) {
		if src, ok := item.Attr("src"); ok {
			*links = append(*links, src)
		}
	})
}

// Fetch the page contents
func (p *Page) getPageContents(url string) *bytes.Buffer {
	var out bytes.Buffer

	// Get the page
	resp, err := getClient().Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// Copy the contents
	io.Copy(&out, resp.Body)

	return &out
}

// Get a configured HTTP client
func getClient() *http.Client {
	timeout := 5 * time.Second

	// Setup transport settings
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout,
	}

	// Create a client
	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	return client
}
