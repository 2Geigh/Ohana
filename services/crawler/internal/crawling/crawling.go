package crawling

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/2Geigh/Ohana/crawler/internal/connection"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-init/pkg/models"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/html/charset"
)

const (
	// How long the crawler waits between requests to the same domain
	CRAWLER_STARTING_POLITENESS_INTERVAL time.Duration = 12 * time.Second

	// How long the crawler keeps ignoring a re-encountered page after crawling it
	CRAWLER_OLDNESS_THRESHOLD time.Duration = 24 * time.Hour //
)

var (
	SeedURLs = []models.Url{
		models.Url("https://pornhub.com"), // test pornographic blacklisted domains
		models.Url("https://google.com"),  // test non-pornographic blacklsited domains
		models.Url("https://download.mozilla.org/?product=firefox-stub&os=win&lang=en-US"), // test non-"text/html" responses
		models.Url("http://nicholasgarcia.com"),                                            // test non-HTTPS links

		models.Url("https://nicholasgarcia.com"),
		models.Url("https://foreverliketh.is/"),
		models.Url("https://angeldolly.com/"),
		models.Url("https://nyscyra.net/"),
		models.Url("https://0xffff.one"),
		models.Url("https://v2ex.com"),
		models.Url("https://asclaria.org/"),
		models.Url("https://skateboard.nekoweb.org/webring/phightring"),
		models.Url("https://bookscorpion.neocities.org/webrings"),
		models.Url("https://ryqrtz.nekoweb.org/webrings.html"),
		models.Url("https://webri.ng/"),
		models.Url("https://ryqrtz.nekoweb.org/home.html"),
		models.Url("https://mikh.net/affiliates.php"),
		models.Url("https://msx.gay/"),
		models.Url("https://piclog.blue/index.php"),
		models.Url("https://silly.city/"),
		models.Url("https://gummyring.neocities.org/"),
		models.Url("https://list-me.com/links.php?cat=1"),
		models.Url("https://indieseek.xyz/"),
		models.Url("https://linklane.net/"),
		models.Url("https://smoothsailing.asclaria.org/"),
		models.Url("https://silly.city/"),
		models.Url("https://runegod.net/"),
		models.Url("https://theotaku.com/"),
		models.Url("https://rollcake.site/"),
		models.Url("https://www.royal-drama.net/theemperorsnewgroove/"),
		models.Url("https://indieweb.org/"),
		models.Url("https://gusbus.space/smallweb-subway/"),
		models.Url("https://nownownow.com/"),
		models.Url("https://sike.pona.la/"),
		models.Url("https://webring.bucketfish.me/"),
		models.Url("https://links.weirdnet.org/"),
		models.Url("https://web1.0hosting.net/"),
		models.Url("https://bearblog.dev/"),
		models.Url("https://webring.xxiivv.com/"),
		models.Url("https://lwn.net/"),
		models.Url("https://fediring.net/"),
		models.Url("https://weirdweboctober.website/"),
		models.Url("https://recordsofenemysurveillance.com/"),
		models.Url("https://comicfury.com/"),
		models.Url("https://baccyflap.com/noai/"),
		models.Url("https://baccyflap.com/rsp/"),
		models.Url("https://www.oocities.org/"),
		models.Url("https://allchans.org/"),
		models.Url("https://xiixiixii.xyz/ring"),
		models.Url("https://猫.移动/webring/"),
		models.Url("https://dynamicland.org/"),
		models.Url("https://webring.htmlhobbyist.com/"),
		models.Url("https://francophonering.neocities.org/"),
		models.Url("https://euroring.neocities.org/"),
		models.Url("https://evehibi.nekoweb.org/ringlink/info"),
		models.Url("https://wiby.me/"),
		models.Url("https://32bit.cafe"),
		models.Url("https://wiki.melonland.net/web_revival"),
		models.Url("https://webring.htmlhobbyist.com/list/"),
		models.Url("https://taylor.town/about"),
		models.Url("https://rip.so/"),
		models.Url("https://wiki.archiveteam.org"),

		models.Url("https://delightful.coding.social/"),
		models.Url("https://delightful.club"),

		models.Url("https://www.unix.dog/users"),
		models.Url("https://bukmark.club/"),
		models.Url("https://250kb.club/"),
		models.Url("https://webring.dinhe.net/"),
		models.Url("https://ring.acab.dev/"),
		models.Url("https://leftwingbooks.net/"),
		models.Url("https://ideas-arquitecturadas.blogspot.com"),
		models.Url("https://マリウス.com/"),
		models.Url("https://personalsit.es/"),
		models.Url("https://faircamp.webr.ing/"),
		models.Url("https://www.2chan.net/"),
		models.Url("https://www.2chan.net/bbsmenu.html"),
		models.Url("https://www.2chan.net/index2.html"),
		models.Url("https://imageboards.net/"),
		models.Url("https://aboutideasnow.com/"),
		models.Url("https://ytmnd.com"),
		models.Url("https://ring.fediverse.radio/"),
		models.Url("https://pages.gay/"),
		models.Url("https://whyarentyou.gay/"),
	}
)

