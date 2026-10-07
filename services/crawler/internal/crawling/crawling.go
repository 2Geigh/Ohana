package crawling

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
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
	CRAWLER_POLITENESS_INTERVAL time.Duration = 12 * time.Second

	// How long the crawler keeps ignoring a re-encountered page after crawling it
	CRAWLER_OLDNESS_THRESHOLD time.Duration = 24 * time.Hour //
)

var (
	SeedURLs = []models.Url{
		models.Url("https://download.mozilla.org/?product=firefox-stub&os=win&lang=en-US"),
		models.Url("http://nicholasgarcia.com"),
		models.Url("https://nicholasgarcia.com"),
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
	}
)

type Queues map[models.Domain][]models.Url

func (queues Queues) Dequeue(fqdn models.Domain) models.Url {

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

func Crawl(
	queues *(Queues),
	domain models.Domain,
	db *sql.DB,
	mu *sync.Mutex,
	numberOfCrawlers *atomic.Int64,
	iterator *atomic.Int64,
	id int64,
) {
	// log.Printf("Starting crawler %d for %s (%d links)...", id, domain, len((*queues)[domain]))

	var (
		page models.Webpage

		// debugging
		currentUrl models.Url = "void"
		checkpoint string
		err        error
	)

	defer func() {
		raisedError := recover()

		if raisedError != nil {
			log.Printf("[%s] PANICKED AFTER %q\npanic: %v\nstack trace:\n%s",
				currentUrl,
				checkpoint,
				raisedError,
				debug.Stack())
		}
	}()

	for func() int {
		mu.Lock()
		defer mu.Unlock()
		return len((*queues)[domain])
	}() > 0 {
		mu.Lock()
		currentUrl = (*queues).Dequeue(domain)
		mu.Unlock()

		page.
			Url = currentUrl
		checkpoint = "set page.Url"

		page.
			FullDomain = page.Url.GetDomain()
		checkpoint = "set page.FullDomain"

		page.
			Fqdn = page.FullDomain.GetFQDN()
		checkpoint = "set page.TopAndSecondLevelDomain"

		var (
			isPageTooRecentlyCrawled bool
		)
		isPageTooRecentlyCrawled, err = page.Url.
			IsTooRecentlyCrawled(
				connection.DB,
				CRAWLER_OLDNESS_THRESHOLD,
			)
		if err != nil {
			log.Printf("[%s] determine page freshness failed: %v", currentUrl, err)
			continue
		}

		if isPageTooRecentlyCrawled {
			continue
		}

		hasDomainBeenRequestedTooRecently, err := page.
			Fqdn.
			HasBeenRequestedTooRecently(
				CRAWLER_POLITENESS_INTERVAL,
				&database.DatabaseMu,
				connection.DB)
		if err != nil {
			log.Printf("[%s] determine necessary politeness failed: %v", currentUrl, err)
			continue
		}

		if hasDomainBeenRequestedTooRecently {
			time.Sleep(CRAWLER_POLITENESS_INTERVAL)
		}

		response, err := http.Get(string(currentUrl))
		if err != nil {
			log.Printf("[%s] GET request unfulfilled: %v", currentUrl, err)
			continue
		}

		var (
			isRequestSuccessful bool = response.StatusCode >= 200 && response.StatusCode < 300
		)
		if !isRequestSuccessful {
			log.Printf("[%s] %s", currentUrl, response.Status)
			continue
		}

		body := struct {
			asReadCloser io.ReadCloser
			asBytes      []byte
		}{
			asReadCloser: response.Body,
		}
		body.asBytes, err = io.ReadAll(response.Body)
		if err != nil {
			log.Printf("[%s] read response body failed: %v", currentUrl, err)
			continue
		}
		err = response.Body.Close()
		if err != nil {
			panic(fmt.Errorf("close resposne body failed: %w", err))
		}

		if !utf8.Valid(body.asBytes) {
			contentType := response.Header.Get("Content-Type")

			reader, err := charset.NewReader(
				bytes.NewReader(body.asBytes),
				contentType,
			)
			if err != nil {
				log.Printf("[%s] UTF-8 transcoding failed: create byte reader failed (likely couldn't determine character encoding): %v", currentUrl, err)
				continue
			}

			body.asBytes, err = io.ReadAll(reader)
			if err != nil {
				log.Printf("[%s] UTF-8 transcoding failed: conversion failed: %v", currentUrl, err)
				continue
			}

			if !utf8.Valid(body.asBytes) {
				log.Printf("[%s] UTF-8 transcoding failed: still non-UTF-8 after conversion", currentUrl)
				continue
			}
		}

		page.ResponseBody = string(body.asBytes)
		checkpoint = "set page.ResponseBody"

		// TODO: Skip XML pages

		doc, err := html.Parse(
			strings.NewReader(string(body.asBytes)),
		)
		if err != nil {
			log.Printf("[%s] parse HTML failed: %v", currentUrl, err)
			continue
		}

		// We do this instantiation step using `make`
		// So that if len(page.Outneighbours) == 0,
		// Postgres will read it as an empty array
		// Instead of as a NULL value
		page.Outneighbours = make([]models.Url, 0)
		page.Outneighbours = append(page.Outneighbours, findHyperlinks(doc, currentUrl)...)
		err = database.EnqueueLinks(page.Outneighbours, connection.DB, &database.DatabaseMu)
		if err != nil {
			log.Printf("[%s] enqueue failed: %v", currentUrl, err)
			continue
		}

		var (
			isDomainBlacklisted bool
		)
		isDomainBlacklisted, err = page.Fqdn.IsBlacklisted(connection.DB)
		if err != nil {
			log.Printf("[%s] determine domain blacklist status failed: %v", currentUrl, err)
			continue
		}
		if isDomainBlacklisted {
			continue
		}

		page.Title = parsePageTitle(doc)
		checkpoint = "set page.Title"

		page.Description = parsePageDescription(doc)
		checkpoint = "set page.Description"

		page.IsFediverseNode, err = isPageOnFediverse(page)
		if err != nil {
			log.Printf("[%s] determine fediverse participation status failed: %v", currentUrl, err)
		}

		err = page.Save(connection.DB)
		if err != nil {
			log.Printf("[%s] save to database failed: %v", currentUrl, err)
			continue
		}

		err = page.EnqueueToIndexer(connection.DB, &database.DatabaseMu)
		if err != nil {
			log.Printf("[%s] enqueue to indexer failed: %v", currentUrl, err)
			continue
		}

		mu.Lock()
		_ = iterator.Add(1)

		log.Printf("[%s] (Q=%d, n=%d, l=%d, i=%d) {id=%d}",
			currentUrl,
			len(*queues), // number of unique domains queued in memory
			numberOfCrawlers.Load(),
			len((*queues)[domain]),
			iterator.Load(),
			id,
		)
		mu.Unlock()

		// log.Println("crawler:       ", crawler_id)
		// log.Println("iter:          ", *iterator)
		// log.Println("url:           ", currentUrl)
		// log.Println("title:         ", page.Title)
		// log.Println("desc:          ", page.Description)
		// log.Println("body:          ", len(page.Text), "bytes long")
		// log.Println("outneighbours: ", len(page.Outneighbours))
		// log.Println("response_body: ", len(page.ResponseBody), "bytes long")
		// log.Println("queue: ", len(queue.Links), "links long")
	}

	// log.Printf("Killing crawler %d for %s...", id, domain)
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

func findHyperlinks(
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
			hyperlinks = append(hyperlinks, newfoundLink.TrimTrailingSlash())
			break
		}
	}

	return hyperlinks
}

func isPageOnFediverse(
	p models.Webpage,
) (
	bool,
	error,
) {
	// TODO: IMPLEMENT THIS FUNCTION USING DATA FROM:
	// https://nodes.fediverse.party/

	return false, nil
}

func parsePageDescription(
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

	return pageDescription
}

func parsePageTitle(
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

	return pageTitle
}
