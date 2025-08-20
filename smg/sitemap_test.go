package smg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type UrlSet struct {
	XMLName xml.Name  `xml:"urlset"`
	Urls    []UrlData `xml:"url"`
}

type UrlData struct {
	XMLName    xml.Name               `xml:"url"`
	Loc        string                 `xml:"loc"`
	LasMod     string                 `xml:"lastmod"`
	ChangeFreq string                 `xml:"changefreq"`
	Priority   string                 `xml:"priority"`
	Images     []SitemapImageData     `xml:"image"`
	Alternate  []SitemapAlternateData `xml:"xhtml link"`
}

type SitemapImageData struct {
	ImageLoc string `xml:"loc,omitempty"`
}

type SitemapAlternateData struct {
	Hreflang string `xml:"hreflang,attr"`
	Href     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr"`
}

// TestSingleSitemap tests the module against Single-file sitemap usage format.
func TestSingleSitemap(t *testing.T) {
	path := t.TempDir()
	now := time.Now().UTC()
	routes := buildRoutes(10, 40, 10)

	sm := NewSitemap(true)
	sm.SetName("single_sitemap")
	sm.SetHostname(baseURL)
	sm.SetOutputPath(path)
	sm.SetLastMod(&now)
	sm.SetCompress(false)

	for _, route := range routes {
		err := sm.Add(&SitemapLoc{
			Loc:        route,
			LastMod:    &now,
			ChangeFreq: Always,
			Priority:   0.4,
			Images:     []*SitemapImage{{"path-to-image.jpg"}},
		})
		if err != nil {
			t.Fatal("Unable to add SitemapLoc:", err)
		}
	}
	// -----------------------------------------------------------------

	// Compressed files;
	filenames, err := sm.Save()
	if err != nil {
		t.Fatal("Unable to Save Compressed Sitemap:", err)
	}
	for _, filename := range filenames {
		assertOutputFile(t, path, filename)
	}

	// Plain files:
	sm.SetCompress(false)
	filenames, err = sm.Save()
	if err != nil {
		t.Fatal("Unable to Save Sitemap:", err)
	}
	for _, filename := range filenames {
		assertOutputFile(t, path, filename)
	}
}

// TestSitemapAdd tests that the Add function produces a proper URL
func TestSitemapAdd(t *testing.T) {
	path := t.TempDir()
	testLocation := "/test?foo=bar"
	testImage := "/path-to-image.jpg"
	testImage2 := "/path-to-image-2.jpg"
	testAlternate1 := &SitemapAlternateLoc{
		Hreflang: "en",
		Href:     fmt.Sprintf("%s%s", baseURL, "/en/test"),
		Rel:      "alternate",
	}
	testAlternate2 := &SitemapAlternateLoc{
		Hreflang: "de",
		Href:     fmt.Sprintf("%s%s", baseURL, "/de/test"),
		Rel:      "alternate",
	}
	now := time.Now().UTC()

	sm := NewSitemap(true)
	sm.SetName("single_sitemap")
	sm.SetHostname(baseURL)
	sm.SetOutputPath(path)
	sm.SetLastMod(&now)
	sm.SetCompress(false)

	err := sm.Add(&SitemapLoc{
		Loc:        testLocation,
		LastMod:    &now,
		ChangeFreq: Always,
		Priority:   0.4,
		Images:     []*SitemapImage{{testImage}, {testImage2}},
		Alternate:  []*SitemapAlternateLoc{testAlternate1, testAlternate2},
	})
	if err != nil {
		t.Fatal("Unable to add SitemapLoc:", err)
	}
	expectedUrl := fmt.Sprintf("%s%s", baseURL, testLocation)
	expectedImage := fmt.Sprintf("%s%s", baseURL, testImage)
	expectedImage2 := fmt.Sprintf("%s%s", baseURL, testImage2)
	filepath, err := sm.Save()
	if err != nil {
		t.Fatal("Unable to Save Sitemap:", err)
	}

	xmlFile, err := os.Open(fmt.Sprintf("%s/%s", path, filepath[0]))
	if err != nil {
		t.Fatal("Unable to open file:", err)
	}
	defer xmlFile.Close()
	byteValue, _ := io.ReadAll(xmlFile)
	var urlSet UrlSet
	err = xml.Unmarshal(byteValue, &urlSet)
	if err != nil {
		t.Fatal("Unable to unmarhsall sitemap byte array into xml: ", err)
	}
	actualUrl := urlSet.Urls[0].Loc
	assert.Equal(t, expectedUrl, actualUrl)

	actualImage := urlSet.Urls[0].Images[0].ImageLoc
	assert.Equal(t, expectedImage, actualImage)

	actualImage2 := urlSet.Urls[0].Images[1].ImageLoc
	assert.Equal(t, expectedImage2, actualImage2)

	for i, expAlter := range []*SitemapAlternateLoc{testAlternate1, testAlternate2} {
		actualAlternate := urlSet.Urls[0].Alternate[i].Href
		assert.Equal(t, expAlter.Href, actualAlternate)

		actualAlternate = urlSet.Urls[0].Alternate[i].Rel
		assert.Equal(t, expAlter.Rel, actualAlternate)

		actualAlternate = urlSet.Urls[0].Alternate[i].Hreflang
		assert.Equal(t, expAlter.Hreflang, actualAlternate)
	}
}

func TestWriteTo(t *testing.T) {
	path := t.TempDir()
	now := time.Now().UTC()
	testLocation := "/test/"

	sm := NewSitemap(true)
	sm.SetHostname(baseURL)
	sm.SetLastMod(&now)
	sm.SetOutputPath(path)
	sm.SetCompress(false)

	err := sm.Add(&SitemapLoc{
		Loc:        testLocation,
		LastMod:    &now,
		ChangeFreq: Always,
		Priority:   0.4,
		Images:     []*SitemapImage{{"path-to-image.jpg"}},
	})
	if err != nil {
		t.Fatal("Unable to add SitemapLoc:", err)
	}
	sm.Finalize()

	//----- Write to buffer

	buf := bytes.Buffer{}
	_, err = sm.WriteTo(&buf)
	if err != nil {
		t.Fatal("Unable to write to buffer:", err)
	}
	//-----
	expectedUrl := fmt.Sprintf("%s%s", baseURL, testLocation)

	var urlSet UrlSet
	err = xml.Unmarshal(buf.Bytes(), &urlSet)
	if err != nil {
		t.Fatal("Unable to unmarhsall sitemap byte array into xml: ", err)
	}
	actualUrl := urlSet.Urls[0].Loc
	assert.Equal(t, expectedUrl, actualUrl)
}