type Crawler struct {
	Id               int64
	Queues           *Queues
	Fqdn             models.Domain
	CrawlerMu        *sync.Mutex
	Db               *sql.DB
	DbMu             *sync.Mutex
	NumberOfCrawlers *atomic.Int64
	CrawlIteration   *atomic.Int64

	currentUrl models.Url // for debugging
	checkpoint string     // for debugging

	currentPage        models.Webpage
	degreeOfPoliteness int
	politenessInterval time.Duration
	recievedResponse   recievedResponse
}

func (c Crawler) Crawl() {
	var (
		err error
	)

	c.currentUrl = "void"
	c.degreeOfPoliteness = 0
	c.politenessInterval = CRAWLER_STARTING_POLITENESS_INTERVAL

	defer func() {
		raisedError := recover()

		if raisedError != nil {
			log.Printf(
				"[%s] PANICKED AFTER %q\npanic: %v\nstack trace:\n%s",
				c.currentUrl,
				c.checkpoint,
				raisedError,
				debug.Stack(),
			)
		}
	}()

	for c.queueLength() > 0 {
		c.currentPage = models.Webpage{}
		c.recievedResponse = recievedResponse{}

		c.CrawlerMu.Lock()
		c.currentUrl = (*c.Queues).Dequeue(c.Fqdn)
		c.CrawlerMu.Unlock()

		// e ** 3 is ~20
		// 20 * 12 seconds is ~4 minutes
		// If we're still getting status code 429 after 4 minutes it's not worth trying the domain
		if c.degreeOfPoliteness > 3 {
			continue
		}

		c.currentPage.Url = c.currentUrl
		c.checkpoint = "set page.Url"

		c.currentPage.FullDomain = c.currentPage.Url.GetDomain()
		c.checkpoint = "set page.FullDomain"

		c.currentPage.Fqdn = c.currentPage.FullDomain.GetFQDN()
		c.checkpoint = "set page.Fqdn"

		c.currentPage.IsDomainBlacklisted, err = c.currentPage.Fqdn.IsBlacklisted(connection.DB, &database.DatabaseMu)
		if err != nil {
			c.logError("determine domain blacklist status failed", err)
			continue
		}
		if c.currentPage.IsDomainBlacklisted {
			c.logError("domain blacklisted", err)
			continue
		}

		c.currentPage.IsTooRecentlyCrawled, err = c.currentPage.
			Url.IsTooRecentlyCrawled(
			connection.DB,
			CRAWLER_OLDNESS_THRESHOLD,
		)
		if err != nil {
			c.logError("determine page freshness failed", err)
			continue
		}

		if c.currentPage.IsTooRecentlyCrawled {
			continue
		}

		c.currentPage.HasDomainBeenRequestedTooRecently, err = c.currentPage.
			Fqdn.HasBeenRequestedTooRecently(
			c.politenessInterval,
			&database.DatabaseMu,
			connection.DB)
		if err != nil {
			c.logError("determine necessary politeness failed", err)
			continue
		}

		if c.currentPage.HasDomainBeenRequestedTooRecently {
			c.waitPolitely()
			continue
		}

		response, err := http.Get(string(c.currentUrl))
		if err != nil {
			c.logError("GET request unfulfilled", err)
			continue
		}

		c.recievedResponse = recievedResponse{
			body: struct {
				asReadCloser io.ReadCloser
				asBytes      []byte
			}{
				asReadCloser: response.Body,
			},
			contentType:          response.Header.Get("Content-Type"),
			statusCode:           response.StatusCode,
			statusMessage:        response.Status,
			wasRequestSuccessful: response.StatusCode >= 200 && response.StatusCode < 300,
		}

		if c.recievedResponse.statusCode == 429 {
			c.waitPolitely()
			continue
		}

		if !c.recievedResponse.wasRequestSuccessful {
			c.logError(c.recievedResponse.statusMessage, nil)
			continue
		}

		c.recievedResponse.body.asBytes, err = io.ReadAll(response.Body)
		if err != nil {
			c.logError("read response body failed", err)
			continue
		}
		err = response.Body.Close()
		if err != nil {
			panic(fmt.Errorf("close response body failed: %w", err))
		}

		if !utf8.Valid(c.recievedResponse.body.asBytes) {
			contentType := response.Header.Get("Content-Type")

			reader, err := charset.NewReader(
				bytes.NewReader(c.recievedResponse.body.asBytes),
				contentType,
			)
			if err != nil {
				c.logError("UTF-8 transcoding failed: create byte reader failed (likely couldn't determine character encoding)", err)
				continue
			}

			c.recievedResponse.body.asBytes, err = io.ReadAll(reader)
			if err != nil {
				c.logError("UTF-8 transcoding failed: conversion failed", err)
				continue
			}

			if !utf8.Valid(c.recievedResponse.body.asBytes) {
				c.logError("UTF-8 transcoding failed: still non-UTF-8 after conversion", err)
				continue
			}
		}

		if !c.recievedResponse.isReadableText() {
			c.logError("page content isn't text-based", nil)
			continue
		}

		c.currentPage.ResponseBody = string(c.recievedResponse.body.asBytes)
		c.checkpoint = "set page.ResponseBody"

		doc, err := html.Parse(
			strings.NewReader(string(c.recievedResponse.body.asBytes)),
		)
		if err != nil {
			c.logError("parse HTML failed", err)
			continue
		}

		// We do this instantiation step using `make`
		// So that if len(page.Outneighbours) == 0,
		// Postgres will read it as an empty array
		// Instead of as a NULL value
		c.currentPage.Outneighbours = make([]models.Url, 0)
		c.currentPage.Outneighbours = append(c.currentPage.Outneighbours, c.findHyperlinks(doc, c.currentUrl)...)
		err = database.EnqueueLinks(c.currentPage.Outneighbours, connection.DB, &database.DatabaseMu)
		if err != nil {
			c.logError("enqueue failed", err)
			continue
		}

		c.currentPage.Title = c.parsePageTitle(doc)
		c.checkpoint = "set page.Title"

		c.currentPage.Description = c.parsePageDescription(doc)
		c.checkpoint = "set page.Description"

		c.currentPage.IsFediverseNode, err = c.isPageOnFediverse()
		if err != nil {
			c.logError("determine fediverse participation status failed", err)
			continue
		}

		err = c.currentPage.Save(connection.DB)
		if err != nil {
			c.logError("save to database failed", err)
			continue
		}

		err = c.currentPage.EnqueueToIndexer(connection.DB, &database.DatabaseMu)
		if err != nil {
			c.logError("enqueue to indexer failed", err)
			continue
		}

		c.CrawlerMu.Lock()
		_ = c.CrawlIteration.Add(1)

		c.logCrawlCompletion()
		c.CrawlerMu.Unlock()
	}
}

