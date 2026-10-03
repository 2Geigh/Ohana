package main

import (
	"log"
	"sync"

	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-init/pkg/models"
)

var (
	seed_urls = []models.Url{
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
		err = database.
			InitializeDomainBlacklist(database.DB)
		if err != nil {
			log.Fatalf("initialize domain blacklist failed: %v", err)
		}
		wg.Done()
	})

	wg.Go(func() {
		for _, seed_url := range seed_urls {
			err := database.EnqueueLinks(
				[]models.Url{seed_url},
				database.DB,
				&database.DatabaseMu,
			)
			if err != nil {
				log.Printf("enqueue seed URLs failed: %v", err)
			}
		}
		wg.Done()
	})

	wg.Wait()
}
