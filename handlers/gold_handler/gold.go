package gold_handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

const (
	ANTAM_GOLD_PRICE_URL = "https://www.logammulia.com/id/harga-emas-hari-ini"
	ANTAM_BUYBACK_URL    = "https://www.logammulia.com/id/sell/gold"
	HFGOLD_PRICE_URL     = "https://muamalahemas.com"
)

type (
	GoldPriceRecap struct {
		Name               string      `json:"name"`
		BuyBackPrice       int64       `json:"buy_back_price"`
		BuyBackPriceChange int64       `json:"buy_back_price_change"`
		Prices             []GoldPrice `json:"prices"`
	}

	GoldPrice struct {
		Weight       float64 `json:"weight"`
		Price        int64   `json:"price"`
		PricePerGram int64   `json:"price_per_gram"`
		Raw          string  `json:"raw,omitempty"`
	}
)

const (
	CacheKeyAntamGoldPrice = "cache:gold:antam_price"
	CacheKeyHFGoldPrice    = "cache:gold:hfgold_price"
	CacheExpiration        = 6 * time.Hour
)

func GetTodayPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		antamGoldPrice GoldPriceRecap
		hfgoldPrice    GoldPriceRecap
		wg             sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		// antamGoldPrice, _ = datastore.WrapCache(ctx, CacheKeyAntamGoldPrice, CacheExpiration, scrapAntamGoldPrice)
	}()

	go func() {
		defer wg.Done()
		hfgoldPrice, _ = datastore.WrapCache(ctx, CacheKeyHFGoldPrice, CacheExpiration, scrapHFGoldPrice)
	}()

	wg.Wait()

	render.Response(w, r, 200, map[string]any{
		"antam":  antamGoldPrice,
		"hfgold": hfgoldPrice,
	})
}

func scrapAntamGoldPrice() (GoldPriceRecap, error) {
	page := browser.MustPage(ANTAM_GOLD_PRICE_URL).MustWaitStable()

	// page.MustScreenshot("antam_gold_price_page.png")

	goldPrices := []GoldPrice{}

	for i := 3; i <= 11; i++ {
		selector := fmt.Sprintf("body > section.section-padding.n-no-padding-top > div > div:nth-child(3) > div > div.grid-child.n-768-1per3.n-768-no-margin-bottom > table:nth-child(3) > tbody > tr:nth-child(%d)", i)

		elem := page.MustElement(selector)

		weightString, _ := elem.MustElement("td:nth-child(1)").Text()
		weight, _ := ExtractFloat(weightString)

		priceString, _ := elem.MustElement("td:nth-child(3)").Text()
		price, _ := ExtractInt(priceString)

		pricePerGram := float64(price) / weight

		goldPrices = append(goldPrices, GoldPrice{
			Weight:       weight,
			Price:        price,
			PricePerGram: int64(pricePerGram),
		})
	}

	pageBuyBack := browser.MustPage(ANTAM_BUYBACK_URL)
	pageBuyBack.MustScreenshot("antam_buyback_page.png")
	// selector := "body > section.section-padding.n-no-padding-bottom > div > div > div.grid-child.n-1200-2per3.n-no-margin-bottom > div > div > div.right > div > div:nth-child(1) > span.value > span.text"
	// elem := pageBuyBack.MustElement(selector)
	// buyBackPriceString, _ := elem.Text()
	// buyBackPrice, _ := ExtractInt(buyBackPriceString)

	// selector = "body > section.section-padding.n-no-padding-bottom > div > div > div.grid-child.n-1200-2per3.n-no-margin-bottom > div > div > div.right > div > div:nth-child(2) > span.value > span.text"
	// elem = pageBuyBack.MustElement(selector)
	// buyBackPriceChangeString, _ := elem.Text()
	// buyBackPriceChange, _ := ExtractInt(buyBackPriceChangeString)

	return GoldPriceRecap{
		Name:         "antam",
		BuyBackPrice: goldPrices[1].Price - 150_000,
		// BuyBackPriceChange: buyBackPriceChange,
		Prices: goldPrices,
	}, nil
}

func scrapHFGoldPrice() (GoldPriceRecap, error) {
	page := browser.MustPage(HFGOLD_PRICE_URL).MustWaitStable()

	goldPrices := []GoldPrice{}

	for i := 10; i >= 2; i-- {
		selector := fmt.Sprintf("body > div.elementor.elementor-84 > div.elementor-element.elementor-element-18f32a6f.e-flex.e-con-boxed.e-con.e-parent.e-lazyloaded > div > div.elementor-element.elementor-element-7d6f871e.e-con-full.e-flex.e-con.e-child > div.elementor-element.elementor-element-41eb1ee2.e-con-full.e-flex.e-con.e-child.animated.fadeInUp > div.elementor-element.elementor-element-422c239.elementor-widget.elementor-widget-shortcode > div > table > tbody > tr:nth-child(%d)", i)

		elem := page.MustElement(selector)

		weightString, _ := elem.MustElement("td:nth-child(1)").Text()
		weight, _ := ExtractFloat(weightString)

		priceString, _ := elem.MustElement("td:nth-child(2)").Text()
		price, _ := ExtractInt(priceString)

		pricePerGram := float64(price) / weight

		goldPrices = append(goldPrices, GoldPrice{
			Weight:       weight,
			Price:        price,
			PricePerGram: int64(pricePerGram),
		})
	}

	buyBackPriceString, _ := page.MustElement("body > div.elementor.elementor-84 > div.elementor-element.elementor-element-18f32a6f.e-flex.e-con-boxed.e-con.e-parent.e-lazyloaded > div > div.elementor-element.elementor-element-7d6f871e.e-con-full.e-flex.e-con.e-child > div.elementor-element.elementor-element-7c0dd45.e-con-full.e-flex.e-con.e-child.animated.fadeInUp > div.elementor-element.elementor-element-af6f5cb.elementor-widget.elementor-widget-shortcode > div > table > tbody > tr:nth-child(2) > td:nth-child(2)").Text()
	buyBackPrice, _ := ExtractInt(buyBackPriceString)

	return GoldPriceRecap{
		Name:         "hfgold",
		BuyBackPrice: buyBackPrice,
		Prices:       goldPrices,
	}, nil
}

var numberRegex = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?`)

func ExtractFloat(input string) (float64, error) {
	// Find the first match in the string
	match := numberRegex.FindString(input)

	if match == "" {
		return 0, fmt.Errorf("no number found in string")
	}

	return strconv.ParseFloat(match, 64)
}

func ExtractInt(input string) (int64, error) {
	input = strings.ReplaceAll(input, ",", "")
	input = strings.ReplaceAll(input, ".", "")

	// Find the first match in the string
	match := numberRegex.FindString(input)

	if match == "" {
		return 0, fmt.Errorf("no number found in string")
	}

	return strconv.ParseInt(match, 10, 64)
}
