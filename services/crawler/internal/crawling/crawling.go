package crawling

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/2Geigh/Ohana/crawler/internal/connection"
	"github.com/2Geigh/Ohana/crawler/internal/queue"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-init/pkg/models"
	"github.com/k3a/html2text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/html/charset"
)

const (
	// How long the crawler waits between requests to the same domain
	CRAWLER_STARTING_POLITENESS_INTERVAL time.Duration = 12 * time.Second

	// How long the crawler keeps ignoring a re-encountered page after crawling it
	CRAWLER_OLDNESS_THRESHOLD time.Duration = 24 * time.Hour //

	// One million webpages
	MAXIMUM_QUEUE_LENGTH = 1000000

	// Fifty thousand websites
	MAXIMUM_DATABASE_WEBSITE_CAPACITY = 50000
)

var (
	SeedURLs = []models.Url{
		// test pornographic blacklisted domains
		models.Url("https://pornhub.com"),

		// test non-pornographic blacklisted domains
		models.Url("https://google.com"),

		// test non-"text/html" responses
		models.Url("https://download.mozilla.org/?product=firefox-stub&os=win&lang=en-US"),

		// test non-HTTPS links
		models.Url("http://nicholasgarcia.com"),

		// Both these pages have different sets of links without linking to eachother
		models.Url("https://delightful.coding.social/"),
		models.Url("https://delightful.club"),

		models.Url("https://0xffff.one"),
		models.Url("https://250kb.club/"),
		models.Url("https://32bit.cafe"),
		models.Url("https://aboutideasnow.com/"),
		models.Url("https://allchans.org/"),
		models.Url("https://angeldolly.com/"),
		models.Url("https://asclaria.org/"),
		models.Url("https://baccyflap.com/noai/"),
		models.Url("https://baccyflap.com/rsp/"),
		models.Url("https://bearblog.dev/"),
		models.Url("https://bookscorpion.neocities.org/webrings"),
		models.Url("https://brisray.com/web/webring-list.htm"),
		models.Url("https://bukmark.club/"),
		models.Url("https://comicfury.com/"),
		models.Url("https://dynamicland.org/"),
		models.Url("https://euroring.neocities.org/"),
		models.Url("https://evehibi.nekoweb.org/ringlink/info"),
		models.Url("https://faircamp.webr.ing/"),
		models.Url("https://fediring.net/"),
		models.Url("https://foreverliketh.is/"),
		models.Url("https://francophonering.neocities.org/"),
		models.Url("https://gummyring.neocities.org/"),
		models.Url("https://gusbus.space/smallweb-subway/"),
		models.Url("https://ideas-arquitecturadas.blogspot.com"),
		models.Url("https://imageboards.net/"),
		models.Url("https://indieseek.xyz/"),
		models.Url("https://indieweb.org/"),
		models.Url("https://justforfunnoreally.dev/"),
		models.Url("https://leftwingbooks.net/"),
		models.Url("https://linklane.net/"),
		models.Url("https://links.weirdnet.org/"),
		models.Url("https://list-me.com/links.php?cat=1"),
		models.Url("https://lwn.net/"),
		models.Url("https://mikh.net/affiliates.php"),
		models.Url("https://msx.gay/"),
		models.Url("https://nekoweb.org"),
		models.Url("https://neocities.org"),
		models.Url("https://nicholasgarcia.com"),
		models.Url("https://nownownow.com/"),
		models.Url("https://nyscyra.net/"),
		models.Url("https://ourfavoritevoid.club"),
		models.Url("https://pages.gay/"),
		models.Url("https://personalsit.es/"),
		models.Url("https://piclog.blue/index.php"),
		models.Url("https://procrastinators.org"),
		models.Url("https://recordsofenemysurveillance.com/"),
		models.Url("https://nicholasgarcia.com"),
		models.Url("https://nightfall.city/"),
		models.Url("https://nownownow.com/"),
		models.Url("https://nyscyra.net/"),
		models.Url("https://pages.gay/"),
		models.Url("https://personalsit.es/"),
		models.Url("https://piclog.blue/index.php"),
		models.Url("https://recordsofenemysurveillance.com/"),
		models.Url("https://rentry.co/webring"),
		models.Url("https://ring.acab.dev/"),
		models.Url("https://ring.fediverse.radio/"),
		models.Url("https://rip.so/"),
		models.Url("https://rollcake.site/"),
		models.Url("https://runegod.net/"),
		models.Url("https://ryqrtz.nekoweb.org/home.html"),
		models.Url("https://ryqrtz.nekoweb.org/webrings.html"),
		models.Url("https://sike.pona.la/"),
		models.Url("https://silly.city/"),
		models.Url("https://skateboard.nekoweb.org/webring/phightring"),
		models.Url("https://smoothsailing.asclaria.org/"),
		models.Url("https://taylor.town/about"),
		models.Url("https://teacake.org/"),
		models.Url("https://theotaku.com/"),
		models.Url("https://v2ex.com"),
		models.Url("https://web1.0hosting.net/"),
		models.Url("https://silly.city/"),
		models.Url("https://skateboard.nekoweb.org/webring/phightring"),
		models.Url("https://smoothsailing.asclaria.org/"),
		models.Url("https://taylor.town/about"),
		models.Url("https://theotaku.com/"),
		models.Url("https://tuffgong.nekoweb.org/webring-list.html"),
		models.Url("https://v2ex.com"),
		models.Url("https://web1.0hosting.net/"),
		models.Url("https://web1.0hosting.net/mostvisited/"),
		models.Url("https://webri.ng/"),
		models.Url("https://webring.bucketfish.me/"),
		models.Url("https://webring.dinhe.net/"),
		models.Url("https://webring.htmlhobbyist.com/"),
		models.Url("https://webring.htmlhobbyist.com/list/"),
		models.Url("https://webring.xxiivv.com/"),
		models.Url("https://weirdweboctober.website/"),
		models.Url("https://whyarentyou.gay/"),
		models.Url("https://wiby.me/"),
		models.Url("https://wiki.archiveteam.org"),
		models.Url("https://wiki.melonland.net/web_revival"),
		models.Url("https://www.2chan.net/"),
		models.Url("https://www.2chan.net/bbsmenu.html"),
		models.Url("https://www.2chan.net/index2.html"),
		models.Url("https://www.bltc.com/"),
		models.Url("https://www.geocities.ws/jaup/jaup.htm"),
		models.Url("https://www.oocities.org/"),
		models.Url("https://www.royal-drama.net/theemperorsnewgroove/"),
		models.Url("https://www.unix.dog/users"),
		models.Url("https://xiixiixii.xyz/ring"),
		models.Url("https://ytmnd.com"),
		models.Url("https://マリウス.com/"),
		models.Url("https://猫.移动/webring/"),
		models.Url("https://delightful.club"),
		models.Url("https://delightful.coding.social/"),
	}
)

