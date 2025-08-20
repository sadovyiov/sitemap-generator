package smg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"
)

// ChangeFreq is used for defining changefreq property in sitemap url items.
type ChangeFreq string

// predefined ChangeFreq frequency values
const (
	Always  ChangeFreq = "always"
	Hourly  ChangeFreq = "hourly"
	Daily   ChangeFreq = "daily"
	Weekly  ChangeFreq = "weekly"
	Monthly ChangeFreq = "monthly"
	Yearly  ChangeFreq = "yearly"
	Never   ChangeFreq = "never"
)

const (
	fileExt             string = ".xml"
	fileGzExt           string = ".xml.gz"
	maxFileSize         int    = 52428000 // decreased 800 byte to prevent a small bug to fail a big program :)
	defaultMaxURLsCount int    = 50000
	xmlUrlsetOpenTag    string = `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1" xmlns:xhtml="http://www.w3.org/1999/xhtml">`
	xmlUrlsetCloseTag   string = "</urlset>\n"
)

// Sitemap struct which contains Options for general attributes,
// SitemapLoc as its location in SitemapIndex, NextSitemap that is
// a Linked-List pointing to the next Sitemap for large files.
type Sitemap struct {
	Options
	SitemapIndexLoc *SitemapIndexLoc
	NextSitemap     *Sitemap
	maxURLsCount    int
	fileNum         int
	urlsCount       int
	content         bytes.Buffer
	tempBuf         *bytes.Buffer
	xmlEncoder      *xml.Encoder
	isFinalized     bool
}

// NewSitemap builds and returns a new Sitemap.
func NewSitemap(prettyPrint bool) *Sitemap {
	t := time.Now().UTC()

	s := &Sitemap{
		SitemapIndexLoc: &SitemapIndexLoc{
			LastMod: &t,
		},
	}
	s.Compress = true
	s.prettyPrint = prettyPrint
	s.content = bytes.Buffer{}
	s.content.Write([]byte(xml.Header))
	s.content.Write([]byte(xmlUrlsetOpenTag))
	s.tempBuf = &bytes.Buffer{}
	s.Name = "sitemap"
	s.maxURLsCount = defaultMaxURLsCount
	s.xmlEncoder = xml.NewEncoder(s.tempBuf)
	if prettyPrint {
		s.content.Write([]byte{'\n'})
		s.xmlEncoder.Indent("", "  ")
	}
	return s
}

// Add adds an URL to a Sitemap.
// in case of exceeding the Sitemaps.org limits, splits the Sitemap
// into several Sitemap instances using a Linked List
func (s *Sitemap) Add(u *SitemapLoc) error {
	if s.isFinalized {
		return fmt.Errorf("sitemap is finalized")
	}
	return s.realAdd(u, 0, nil)
}

func (s *Sitemap) realAdd(u *SitemapLoc, locN int, locBytes []byte) error {
	if s.NextSitemap != nil {
		s.NextSitemap.realAdd(u, locN, locBytes)
		return nil
	}

	if s.urlsCount >= s.maxURLsCount {
		s.buildNextSitemap()
		return s.NextSitemap.realAdd(u, locN, locBytes)
	}

	if len(u.Images) > 0 {
		for _, image := range u.Images {
			output, err := url.Parse(s.Hostname)
			if err != nil {
				return err
			}
			output.Path = path.Join(output.Path, image.ImageLoc)
			image.ImageLoc = output.String()
		}
	}

	if locBytes == nil {
		output, err := url.Parse(s.Hostname)
		if err != nil {
			return err
		}
		loc, err := url.Parse(u.Loc)
		if err != nil {
			return err
		}
		u.Loc = output.ResolveReference(loc).String()
		locN, locBytes, err = s.encodeToXML(u)
		if err != nil {
			return err
		}
	}

	if locN+s.content.Len() >= maxFileSize {
		s.buildNextSitemap()
		return s.NextSitemap.realAdd(u, locN, locBytes)
	}

	_, err := s.content.Write(locBytes)
	if err != nil {
		return err
	}
	s.urlsCount++
	return nil
}

// buildNextSitemap builds a new Sitemap instance based on current one
// and connects to it via NextSitemap.
func (s *Sitemap) buildNextSitemap() {
	s.NextSitemap = NewSitemap(s.prettyPrint)
	s.NextSitemap.Compress = s.Compress
	s.NextSitemap.Name = s.Name
	s.NextSitemap.Hostname = s.Hostname
	s.NextSitemap.OutputPath = s.OutputPath
	s.NextSitemap.maxURLsCount = s.maxURLsCount
	s.NextSitemap.fileNum = s.fileNum + 1
}