func (c Crawler) findHyperlinks(
	root_node *html.Node,
	root_url models.Url,
) []models.Url {
	var (
		hyperlinks []models.Url
	)

	for node := range root_node.Descendants() {
		isAnchor :=
			node.Type == html.ElementNode &&
				node.DataAtom == atom.A

		if !isAnchor {
			continue
		}

		for _, attribute := range node.Attr {
			if attribute.Key != "href" {
				continue
			}

			var (
				trimmedRootUrl models.Url = root_url
				anchorHref     string     = attribute.Val
				newfoundLink   models.Url
			)

			if len(anchorHref) < 2 {
				continue
			}

			if string(root_url[len(root_url)-1]) == "/" {
				trimmedRootUrl = root_url[0 : len(root_url)-1]
			}

			if string(anchorHref[0]) == "/" { // ex: <a href="/about">
				newfoundLink = models.Url(string(trimmedRootUrl) + anchorHref)
			} else if len(anchorHref) >= len("http") &&
				string(anchorHref[0:len("http")]) != "http" { // ex: <a href="intro.html">
				newfoundLink = models.Url(fmt.Sprintf("%s/%s", trimmedRootUrl, anchorHref))
			} else {
				newfoundLink = models.Url(anchorHref)
			}

			if len(newfoundLink) < len("http://") {
				continue
			}

			var (
				isHttpUriScheme        bool = len(anchorHref) >= len("http") && (anchorHref[0:len("http")] == "http")
				isHttpsUriScheme       bool = len(anchorHref) >= len("https") && (anchorHref[0:len("https")] == "https")
				isAlternativeUriScheme bool = !isHttpUriScheme && !isHttpsUriScheme
			)
			if isAlternativeUriScheme {
				continue
			}

			// REMOVE ? QUERIES FROM URLs

			// REMOVE mailto: AND ANY OTHER SUCH TYPES OF URLS

			// log.Println()
			// log.Println("root_url", root_url)
			// log.Println("trimmedRootUrl", trimmedRootUrl)
			// log.Println("anchorHref", anchorHref)
			// log.Println("newfoundLink", newfoundLink)
			// log.Println("isAlternativeUriScheme", isAlternativeUriScheme)
			// log.Printf("[%s] Found: %v", root_url, newfoundLink)

			isBlacklisted, err := newfoundLink.GetDomain().IsBlacklisted(c.Db, c.DbMu)
			if err != nil {
				log.Printf("determine blacklist status of newfound link %s failed: %v", newfoundLink, err)
				continue
			}

			if isBlacklisted {
				continue
			}

			hyperlinks = append(hyperlinks, newfoundLink.TrimTrailingSlash())
			break
		}
	}

	return hyperlinks
}

