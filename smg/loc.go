package smg

import (
	"encoding/xml"
	"time"
)

// SitemapLoc contains data related to <url> tag in Sitemap.
type SitemapLoc struct {
	XMLName    xml.Name               `xml:"url"`
	Loc        string                 `xml:"loc"`
	LastMod    *time.Time             `xml:"lastmod,omitempty"`
	ChangeFreq ChangeFreq             `xml:"changefreq,omitempty"`
	Priority   float32                `xml:"priority,omitempty"`
	Images     []*SitemapImage        `xml:"image:image,omitempty"`
	Alternate  []*SitemapAlternateLoc `xml:"xhtml:link,omitempty"`
}

// SitemapImage contains data related to <image:image> tag in Sitemap <url>
type SitemapImage struct {
	ImageLoc string `xml:"image:loc,omitempty"`
}

// SitemapIndexLoc contains data related to <sitemap> tag in SitemapIndex.
type SitemapIndexLoc struct {
	XMLName xml.Name   `xml:"sitemap"`
	Loc     string     `xml:"loc"`
	LastMod *time.Time `xml:"lastmod,omitempty"`
}

// SitemapAlternateLoc contains data related to <xhtml:link> tag in Sitemap <url>
type SitemapAlternateLoc struct {
	Hreflang string `xml:"hreflang,attr"`
	Href     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr"`
}
