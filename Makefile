#
# Simple Makefile for conviently testing, building and deploying experiment.
#
PROJECT = cait

VERSION = $(shell grep -m 1 'Version =' $(PROJECT).go | cut -d\` -f 2)

BRANCH = $(shell git branch | grep '* ' | cut -d\  -f 2)

PROGRAM_LIST = bin/cait bin/cait-genpages bin/cait-indexpages bin/cait-servepages 

API = cait.go io.go export.go schema.go search.go views.go

CMDS = cmds/*/*.go

build: $(API) $(PROGRAM_LIST) $(CMDS)


api: $(API)
	go build

cait: bin/cait

cait-genpages: bin/cait-genpages

cait-indexpages: bin/cait-indexpages

cait-servepages: bin/cait-servepages

bin/cait: $(API) cmds/cait/cait.go
	go build -o bin/cait cmds/cait/cait.go

bin/cait-genpages: $(API)  cmds/cait-genpages/cait-genpages.go
	go build -o bin/cait-genpages cmds/cait-genpages/cait-genpages.go

bin/cait-indexpages: $(API) cmds/cait-indexpages/cait-indexpages.go
	go build -o bin/cait-indexpages cmds/cait-indexpages/cait-indexpages.go

bin/cait-servepages: $(API) cmds/cait-servepages/cait-servepages.go
	go build -o bin/cait-servepages cmds/cait-servepages/cait-servepages.go

test:
	go test

clean:
	if [ -d bin ]; then /bin/rm -fR bin; fi
	if [ -d dist ]; then /bin/rm -fR dist; fi
	if [ -f $(PROJECT)-$(VERSION)-release.zip ]; then /bin/rm $(PROJECT)-$(VERSION)-release.zip; fi

install:
	env GOBIN=$(GOPATH)/bin go install cmds/cait/cait.go
	env GOBIN=$(GOPATH)/bin go install cmds/cait-genpages/cait-genpages.go
	env GOBIN=$(GOPATH)/bin go install cmds/cait-indexpages/cait-indexpages.go
	env GOBIN=$(GOPATH)/bin go install cmds/cait-servepages/cait-servepages.go

website:
	./mk-website.bash

save:
	if [ "$(msg)" != "" ]; then git commit -am "$(msg)"; else git commit -am "Quick save"; fi
	git push origin $(BRANCH)

refresh:
	git fetch origin
	git pull origin $(BRANCH)

status:
	git status

publish:
	./mk-website.bash
	./publish.bash

dist/linux-amd64: *.go cmds/cait/cait.go cmds/cait-genpages/cait-genpages.go cmds/cait-indexpages/cait-indexpages.go cmds/cait-servepages/cait-servepages.go
	env GOOS=linux GOARCH=amd64 go build -o dist/linux-amd64/cait cmds/cait/cait.go
	env GOOS=linux GOARCH=amd64 go build -o dist/linux-amd64/cait-genpages cmds/cait-genpages/cait-genpages.go
	env GOOS=linux GOARCH=amd64 go build -o dist/linux-amd64/cait-indexpages cmds/cait-indexpages/cait-indexpages.go
	env GOOS=linux GOARCH=amd64 go build -o dist/linux-amd64/cait-servepages cmds/cait-servepages/cait-servepages.go

dist/windows-amd64: *.go cmds/cait/cait.go cmds/cait-genpages/cait-genpages.go cmds/cait-indexpages/cait-indexpages.go cmds/cait-servepages/cait-servepages.go
	env GOOS=windows GOARCH=amd64 go build -o dist/windows-amd64/cait.exe cmds/cait/cait.go
	env GOOS=windows GOARCH=amd64 go build -o dist/windows-amd64/cait-genpages.exe cmds/cait-genpages/cait-genpages.go
	env GOOS=windows GOARCH=amd64 go build -o dist/windows-amd64/cait-indexpages.exe cmds/cait-indexpages/cait-indexpages.go
	env GOOS=windows GOARCH=amd64 go build -o dist/windows-amd64/cait-servepages.exe cmds/cait-servepages/cait-servepages.go

dist/macosx-amd64: *.go cmds/cait/cait.go cmds/cait-genpages/cait-genpages.go cmds/cait-indexpages/cait-indexpages.go cmds/cait-servepages/cait-servepages.go
	env GOOS=darwin GOARCH=amd64 go build -o dist/macosx-amd64/cait cmds/cait/cait.go
	env GOOS=darwin GOARCH=amd64 go build -o dist/macosx-amd64/cait-genpages cmds/cait-genpages/cait-genpages.go
	env GOOS=darwin GOARCH=amd64 go build -o dist/macosx-amd64/cait-indexpages cmds/cait-indexpages/cait-indexpages.go
	env GOOS=darwin GOARCH=amd64 go build -o dist/macosx-amd64/cait-servepages cmds/cait-servepages/cait-servepages.go

dist/raspbian-arm7: *.go cmds/cait/cait.go cmds/cait-genpages/cait-genpages.go cmds/cait-indexpages/cait-indexpages.go cmds/cait-servepages/cait-servepages.go
	env GOOS=linux GOARCH=arm GOARM=7 go build -o dist/raspbian-arm7/cait cmds/cait/cait.go
	env GOOS=linux GOARCH=arm GOARM=7 go build -o dist/raspbian-arm7/cait-genpages cmds/cait-genpages/cait-genpages.go
	env GOOS=linux GOARCH=arm GOARM=7 go build -o dist/raspbian-arm7/cait-indexpages cmds/cait-indexpages/cait-indexpages.go
	env GOOS=linux GOARCH=arm GOARM=7 go build -o dist/raspbian-arm7/cait-servepages cmds/cait-servepages/cait-servepages.go


release: dist/linux-amd64 dist/windows-amd64 dist/macosx-amd64 dist/raspbian-arm7
	mkdir -p dist
	mkdir -p dist/etc/systemd/system
	mkdir -p dist/scripts
	cp -v README.md dist/
	cp -v LICENSE dist/
	cp -v INSTALL.md dist/
	cp -v NOTES.md dist/
	cp -vR templates dist/
	cp -vR scripts/harvest-*.bash dist/scripts/
	cp -vR etc/*-example dist/etc/
	cp -vR etc/systemd/system/*-example dist/etc/systemd/system/
	./package-versions.bash > dist/package-versions.txt
	zip -r $(PROJECT)-$(VERSION)-release.zip dist/*