func (c Crawler) isPageOnFediverse() (
	bool,
	error,
) {
	// TODO: IMPLEMENT THIS FUNCTION USING DATA FROM:
	// https://nodes.fediverse.party/

	return false, nil
}

func (c Crawler) logCrawlCompletion() {
	log.Println(c.runtimeStats())
}

func (c Crawler) logError(
	message string,
	err error,
) {
	if err == nil {
		log.Printf("%s %s", c.runtimeStats(), message)
		return
	}

	log.Printf("%s %s: %v", c.runtimeStats(), message, err)
}

func (c Crawler) parsePageDescription(
	root_node *html.Node,
) string {
	var (
		pageDescription string
	)

	for node := range root_node.Descendants() {
		var (
			isMeta bool = node.DataAtom == atom.Meta
		)
		if !isMeta {
			continue
		}

		var (
			isMetaDescription bool = slices.Contains(
				node.Attr,
				html.Attribute{
					Key: "name",
					Val: "description"},
			)
		)
		if !isMetaDescription {
			continue
		}

		var (
			contentIndex = slices.IndexFunc(
				node.Attr,
				func(attr html.Attribute) bool {
					return attr.Key == "content"
				},
			)
			hasContentKey bool = contentIndex != -1
		)
		if !hasContentKey {
			continue
		}

		pageDescription = strings.TrimSpace(node.Attr[contentIndex].Val)
		break
	}

	c.checkpoint = "set page.Description"
	return pageDescription
}

