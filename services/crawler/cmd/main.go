package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/2Geigh/Ohana/crawler/internal/database"
	"github.com/2Geigh/Ohana/crawler/internal/models"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	CRAWLER_POLITENESS_INTERVAL time.Duration = 12 * time.Second
	CRAWLER_OLDNESS_THRESHOLD   time.Duration = 86400 * time.Second // 7 days

	NUMBER_OF_CRAWLERS = 25
)

var (
	seed_urls = []models.Url{
		models.Url("https://nicholasgarcia.com").TrimTrailingSlash(),
		models.Url("https://angeldolly.com/").TrimTrailingSlash(),
		models.Url("https://nyscyra.net/").TrimTrailingSlash(),
		models.Url("https://0xffff.one").TrimTrailingSlash(),
		models.Url("https://v2ex.com").TrimTrailingSlash(),
		models.Url("https://asclaria.org/").TrimTrailingSlash(),
		models.Url("https://skateboard.nekoweb.org/webring/phightring").TrimTrailingSlash(),
		models.Url("https://bookscorpion.neocities.org/webrings").TrimTrailingSlash(),
		models.Url("https://ryqrtz.nekoweb.org/webrings.html").TrimTrailingSlash(),
		models.Url("https://webri.ng/").TrimTrailingSlash(),
		models.Url("https://ryqrtz.nekoweb.org/home.html").TrimTrailingSlash(),
		models.Url("https://mikh.net/affiliates.php").TrimTrailingSlash(),
		models.Url("https://msx.gay/").TrimTrailingSlash(),
		models.Url("https://piclog.blue/index.php").TrimTrailingSlash(),
		models.Url("https://silly.city/").TrimTrailingSlash(),
		models.Url("https://gummyring.neocities.org/").TrimTrailingSlash(),
		models.Url("https://list-me.com/links.php?cat=1").TrimTrailingSlash(),
		models.Url("https://indieseek.xyz/").TrimTrailingSlash(),
		models.Url("https://linklane.net/").TrimTrailingSlash(),
		models.Url("https://smoothsailing.asclaria.org/").TrimTrailingSlash(),
		models.Url("https://silly.city/").TrimTrailingSlash(),
		models.Url("https://runegod.net/").TrimTrailingSlash(),
		models.Url("https://theotaku.com/").TrimTrailingSlash(),
		models.Url("https://rollcake.site/").TrimTrailingSlash(),
		models.Url("https://www.royal-drama.net/theemperorsnewgroove/").TrimTrailingSlash(),
		models.Url("https://indieweb.org/").TrimTrailingSlash(),
		models.Url("https://gusbus.space/smallweb-subway/").TrimTrailingSlash(),
		models.Url("https://nownownow.com/").TrimTrailingSlash(),
		models.Url("https://sike.pona.la/").TrimTrailingSlash(),
		models.Url("https://webring.bucketfish.me/").TrimTrailingSlash(),
		models.Url("https://links.weirdnet.org/").TrimTrailingSlash(),
		models.Url("https://web1.0hosting.net/").TrimTrailingSlash(),
		models.Url("https://bearblog.dev/").TrimTrailingSlash(),
		models.Url("https://webring.xxiivv.com/").TrimTrailingSlash(),
		models.Url("https://lwn.net/").TrimTrailingSlash(),
		models.Url("https://fediring.net/").TrimTrailingSlash(),
		models.Url("https://weirdweboctober.website/").TrimTrailingSlash(),
		models.Url("https://recordsofenemysurveillance.com/").TrimTrailingSlash(),
		models.Url("https://comicfury.com/").TrimTrailingSlash(),
	}
)

func init() {
	err := database.InitializeDB()
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}

	err = database.InitializeDomainBlacklist(database.DB)
	if err != nil {
		log.Fatalf("initialize domain blacklist failed: %v", err)
	}

	for _, seed_url := range seed_urls {
		err := database.EnqueueLinks([]models.Url{seed_url}, database.DB, &database.DatabaseMu)
		if err != nil {
			log.Printf("enqueue seed URLs failed: %v", err)
		}
	}
}

func main() {
	var (
		wg sync.WaitGroup

		startCrawler func()
	)

	defer database.DB.Close()

	startCrawler = func() {
		wg.Add(1)

		go func() {
			crawl(&wg)

			// Crawler replaces itself when it returns
			startCrawler()
		}()
	}

	for range NUMBER_OF_CRAWLERS {
		startCrawler()
	}

	wg.Wait()
}