func (s *Sitemap) encodeToXML(loc *SitemapLoc) (int, []byte, error) {
	err := s.xmlEncoder.Encode(loc)
	if err != nil {
		return 0, nil, err
	}
	defer s.tempBuf.Reset()
	return s.tempBuf.Len(), s.tempBuf.Bytes(), nil
}

// SetName sets the Name of Sitemap output xml file
// It must be without ".xml" extension
func (s *Sitemap) SetName(name string) {
	s.Name = name
	if s.NextSitemap != nil {
		s.NextSitemap.SetName(name)
	}
}

// SetHostname sets the Hostname of Sitemap urls which be prepended to all URLs.
// Note: you do not have to call SetHostname in case you are building Sitemap using SitemapIndex.NewSitemap
// but you can set a separate Hostname for a specific Sitemap using SetHostname,
// else the SitemapIndex.SetHostname does this action for all Sitemaps of the entire SitemapIndex.
func (s *Sitemap) SetHostname(hostname string) {
	s.Hostname = hostname
	if s.NextSitemap != nil {
		s.NextSitemap.SetHostname(hostname)
	}
}

// SetOutputPath sets the OutputPath of Sitemap which will be used to save the xml file.
// Note: you do not have to call SetOutputPath in case you are building Sitemap using SitemapIndex.NewSitemap
// but you can set a separate OutputPath for a specific Sitemap using SetOutputPath,
// else the SitemapIndex.SetOutputPath does this action for all Sitemaps of the entire SitemapIndex.
func (s *Sitemap) SetOutputPath(outputPath string) {
	s.OutputPath = outputPath
	if s.NextSitemap != nil {
		s.NextSitemap.SetOutputPath(outputPath)
	}
}

// SetLastMod sets the LastMod if this Sitemap which will be used in it's URL in SitemapIndex
func (s *Sitemap) SetLastMod(lastMod *time.Time) {
	s.SitemapIndexLoc.LastMod = lastMod
	if s.NextSitemap != nil {
		s.NextSitemap.SetLastMod(lastMod)
	}
}

// SetCompress sets the Compress option to be either enabled or disabled for Sitemap
// When Compress is enabled, the output file is compressed using gzip with .xml.gz extension.
func (s *Sitemap) SetCompress(compress bool) {
	s.Compress = compress
	if s.NextSitemap != nil {
		s.NextSitemap.SetCompress(compress)
	}
}

// SetMaxURLsCount sets the maximum # of URLs for a sitemap
func (s *Sitemap) SetMaxURLsCount(maxURLsCount int) {
	s.maxURLsCount = maxURLsCount
}

// GetURLsCount returns the number of added URL items into this single sitemap.
func (s *Sitemap) GetURLsCount() int {
	return s.urlsCount
}

// Finalize closes the XML data set and do not allow any further sm.Add() calls
func (s *Sitemap) Finalize() {
	if s.prettyPrint {
		s.content.Write([]byte{'\n'})
	}
	s.content.Write([]byte(xmlUrlsetCloseTag))

	s.isFinalized = true
}

// Save makes the OutputPath in case of absence and saves the Sitemap into OutputPath using it's Name.
// it returns the filename.
func (s *Sitemap) Save() (filenames []string, err error) {
	err = checkAndMakeDir(s.OutputPath)
	if err != nil {
		return
	}

	// Appends the fileNum at the end of filename in case of more than 0 (it is extended Sitemap)
	var filename string
	if s.fileNum > 0 {
		filename = fmt.Sprintf("%s%d", s.Name, s.fileNum)
	} else {
		filename = s.Name
	}

	if s.Compress {
		filename += fileGzExt
	} else {
		filename += fileExt
	}

	if !s.isFinalized {
		s.Finalize()
	}

	_, err = writeToFile(filename, s.OutputPath, s.Compress, s.content.Bytes())
	if err != nil {
		return
	}

	if s.NextSitemap != nil {
		filenames, err = s.NextSitemap.Save()
		if err != nil {
			return nil, err
		}
	}
	return append(filenames, filename), nil
}

func (s *Sitemap) WriteTo(w io.Writer) (n int64, err error) {
	return s.content.WriteTo(w)
}
