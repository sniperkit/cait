//
// sitemapper generates a sitemap.xml file by crawling the content generate with genpages
//
// @author R. S. Doiel, <rsdoiel@caltech.edu>
//
// Copyright (c) 2016, Caltech
// All rights not granted herein are expressly reserved by Caltech.
//
// Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	// Caltech Library packages
	"github.com/caltechlibrary/cait"
	"github.com/caltechlibrary/cli"
)

type locInfo struct {
	Loc     string
	LastMod string
}

var (
	usage = `USAGE: %s [OPTIONS]`

	description = `
OVERVIEW

%s generates a sitemap for the website.

CONFIGURATION

+ CAIT_HTDOCS is the path to the HTML document root
+ CAIT_SITE_URL is the base path for the URL to website (e.g. http://localhost:8001)
+ CAIT_SITEMAP is the path to the sitemap xml file (e.g. htdocs/sitemap.xml)
`

	examples = `
EXAMPLE

    %s -htdocs htdocs \
	   -sitemap htdocs/sitemap.xml \
	   -url http://archives.example.org
`

	license = `
%s %s

Copyright (c) 2016, Caltech
All rights not granted herein are expressly reserved by Caltech.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
`
	// Standard cli options
	showHelp    bool
	showVersion bool
	showLicense bool

	// App specific options
	htdocs   string
	sitemap  string
	siteURL  string
	excluded string

	changefreq string
	locList    []*locInfo
)

func init() {
	// We are going to log to standard out rather than standard err
	log.SetOutput(os.Stdout)

	// Setup options
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "v", false, "display version")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "l", false, "display license")
	flag.BoolVar(&showLicense, "license", false, "display license")

	// App specific options
	flag.StringVar(&htdocs, "htdocs", "", "Path to website directory (e.g. www or htdocs)")
	flag.StringVar(&siteURL, "url", "", "Basepath for website")
	flag.StringVar(&sitemap, "sitemap", "", "path to sitemap file (e.g. htdocs/sitemap.xml)")
	flag.StringVar(&changefreq, "u", "daily", "Set the change frequencely value, e.g. daily, weekly, monthly")
	flag.StringVar(&changefreq, "update-frequency", "daily", "Set the change frequencely value, e.g. daily, weekly, monthly")
	flag.StringVar(&excluded, "e", "", "A colon delimited list of path parts to exclude from sitemap")
	flag.StringVar(&excluded, "exclude", "", "A colon delimited list of path parts to exclude from sitemap")
}

// IsExcluded returns true if a fname fragment is included in set of dirList
func IsExcluded(el []string, p string) bool {
	if len(el) == 0 {
		return false
	}
	for _, item := range el {
		if strings.Contains(p, item) == true {
			log.Printf("Skipping %q, because %q", p, item)
			return true
		}
	}
	return false
}

func main() {
	appName := path.Base(os.Args[0])
	flag.Parse()

	cfg := cli.New(appName, "CAIT", fmt.Sprintf(license, appName, cait.Version), cait.Version)
	cfg.UsageText = fmt.Sprintf(usage, appName)
	cfg.DescriptionText = fmt.Sprintf(description, appName)
	cfg.OptionsText = "OPTIONS\n"
	cfg.ExampleText = fmt.Sprintf(examples, appName)

	if showHelp == true {
		fmt.Println(cfg.Usage())
		os.Exit(0)
	}
	if showVersion == true {
		fmt.Println(cfg.Version())
		os.Exit(0)
	}
	if showLicense == true {
		fmt.Println(cfg.License())
		os.Exit(0)
	}

	if changefreq == "" {
		changefreq = "daily"
	}

	// Required
	htdocs = cfg.CheckOption("htdocs", cfg.MergeEnv("htdocs", htdocs), true)
	siteURL = cfg.CheckOption("site_url", cfg.MergeEnv("site_url", siteURL), true)
	sitemap = cfg.CheckOption("sitemap", cfg.MergeEnv("sitemap", sitemap), true)

	excludeList := strings.Split(excluded, ":")
	// NOTE: Exclude the error pages in htdocs
	excludeList = append(excludeList, "htdocs/40")
	excludeList = append(excludeList, "htdocs/50")

	log.Printf("Starting map of %s\n", htdocs)
	filepath.Walk(htdocs, func(p string, info os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".html") {
			if IsExcluded(excludeList, p) == false {
				finfo := new(locInfo)
				finfo.Loc = fmt.Sprintf("%s%s", siteURL, strings.TrimPrefix(p, htdocs))
				yr, mn, dy := info.ModTime().Date()
				finfo.LastMod = fmt.Sprintf("%d-%0.2d-%0.2d", yr, mn, dy)
				log.Printf("Adding %s\n", finfo.Loc)
				locList = append(locList, finfo)
			}
		}
		return nil
	})
	fmt.Printf("Writing %s\n", sitemap)
	fp, err := os.OpenFile(sitemap, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		log.Fatalf("Can't create %s, %s\n", sitemap, err)
	}
	defer fp.Close()
	fp.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`))
	for _, item := range locList {
		fp.WriteString(fmt.Sprintf(`
    <url>
            <loc>%s</loc>
            <lastmod>%s</lastmod>
            <changefreq>%s</changefreq>
    </url>
`, item.Loc, item.LastMod, changefreq))
	}
	fp.Write([]byte(`
</urlset>
`))
}
