package smg

import (
	"encoding/xml"
	"time"
)

// LastModValue holds either a *time.Time or a string for the lastmod XML element.
type LastModValue struct {
	t   *time.Time
	str string
}

// LastModTime creates a LastModValue from a *time.Time.
func LastModTime(t *time.Time) *LastModValue {
	return &LastModValue{t: t}
}

// LastModString creates a LastModValue from a string.
func LastModString(s string) *LastModValue {
	return &LastModValue{str: s}
}

// MarshalXML implements xml.Marshaler. Formats time using RFC3339Nano when set from *time.Time.
func (lm LastModValue) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if lm.t != nil {
		b, err := lm.t.MarshalText()
		if err != nil {
			return err
		}
		return e.EncodeElement(string(b), start)
	}
	return e.EncodeElement(lm.str, start)
}

// SitemapLoc contains data related to <url> tag in Sitemap.
type SitemapLoc struct {
	XMLName    xml.Name               `xml:"url"`
	Loc        string                 `xml:"loc"`
	LastMod    *LastModValue          `xml:"lastmod,omitempty"`
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