func crawl(wg *sync.WaitGroup) {
	var (
		page models.Webpage

		localQueue models.LocalQueue

		// debugging
		currentUrl models.Url = "void"
		checkpoint string
		err        error
	)

	defer wg.Done()

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

	localQueue.Links, err = database.DequeueLinks(database.DB, &database.DatabaseMu)
	if err != nil {
		log.Printf("dequeue from database's global queue failed: %v", err)
		return
	}

	for len(localQueue.Links) > 0 {
		currentUrl = localQueue.Dequeue()

		page.
			Url = currentUrl.TrimTrailingSlash()
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
		isPageTooRecentlyCrawled, err = page.Url.IsTooRecentlyCrawled(database.DB, CRAWLER_OLDNESS_THRESHOLD)
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
				database.DB)
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

		page.ResponseBody = string(body.asBytes)
		checkpoint = "set page.ResponseBody"

		if !utf8.Valid(body.asBytes) {
			log.Printf("[%s] invalid UTF-8", currentUrl)
			continue
		}

		doc, err := html.Parse(
			strings.NewReader(string(body.asBytes)),
		)
		if err != nil {
			log.Printf("[%s] parse HTML failed: %v", currentUrl, err)
			continue
		}

		page.Text, err = parsePageBody(doc)
		if err != nil {
			log.Printf("[%s] parse body content failed: %v", currentUrl, err)
			continue
		}

		// We do this instantiation step using `make`
		// So that if len(page.Outneighbours) == 0,
		// Postgres will read it as an empty array
		// Instead of as a NULL value
		page.Outneighbours = make([]models.Url, 0)
		page.Outneighbours = append(page.Outneighbours, findHyperlinks(doc, currentUrl)...)
		err = database.EnqueueLinks(page.Outneighbours, database.DB, &database.DatabaseMu)
		if err != nil {
			log.Printf("[%s] enqueue failed: %v", currentUrl, err)
			continue
		}

		var (
			isDomainBlacklisted bool
		)
		isDomainBlacklisted, err = page.Fqdn.IsBlacklisted(database.DB)
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

		err = page.Save(database.DB)
		if err != nil {
			log.Printf("[%s] save to database failed: %v", currentUrl, err)
			continue
		}

		// log.Println()
		// log.Println(page.Fqdn)
		log.Printf("[%s]", currentUrl)

		err = page.EnqueueToIndexer(database.DB, &database.DatabaseMu)
		if err != nil {
			log.Printf("[%s] enqueue to indexer failed: %v", currentUrl, err)
			continue
		}
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
}

func findHyperlinks(root_node *html.Node, root_url models.Url) []models.Url {
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

func parsePageBody(root_node *html.Node) (string, error) {
	var (
		body_node *html.Node

		sb  strings.Builder
		err error
	)

	// Find <body> node
	for node := range root_node.Descendants() {
		if node.DataAtom != atom.Body {
			continue
		}

		body_node = node
		break
	}

	if body_node == nil {
		return sb.String(), fmt.Errorf("no <body> node found")
	}

	for node := range body_node.Descendants() {
		var (
			isBodyText = node.DataAtom != atom.Script &&
				node.DataAtom != atom.Style &&

				node.DataAtom != atom.Math &&
				node.DataAtom != atom.Embed &&
				node.DataAtom != atom.Iframe &&
				node.DataAtom != atom.Object &&
				node.DataAtom != atom.Picture &&
				node.DataAtom != atom.Source &&

				node.DataAtom != atom.Area &&
				node.DataAtom != atom.Audio &&
				node.DataAtom != atom.B &&
				node.DataAtom != atom.Canvas &&
				node.DataAtom != atom.Command &&
				node.DataAtom != atom.I &&
				node.DataAtom != atom.Img &&
				node.DataAtom != atom.Map &&
				node.DataAtom != atom.Svg &&
				node.DataAtom != atom.Track &&
				node.DataAtom != atom.Video &&

				node.Type != html.CommentNode
		)

		if !isBodyText {
			continue
		}

		if node.FirstChild == nil {
			continue
		}

		if node.FirstChild.Type != html.TextNode {
			continue
		}

		// fmt.Println()
		// fmt.Println("               DATA", strings.TrimSpace(node.Data))
		// fmt.Println("           DATAATOM", node.DataAtom)
		// fmt.Println("           DATATYPE", node.Type)
		// fmt.Println("    FIRSTCHILD_DATA", strings.TrimSpace(node.FirstChild.Data))
		// fmt.Println("FIRSTCHILD_DATAATOM", node.FirstChild.DataAtom)
		// fmt.Println("FIRSTCHILD_DATATYPE", node.FirstChild.Type)

		_, err = sb.WriteString(fmt.Sprintf("%s ", strings.TrimSpace(node.FirstChild.Data)))
		if err != nil {
			err = fmt.Errorf("write to string builder failed: %w", err)
		}
	}

	return strings.TrimSpace(sb.String()), err
}

func parsePageDescription(root_node *html.Node) string {
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

func parsePageTitle(root_node *html.Node) string {
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