type Crawler struct {
	Id int

	CrawlIteration   *atomic.Int64
	Queue            *sql.DB
	CrawlerMu        *sync.Mutex
	GlobalDb         *sql.DB
	GlobalDbMu       *sync.Mutex
	NumberOfCrawlers int

	currentPage              models.Webpage
	hasLinksRemainingToCrawl bool
	degreeOfPoliteness       int
	politenessInterval       time.Duration
	recievedResponse         recievedResponse

	// for debugging
	checkpoint     string
	useCheckpoints bool
}

func (c *Crawler) Crawl() {
	// UNCOMMENT TO ENABLE DEBUGGING LOGGING
	// c.useCheckpoints = true

	c.degreeOfPoliteness = 0
	c.politenessInterval = CRAWLER_STARTING_POLITENESS_INTERVAL
	c.hasLinksRemainingToCrawl = true

	for {
		func() {

			defer func() {
				raisedError := recover()

				if raisedError != nil {
					log.Printf(
						"[%s] PANICKED AFTER %q\npanic: %v\nstack trace:\n%s",
						c.currentPage.Url,
						c.checkpoint,
						raisedError,
						debug.Stack(),
					)
				}
			}()

			newLink, err := queue.Dequeue(c.Queue, c.CrawlerMu)

			if strings.TrimSpace(string(newLink)) == "" || err != nil {
				if err != sql.ErrNoRows {
					c.logError("dequeue link failed", err)
				}

				secondsToWait := rand.Intn(5) + 1
				time.Sleep(
					time.Duration(secondsToWait) * time.Second,
				)

				return
			}

			err = c.processLink(newLink)
			if err != nil {
				c.logError("process link failed", err)
			}
		}()
	}
}

