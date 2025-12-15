package poki_handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/utils/file_bucket"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

type (
	PokiGame struct {
		Name           string `json:"name"`
		Url            string `json:"url"`
		Image          string `json:"image"`
		ImageThumbnail string `json:"image_thumbnail"`
		Section        string `json:"section"`
		GameLink       string `json:"game_link"`
	}
)

const (
	POKI_HOST = "https://poki.com"
)

func GetGameList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pokiGames, err := scrapPokiHome(ctx)
	if err != nil {
		render.Response(w, r, http.StatusInternalServerError, map[string]any{
			"message": "failed to scrap poki home",
		})
		return
	}

	render.Response(w, r, 200, map[string]any{
		"games": pokiGames,
	})
}

func GetGameDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameDetail, err := scrapPokiGameDetail(ctx, r.URL.Query().Get("url"))
	if err != nil {
		render.Response(w, r, http.StatusInternalServerError, map[string]any{
			"message": "failed to scrap poki game detail",
		})
		return
	}

	render.Response(w, r, 200, gameDetail)
}

func scrapPokiHome(ctx context.Context) ([]PokiGame, error) {
	pokiGames := []PokiGame{}

	c := colly.NewCollector()

	// c.OnHTML("#app-root > div.lStd1276e_IhuA3g3FIs.s9w4UjUUDL2klmhRDNdo > div:nth-child(2) > div:nth-child(1) > ul > li", func(e *colly.HTMLElement) {
	// 	image := e.ChildAttr("a > div > figure > img", "src")
	// 	if image == "" {
	// 		image = e.ChildAttr("a > picture > img", "src")
	// 	}
	// 	if image == "" {
	// 		image = e.ChildAttr("a > img", "src")
	// 	}

	// 	pokiGames = append(pokiGames, PokiGame{
	// 		Name:    e.ChildText("a > span"),
	// 		Url:     e.ChildAttr("a", "href"),
	// 		Image:   image,
	// 		Section: "big",
	// 	})
	// })

	c.OnHTML("#app-root > div.lStd1276e_IhuA3g3FIs.s9w4UjUUDL2klmhRDNdo > div:nth-child(2) > div:nth-child(2) > a", func(e *colly.HTMLElement) {
		image := e.ChildAttr("div > figure > img", "src")
		if image == "" {
			image = e.ChildAttr("picture > img", "src")
		}
		if image == "" {
			image = e.ChildAttr("img", "src")
		}

		pokiGames = append(pokiGames, PokiGame{
			Name:           e.ChildText("span"),
			Url:            e.Attr("href"),
			Image:          image,
			ImageThumbnail: file_bucket.GenCacheUrl(image, "", 200, 200),
			Section:        "small",
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	err := c.Visit(POKI_HOST + "/id")
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return pokiGames, err
	}

	return pokiGames, nil
}

func scrapPokiGameDetail(ctx context.Context, gameUrl string) (PokiGame, error) {
	visitTarget := POKI_HOST + gameUrl
	logrus.Infof("visit poki game detail: %s", visitTarget)

	page := browser.MustPage(visitTarget).MustWaitStable()

	// page.MustScreenshot("pokigame.png")

	iframeElement := page.MustElement("#game-element")

	iframePage := iframeElement.MustFrame()

	el := iframePage.MustElement("#gameframe")

	// htmlContent := iframePage.MustHTML()
	// fmt.Println("HTML CONTENT", htmlContent)

	return PokiGame{
		GameLink: *el.MustAttribute("src"),
	}, nil
}
