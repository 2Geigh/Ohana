package main

import (
	"log"
	"sync"

	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-init/pkg/models"
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
		models.Url("https://baccyflap.com/noai/").TrimTrailingSlash(),
		models.Url("https://baccyflap.com/rsp/").TrimTrailingSlash(),
		models.Url("https://www.oocities.org/").TrimTrailingSlash(),
		models.Url("https://allchans.org/").TrimTrailingSlash(),
		models.Url("https://xiixiixii.xyz/ring").TrimTrailingSlash(),
		models.Url("https://猫.移动/webring/").TrimTrailingSlash(),
		models.Url("https://dynamicland.org/").TrimTrailingSlash(),
		models.Url("https://webring.htmlhobbyist.com/").TrimTrailingSlash(),
		models.Url("https://francophonering.neocities.org/").TrimTrailingSlash(),
		models.Url("https://euroring.neocities.org/").TrimTrailingSlash(),
		models.Url("https://evehibi.nekoweb.org/ringlink/info").TrimTrailingSlash(),
		models.Url("https://wiby.me/").TrimTrailingSlash(),
		models.Url("https://32bit.cafe/"),
		models.Url("https://wiki.melonland.net/web_revival").TrimTrailingSlash(),
		models.Url("https://webring.htmlhobbyist.com/list/").TrimTrailingSlash(),
		models.Url("https://taylor.town/about").TrimTrailingSlash(),
		models.Url("https://rip.so/").TrimTrailingSlash(),
		models.Url("https://wiki.archiveteam.org").TrimTrailingSlash(),

		models.Url("https://delightful.coding.social/").TrimTrailingSlash(),
		models.Url("https://delightful.club").TrimTrailingSlash(),

		models.Url("https://www.unix.dog/users").TrimTrailingSlash(),
		models.Url("https://bukmark.club/").TrimTrailingSlash(),
		models.Url("https://250kb.club/").TrimTrailingSlash(),
		models.Url("https://webring.dinhe.net/").TrimTrailingSlash(),
		models.Url("https://ring.acab.dev/").TrimTrailingSlash(),
		models.Url("https://leftwingbooks.net/").TrimTrailingSlash(),
	}
)

func init() {
	err := database.InitializeDB()
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}
}

func main() {
	var (
		wg  sync.WaitGroup
		err error
	)

	defer database.DB.Close()

	wg.Go(func() {
		err = database.InitializeDomainBlacklist(database.DB)
		if err != nil {
			log.Fatalf("initialize domain blacklist failed: %v", err)
		}
		wg.Done()
	})

	wg.Go(func() {
		for _, seed_url := range seed_urls {
			err := database.EnqueueLinks([]models.Url{seed_url}, database.DB, &database.DatabaseMu)
			if err != nil {
				log.Printf("enqueue seed URLs failed: %v", err)
			}
		}
		wg.Done()
	})

	wg.Wait()
}