func (c Crawler) parsePageTitle(
	root_node *html.Node,
) string {
	var (
		pageTitle string
	)

	for node := range root_node.Descendants() {
		var (
			isTitle bool = node.DataAtom == atom.Title
			isH1    bool = node.DataAtom == atom.H1
			isH2    bool = node.DataAtom == atom.H2
			isH3    bool = node.DataAtom == atom.H3
		)

		if isTitle {
			if node.FirstChild != nil {
				pageTitle = strings.TrimSpace(node.FirstChild.Data)
				break
			}

			pageTitle = strings.TrimSpace(node.Data)
			break
		}

		if isH1 { // If no <title> found, use <h1> as fallback
			if node.FirstChild != nil {
				pageTitle = strings.TrimSpace(node.FirstChild.Data)
				break
			}

			pageTitle = strings.TrimSpace(node.Data)
			break
		}

		if isH2 { // If no <h1> found, use <h2> as fallback
			if node.FirstChild != nil {
				pageTitle = strings.TrimSpace(node.FirstChild.Data)
				break
			}

			pageTitle = strings.TrimSpace(node.Data)
			break
		}

		if isH3 { // If no <h2> found, use <h3> as fallback
			if node.FirstChild != nil {
				pageTitle = strings.TrimSpace(node.FirstChild.Data)
				break
			}

			pageTitle = strings.TrimSpace(node.Data)
			break
		}
	}

	c.checkpoint = "set page.Title"
	return pageTitle
}

func (c Crawler) queueLength() int {
	c.CrawlerMu.Lock()
	defer c.CrawlerMu.Unlock()

	return len((*c.Queues)[c.Fqdn])
}

func (c Crawler) runtimeStats() string {
	return fmt.Sprintf("(Q=%d, n=%d, l=%d, i=%d) {id=%d}【%s】[%s]",
		len(*c.Queues), // number of unique domains queued in memory
		c.NumberOfCrawlers.Load(),
		len((*c.Queues)[c.Fqdn]),
		c.CrawlIteration.Load(),
		c.Id,
		c.currentPage.Fqdn,
		c.currentPage.Url,
	)
}

func (c Crawler) waitPolitely() {
	time.Sleep(c.politenessInterval)

	c.degreeOfPoliteness += 1
	c.politenessInterval = time.Duration(
		float64(CRAWLER_STARTING_POLITENESS_INTERVAL) * math.Exp(float64(c.degreeOfPoliteness)),
	)
}

type Queues map[models.Domain][]models.Url

func (queues Queues) Dequeue(
	fqdn models.Domain,
) models.Url {

	if len(queues[fqdn]) < 1 {
		return models.Url("")
	}

	toReturn := queues[fqdn][0]

	if len(queues[fqdn]) == 1 {
		queues[fqdn] = []models.Url{}
	} else {
		queues[fqdn] = queues[fqdn][1:]
	}

	return toReturn
}

type recievedResponse struct {
	body struct {
		asReadCloser io.ReadCloser
		asBytes      []byte
	}
	statusCode           int
	statusMessage        string
	wasRequestSuccessful bool
	contentType          string
}

func (r recievedResponse) isReadableText() bool {
	readableTextTypes := []string{
		"text/html",
		"text/text",
	}

	for _, contentType := range readableTextTypes {
		if strings.Contains(r.contentType, contentType) {
			return true
		}
	}

	return false
}

func CleanCrawlerQueue(
	db *sql.DB,
	mu *sync.Mutex,
) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if mu == nil {
		return fmt.Errorf("mutex is nil")
	}

	mu.Lock()
	result, err := db.Exec(
		`DELETE FROM crawler_queue cq
			USING pages p
			WHERE cq.hyperlink = p.link
				AND p.date_last_crawled > CURRENT_TIMESTAMP - $1::interval;`,
		fmt.Sprintf("%d seconds", int(CRAWLER_OLDNESS_THRESHOLD/time.Second)),
	)
	mu.Unlock()
	if err != nil {
		return fmt.Errorf("execute DELETE FROM crawler_queue failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}

	mu.Lock()
	log.Println()
	log.Printf("Crawler queue purged successfully, eliminating %d links.", rowsAffected)
	log.Println()
	mu.Unlock()
	return nil
}
