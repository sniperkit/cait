//
// cmds/genpages/genpages.go - A command line utility that builds pages from the exported results of cait.go
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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	// Caltech Library packages
	"github.com/caltechlibrary/cait"
	"github.com/caltechlibrary/cli"
	"github.com/caltechlibrary/tmplfn"
)

var (
	usage = `USAGE: %s [OPTIONS]`

	description = `
SYNOPSIS

%s generates HTML, .include pages and normalized JSON based on the JSON output form
cait and templates associated with the command.

CONFIGURATION

%s can be configured through setting the following environment
variables-

    CAIT_DATASET    this is the directory that contains the output of the
                      'cait archivesspace export' command.

    CAIT_TEMPLATES  this is the directory that contains the templates
                      used used to generate the static content of the website.

    CAIT_HTDOCS     this is the directory where the HTML files are written.
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

	showHelp    bool
	showVersion bool
	showLicense bool

	htdocsDir   string
	datasetDir  string
	templateDir string
)

func loadTemplates(templateDir, aHTMLTmplName, aIncTmplName string) (*template.Template, *template.Template, error) {
	tmplFuncs := tmplfn.Join(tmplfn.TimeMap, tmplfn.PageMap, cait.TmplMap)
	aHTMLTmpl, err := tmplfn.Assemble(tmplFuncs, path.Join(templateDir, aHTMLTmplName), path.Join(templateDir, aIncTmplName))
	if err != nil {
		return nil, nil, fmt.Errorf("Can't parse template %s, %s, %s", aHTMLTmplName, aIncTmplName, err)
	}
	aIncTmpl, err := tmplfn.Assemble(tmplFuncs, path.Join(templateDir, aIncTmplName))
	if err != nil {
		return aHTMLTmpl, nil, fmt.Errorf("Can't parse template %s, %s", aIncTmplName, err)
	}
	return aHTMLTmpl, aIncTmpl, nil
}

func processAgentsPeople(templateDir string, aHTMLTmplName string, aIncTmplName string) error {
	log.Printf("Reading templates from %s\n", templateDir)
	aHTMLTmpl, aIncTmpl, err := loadTemplates(templateDir, aHTMLTmplName, aIncTmplName)
	if err != nil {
		log.Fatalf("template error %q, %q: %s", aHTMLTmplName, aIncTmplName, err)
	}

	return filepath.Walk(path.Join(datasetDir, "agents/people"), func(p string, f os.FileInfo, err error) error {
		// Process accession records
		if strings.Contains(p, "agents/people") == true && strings.HasSuffix(p, ".json") == true {
			src, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}
			person := new(cait.Agent)
			err = json.Unmarshal(src, &person)
			if err != nil {
				return err
			}

			// FIXME: which restrictions do we care about agent/people?--
			//        agent.Published, person.DisplayName.IsDisplayName, person.DisplayName.Authorized
			if person.Published == true && person.IsLinkedToPublishedRecord == true && person.DisplayName.IsDisplayName == true && person.DisplayName.Authorized == true {
				// Create a normalized view of the accession to make it easier to work with

				// If the accession is published and the accession is not suppressed then generate the webpage
				fname := path.Join(htdocsDir, fmt.Sprintf("%s.html", person.URI))
				dname := path.Dir(fname)
				err = os.MkdirAll(dname, 0775)
				if err != nil {
					return fmt.Errorf("Can't create %s, %s", dname, err)
				}

				// Process HTML file
				fp, err := os.Create(fname)
				if err != nil {
					return fmt.Errorf("Problem creating %s, %s", fname, err)
				}
				log.Printf("Writing %s", fname)
				err = aHTMLTmpl.Execute(fp, person)
				if err != nil {
					log.Fatalf("template execute error %s, %s", aHTMLTmplName, err)
					return err
				}
				fp.Close()

				// Process Include file (just the HTML content)
				fname = path.Join(htdocsDir, fmt.Sprintf("%s.include", person.URI))
				fp, err = os.Create(fname)
				if err != nil {
					return fmt.Errorf("Problem creating %s, %s", fname, err)
				}
				log.Printf("Writing %s", fname)
				err = aIncTmpl.Execute(fp, person)
				if err != nil {
					log.Fatalf("template execute error %s, %s", aIncTmplName, err)
					return err
				}
				fp.Close()

				// Process JSON file (an abridged version of the JSON output in data)
				fname = path.Join(htdocsDir, fmt.Sprintf("%s.json", person.URI))
				src, err := json.Marshal(person)
				if err != nil {
					return fmt.Errorf("Could not JSON encode %s, %s", fname, err)
				}
				log.Printf("Writing %s", fname)
				err = ioutil.WriteFile(fname, src, 0664)
				if err != nil {
					log.Fatalf("could not write JSON view %s, %s", fname, err)
					return err
				}
				fp.Close()
			}
		}
		return nil
	})
}

func processAccessions(templateDir string, aHTMLTmplName string, aIncTmplName string, agents []*cait.Agent, subjects map[string]*cait.Subject, digitalObjects map[string]*cait.DigitalObject) error {
	log.Printf("Reading templates from %s\n", templateDir)
	aHTMLTmpl, aIncTmpl, err := loadTemplates(templateDir, aHTMLTmplName, aIncTmplName)
	if err != nil {
		log.Fatalf("template error %q, %q: %s", aHTMLTmplName, aIncTmplName, err)
	}

	return filepath.Walk(path.Join(datasetDir, "repositories"), func(p string, f os.FileInfo, err error) error {
		// Process accession records
		if strings.Contains(p, "accessions") == true && strings.HasSuffix(p, ".json") == true {
			src, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}
			accession := new(cait.Accession)
			err = json.Unmarshal(src, &accession)
			if err != nil {
				return err
			}

			// FIXME: which restrictions do we care about--
			//        accession.Publish, accession.Suppressed, accession.AccessRestrictions,
			//        accession.RestrictionsApply, accession.UseRestrictions
			if accession.Publish == true && accession.Suppressed == false && accession.RestrictionsApply == false {
				// Create a normalized view of the accession to make it easier to work with
				view, err := accession.NormalizeView(agents, subjects, digitalObjects)
				if err != nil {
					return fmt.Errorf("Could not generate normalized view, %s", err)
				}

				// If the accession is published and the accession is not suppressed then generate the webpage
				fname := path.Join(htdocsDir, fmt.Sprintf("%s.html", accession.URI))
				dname := path.Dir(fname)
				err = os.MkdirAll(dname, 0775)
				if err != nil {
					return fmt.Errorf("Can't create %s, %s", dname, err)
				}

				// Process HTML file
				fp, err := os.Create(fname)
				if err != nil {
					return fmt.Errorf("Problem creating %s, %s", fname, err)
				}
				log.Printf("Writing %s", fname)
				err = aHTMLTmpl.Execute(fp, view)
				if err != nil {
					log.Fatalf("template execute error %s, %s", aHTMLTmplName, err)
					return err
				}
				fp.Close()

				// Process Include file (just the HTML content)
				fname = path.Join(htdocsDir, fmt.Sprintf("%s.include", accession.URI))
				fp, err = os.Create(fname)
				if err != nil {
					return fmt.Errorf("Problem creating %s, %s", fname, err)
				}
				log.Printf("Writing %s", fname)
				err = aIncTmpl.Execute(fp, view)
				if err != nil {
					log.Fatalf("template execute error %s, %s", aIncTmplName, err)
					return err
				}
				fp.Close()

				// Process JSON file (an abridged version of the JSON output in data)
				fname = path.Join(htdocsDir, fmt.Sprintf("%s.json", accession.URI))
				src, err := json.Marshal(view)
				if err != nil {
					return fmt.Errorf("Could not JSON encode %s, %s", fname, err)
				}
				log.Printf("Writing %s", fname)
				err = ioutil.WriteFile(fname, src, 0664)
				if err != nil {
					log.Fatalf("could not write JSON view %s, %s", fname, err)
					return err
				}
				fp.Close()
			}
		}
		return nil
	})
}

func init() {
	// We are going to log to standard out rather than standard err
	log.SetOutput(os.Stdout)

	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "v", false, "display version")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "l", false, "display license")
	flag.BoolVar(&showLicense, "license", false, "display license")

	flag.StringVar(&htdocsDir, "htdocs", "", "specify where to write the HTML files to")
	flag.StringVar(&datasetDir, "dataset", "", "specify where to read the JSON files from")
	flag.StringVar(&templateDir, "templates", "", "specify where to read the templates from")
}

func main() {
	appName := path.Base(os.Args[0])
	flag.Parse()
	//args := flag.Args()

	cfg := cli.New(appName, "CAIT", fmt.Sprintf(license, appName, cait.Version), cait.Version)
	cfg.UsageText = fmt.Sprintf(usage, appName)
	cfg.DescriptionText = fmt.Sprintf(description, appName, appName)
	//cfg.ExampleText = fmt.Sprintf(examples, appName)
	cfg.OptionsText = "OPTIONS\n"

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

	datasetDir = cfg.CheckOption("dataset", cfg.MergeEnv("dataset", datasetDir), true)
	templateDir = cfg.CheckOption("templates", cfg.MergeEnv("templates", templateDir), true)
	htdocsDir = cfg.CheckOption("htdocs", cfg.MergeEnv("htdocs", htdocsDir), true)

	if htdocsDir != "" {
		if _, err := os.Stat(htdocsDir); os.IsNotExist(err) {
			os.MkdirAll(htdocsDir, 0775)
		}
	}

	//
	// Setup directories relationships
	//
	digitalObjectDir := ""
	filepath.Walk(datasetDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() == true && strings.HasSuffix(p, "digital_objects") {
			digitalObjectDir = p
			return nil
		}
		return err
	})
	if digitalObjectDir == "" {
		log.Fatalf("Can't find the digital object directory in %s", datasetDir)
	}
	subjectDir := path.Join(datasetDir, "subjects")
	agentsDir := path.Join(datasetDir, "agents", "people")

	//
	// Setup Maps and generate the accessions pages
	//
	log.Printf("Reading Subjects from %s", subjectDir)
	subjectsMap, err := cait.MakeSubjectMap(subjectDir)
	if err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("Reading Digital Objects from %s", digitalObjectDir)
	digitalObjectsMap, err := cait.MakeDigitalObjectMap(digitalObjectDir)
	if err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("Reading Agents/People from %s", agentsDir)
	agentsList, err := cait.MakeAgentList(agentsDir)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("Processing Agents/People in %s\n", agentsDir)
	err = processAgentsPeople(templateDir, "agents-people.html", "agents-people.include")

	log.Printf("Processing accessions in %s\n", datasetDir)
	err = processAccessions(templateDir, "accession.html", "accession.include", agentsList, subjectsMap, digitalObjectsMap)
	if err != nil {
		log.Fatalf("%s", err)
	}
}