func (c *Crawler) setCheckpoint(message string) {
	c.checkpoint = message

	if c.useCheckpoints {
		log.Println(message)
	}
}

func (c *Crawler) processLink(link models.Url) error {
	var (
		err error
	)

	c.currentPage = models.Webpage{}
	c.setCheckpoint("set currentPage = models.Webpage{}")

	c.recievedResponse = recievedResponse{}
	c.setCheckpoint("set recievedResponse = recievedResponse{}")

	c.currentPage.Url = link
	c.setCheckpoint(fmt.Sprintf("set currentPage.Url = %s", link))

	// e ** 3 is ~20
	// 20 * 12 seconds is ~4 minutes
	// If we're still getting status code 429 after 4 minutes it's not worth trying the domain
	if c.degreeOfPoliteness > 5 {
		return fmt.Errorf("this domain seems to be permablocking our requests")
	}
	c.setCheckpoint("check existing degree of politeness")

	c.currentPage.IsDomainBlacklisted, err = c.currentPage.Url.GetDomain().GetFQDN().IsBlacklisted(
		connection.DB, &database.DatabaseMu,
	)
	if err != nil {
		return fmt.Errorf("determine domain blacklist status failed: %w", err)
	}
	if c.currentPage.IsDomainBlacklisted {
		return fmt.Errorf("domain blacklisted")
	}
	c.setCheckpoint("evaluated domain blacklist status")

	c.currentPage.HasDomainBeenRequestedTooRecently, err = c.currentPage.
		Url.GetDomain().GetFQDN().HasBeenRequestedTooRecently(
		c.politenessInterval,
		c.Queue,
		c.CrawlerMu,
	)
	if err != nil {
		return fmt.Errorf("determine necessary politeness failed: %w", err)
	}
	c.setCheckpoint("Determined domain crawling politeness status")
	c.setCheckpoint(fmt.Sprintf(
		"recent request? %t; interval=%s",
		c.currentPage.HasDomainBeenRequestedTooRecently,
		c.politenessInterval,
	))

	if c.currentPage.HasDomainBeenRequestedTooRecently {
		c.setCheckpoint(fmt.Sprintf(
			"starting politeness wait at %s",
			time.Now().Format(time.RFC3339)),
		)
		c.waitPolitely()
		c.setCheckpoint(fmt.Sprintf(
			"finished politeness wait at %s",
			time.Now().Format(time.RFC3339)),
		)
	}

	// this comes before the request to make timing more consistent between crawler processes
	// in the event that a request takes a really long time to retrieve
	err = queue.UpdateTimeLastRequested(c.currentPage.Url, c.Queue, c.CrawlerMu, c.useCheckpoints)
	if err != nil {
		return fmt.Errorf("update time last requested for %s failed: %w", c.currentPage.Url, err)
	}
	c.setCheckpoint("executed queue.UpdateTimeLastRequested")

	response, err := http.Get(string(c.currentPage.Url))
	if err != nil {
		return fmt.Errorf("GET request unfulfilled: %w", err)
	}
	c.setCheckpoint(fmt.Sprintf("recieved response from GET %s", c.currentPage.Url))

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
		c.increasePoliteness()
		c.waitPolitely()
	} else if !c.recievedResponse.wasRequestSuccessful {
		return fmt.Errorf("%s", c.recievedResponse.statusMessage)
	}

	c.recievedResponse.body.asBytes, err = io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
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
			return fmt.Errorf("%s: %s: %w",
				"UTF-8 transcoding failed",
				"create byte reader failed (likely couldn't determine character encoding)",
				err,
			)
		}

		c.recievedResponse.body.asBytes, err = io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("UTF-8 transcoding failed: conversion failed: %w", err)

		}

		if !utf8.Valid(c.recievedResponse.body.asBytes) {
			return fmt.Errorf("UTF-8 transcoding failed: still non-UTF-8 after conversion: %w", err)
		}
	}

	if !c.recievedResponse.isReadableText() {
		return fmt.Errorf("page content isn't text-based")
	}

	c.currentPage.ResponseBody = string(c.recievedResponse.body.asBytes)
	c.setCheckpoint("set page.ResponseBody")

	doc, err := html.Parse(
		strings.NewReader(string(c.recievedResponse.body.asBytes)),
	)
	if err != nil {
		return fmt.Errorf("parse HTML failed: %w", err)
	}

	c.currentPage.Title = c.parsePageTitle(doc)
	c.setCheckpoint("set page.Title")

	c.currentPage.Description = c.parsePageDescription(doc)
	c.setCheckpoint("set page.Description")

	c.currentPage.Text = c.parsePageText(string(c.recievedResponse.body.asBytes))
	c.setCheckpoint("set page.Text")

	// We do this instantiation step using `make`
	// So that if len(page.Outneighbours) == 0,
	// Postgres will read it as an empty array
	// Instead of as a NULL value
	c.currentPage.Outneighbours = make([]models.Url, 0)
	c.currentPage.Outneighbours = append(c.currentPage.Outneighbours,
		c.findHyperlinks(doc, c.currentPage.Url)...,
	)

	err = queue.Enqueue(c.Queue, c.CrawlerMu, c.currentPage.Outneighbours, MAXIMUM_QUEUE_LENGTH)
	if err != nil {
		return fmt.Errorf("enqueue failed: %w", err)
	}

	err = c.currentPage.Save(connection.DB, c.CrawlerMu, MAXIMUM_DATABASE_WEBSITE_CAPACITY)
	if err != nil {
		return fmt.Errorf("save to database failed: %w", err)
	}

	_ = c.CrawlIteration.Add(1)
	c.logCrawlCompletion()

	return nil
}

