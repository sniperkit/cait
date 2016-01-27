
# overview

_aspace_ is a go library wrapping the [ArchivesSpace](http://archivesspace.org) REST API.
It include support for content export, static site generation, indexing and independent
search engine service.  This means you can manage your content in ArchivesSpace but
server and search the public content independent of the status of ArchivesSpace itself.
This gives you more options for deployment as well as providing a clean seperation of
concerns for public/admin uses.

All tools can be configured through environment variables. Some have additional
command line options that can be invoked.  Generally launching the tool with a
"-h" or "--help" will get you a list of basic features and options.

## tools

### aspace

_aspace_ command line utility is the workhorse for getting content out of ArchivesSpace
and onto your local file system in a useful static form (JSON blobs).  _aspace_ will
eventually support putting content back into ArchivesSpace. At that stage you'll have
more options for batch editing content with more general tools like R, Open Refine, etc.

### aspacepage

_aspacepage_ renders the content dumped by _aspace_ into a website structure suitable
for hosting with _aspacesearch_ search engine and webserver.  It does NOT talk
directly to ArchivesSpace so can does not increase the load on your ArchivesSpace server.

### aspaceindexer

_aspaceindexer_ is a utility to creating or updating a Bleve index used by _aspacesearch_
web server.  It crawls the website tree an ingests JSON files found in the
accessions directories. It can be run manually but is more suited to run periodically
via a cronjob (say once every day as needed).   For my collection of about 10,000
JSON blobs it run to completion takes 45 minutes or so. A little faster creating a new
index structure then updating an existing one.  The current implementation is overly
simplistic and certainly can be improved (e.g. rather than indexing files
individually it could batch and index)

### aspacesearch

_aspacesearch_ is a webserver and search engine. It is intended to run behind a more
traditional webserver like NginX or Apache.  Output of the search results are controlled
by the Golang HTML templates.  This is an early implementation so this will see change
as the project gets deployed into a production setting.

_aspacesearch_ can be started manually but more typically would be brought up by
your init process (e.g. /etc/init.d/aspacesearch start). An example init file
is provided