func (c *Crawler) findHyperlinks(root_node *html.Node, root_url models.Url) []models.Url {
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

			// TODO: REMOVE ? QUERIES FROM URLs

			// TODO: REMOVE mailto: AND ANY OTHER SUCH TYPES OF URLS

			// log.Println()
			// log.Println("root_url", root_url)
			// log.Println("trimmedRootUrl", trimmedRootUrl)
			// log.Println("anchorHref", anchorHref)
			// log.Println("newfoundLink", newfoundLink)
			// log.Println("isAlternativeUriScheme", isAlternativeUriScheme)
			// log.Printf("[%s] Found: %v", root_url, newfoundLink)

			isBlacklisted, err := newfoundLink.GetDomain().IsBlacklisted(c.GlobalDb, c.GlobalDbMu)
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

func (c *Crawler) isPageOnFediverse() (bool, error) {
	// TODO: IMPLEMENT THIS FUNCTION USING DATA FROM:
	// https://nodes.fediverse.party/

	return false, nil
}

func (c *Crawler) logCrawlCompletion() {
	log.Println(c.runtimeStats())
}

func (c *Crawler) logError(message string, err error) {
	if err == nil {
		log.Printf("%s %s", c.runtimeStats(), message)
		return
	}

	log.Printf("%s %s: %v", c.runtimeStats(), message, err)
}

func (c *Crawler) parsePageDescription(root_node *html.Node) string {
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

	c.setCheckpoint("set page.Description")
	return pageDescription
}

func (c *Crawler) parsePageText(html string) string {
	pageText := html2text.HTML2TextWithOptions(html, html2text.WithLinksInnerText())

	pageTextWithWhitespaceSquashed := strings.Join(strings.Fields(pageText), " ")

	return strings.TrimSpace(pageTextWithWhitespaceSquashed)
}

func (c *Crawler) parsePageTitle(root_node *html.Node) string {
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

	c.setCheckpoint("set page.Title")
	return pageTitle
}

func (c *Crawler) runtimeStats() string {
	var (
		queueLength                       int64
		numberOfLinksLeftForCurrentDomain int64
	)

	c.CrawlerMu.Lock()

	_ = c.Queue.QueryRow(
		`SELECT count(*)
		FROM crawler_queue
		LEFT JOIN domains ON domains.id == crawler_queue.domain_id
		WHERE domains.fqdn = ?;
		`, c.currentPage.Url.GetDomain().GetFQDN(),
	).Scan(&numberOfLinksLeftForCurrentDomain)

	_ = c.Queue.QueryRow(
		`SELECT count(*)
		FROM crawler_queue;`,
	).Scan(&queueLength)

	c.CrawlerMu.Unlock()

	return fmt.Sprintf("(n=%d, L=%d, l=%d, i=%d) {id=%d}【%s】[%s]",
		c.NumberOfCrawlers,
		queueLength,
		numberOfLinksLeftForCurrentDomain,
		c.CrawlIteration.Load(),
		c.Id,
		c.currentPage.Url.GetDomain().GetFQDN(),
		c.currentPage.Url,
	)
}

func (c *Crawler) waitPolitely() {
	c.setCheckpoint(fmt.Sprintf("Waiting politely (degree = %d)...", c.degreeOfPoliteness))

	time.Sleep(c.politenessInterval)

	c.politenessInterval = time.Duration(
		float64(CRAWLER_STARTING_POLITENESS_INTERVAL) * math.Exp(float64(c.degreeOfPoliteness)),
	)

	c.setCheckpoint(fmt.Sprintf("Waited politely (NEW degree = %d)...", c.degreeOfPoliteness))
}

func (c *Crawler) increasePoliteness() {
	c.degreeOfPoliteness += 1
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

	// log.Println("actual content-type:", r.contentType)
	return false
}

func CleanCrawlerQueue(queueDb *sql.DB, queueDbMu *sync.Mutex, globalDb *sql.DB) error {
	rows, err := globalDb.Query(`
		SELECT link
		FROM pages
		WHERE date_last_crawled > CURRENT_TIMESTAMP - $1::interval
		`,
		fmt.Sprintf("%d seconds", int(CRAWLER_OLDNESS_THRESHOLD/time.Second)),
	)
	if err != nil {
		return fmt.Errorf("get pages from global DB failed: %w", err)
	}
	defer rows.Close()

	queueDbMu.Lock()
	defer queueDbMu.Unlock()

	tx, err := queueDb.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`CREATE TEMP TABLE IF NOT EXISTS recently_crawled_pages (link TEXT PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("create temporary table in embedded queue database transaction failed: %w", err)
	}
	_, err = tx.Exec(`DELETE FROM recently_crawled_pages`)
	if err != nil {
		return fmt.Errorf("clear temporary table in embedded queue database transaction failed: %w", err)
	}

	insertStmt, err := tx.Prepare("INSERT OR IGNORE INTO recently_crawled_pages (link) VALUES (?)")
	if err != nil {
		return fmt.Errorf("prepare insert stmt failed: %w", err)
	}

	for rows.Next() {
		var link string

		err = rows.Scan(&link)
		if err != nil {
			return fmt.Errorf("failed to scan link into local variable: %w", err)
		}

		_, err = insertStmt.Exec(link)
		if err != nil {
			return fmt.Errorf("failed to execute insert stmt: %w", err)
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("rows error: %w", rows.Err())
	}

	result, err := tx.Exec(`
		DELETE FROM crawler_queue
		WHERE EXISTS (
			SELECT 1
			FROM recently_crawled_pages p
			WHERE p.link = crawler_queue.hyperlink
		)`,
	)
	if err != nil {
		return fmt.Errorf("execute DELETE FROM crawler_queue failed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}

	log.Println()
	log.Printf("Crawler queue purged successfully, eliminating %d links.", rowsAffected)
	log.Println()

	return nil
}
