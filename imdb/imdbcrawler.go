package imdb

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type (
	Film struct {
		ID          int                 `json:"id" bson:"_id"`
		Rank        int                 `json:"rank" bson:"rank"`
		URL         string              `json:"url" bson:"url"`
		Title       string              `json:"title" bson:"title"`
		Rate        string              `json:"rate" bson:"rate"`
		ReleaseDate int                 `json:"release_date" bson:"release_date"`
		Description string              `json:"description" bson:"description"`
		Credit      map[string][]string `json:"credit" bson:"credit"`
	}
)

func Crawler() {
	films := MakeURLTopRateNonChan()
	ll := ExtractDetailNonChan(films)

	jstring, err := json.MarshalIndent(ll, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", jstring)

}
func MakeURLTopRate(cfg IMDBConf, filmURLChan chan Film) {
	topURL := "https://www.imdb.com/chart/top?ref_=nv_mv_250"
	doc, err := GetDocFormURL(topURL)
	if err != nil {
		return
	}
	doc.Find("table.chart full-width").Find("tbody.lister-list").Find("tr").Each(func(rank int, films *goquery.Selection) {
		f := Film{}
		f.Rank = rank + 1
		filmURL, ok := films.Find("td.titleColumn").Find("a").Attr("href")
		if ok {
			f.URL = NormalizeURL(filmURL)
		}
		f.Title = films.Find("td.titleColumn").Find("a").Text()
		rawRelease := films.Find("td.titleColumn").Find("span.secondaryInfo").Text()
		rawRelease = strings.TrimLeft(rawRelease, "(")
		rawRelease = strings.TrimRight(rawRelease, ")")
		releaseDate, err := strconv.Atoi(rawRelease)
		if err != nil {
			f.ReleaseDate = 0
		} else {
			f.ReleaseDate = releaseDate
		}
		f.Rate = films.Find("td.imdbRating").Find("strong").Text()
		filmURLChan <- f
	})

}
func MakeURLTopRateNonChan() []Film {
	topURL := "https://www.imdb.com/chart/top?ref_=nv_mv_250"
	fmt.Println("make url")
	doc, err := GetDocFormURL(topURL)
	if err != nil {
		fmt.Println(err)
	}
	filmURLChan := []Film{}
	doc.Find("table.chart.full-width").Find("tbody.lister-list").Find("tr").Each(func(rank int, films *goquery.Selection) {
		f := Film{}
		f.Rank = rank + 1
		filmURL, ok := films.Find("td.titleColumn").Find("a").Attr("href")
		if ok {
			f.URL = NormalizeURL(filmURL)
		}
		f.Title = films.Find("td.titleColumn").Find("a").Text()
		rawRelease := films.Find("td.titleColumn").Find("span.secondaryInfo").Text()
		rawRelease = strings.TrimLeft(rawRelease, "(")
		rawRelease = strings.TrimRight(rawRelease, ")")
		releaseDate, err := strconv.Atoi(rawRelease)
		if err != nil {
			fmt.Println(err)
			f.ReleaseDate = 0
		} else {
			f.ReleaseDate = releaseDate
		}
		f.Rate = films.Find("td.imdbRating").Find("strong").Text()
		filmURLChan = append(filmURLChan, f)
	})
	return filmURLChan

}
func ExtractDetailNonChan(filmIn []Film) []Film {
	listReturn := []Film{}
	fmt.Println("extract")
	for _, film := range filmIn {
		filmDoc, _ := GetDocFormURL(film.URL)
		//if err != nil {
		//	fmt.Println(err)
		//}
		film.Description = strings.TrimSpace(filmDoc.Find("div.plot_summary").Find("div.summary_text").Text())
		m := make(map[string][]string)
		filmDoc.Find("div.plot_summary").Find("div.credit_summary_item").Each(func(i int, item *goquery.Selection) {
			key := strings.TrimRight(item.Find("h4.inline").Text(), ":")
			value := []string{}
			item.Find("a").Each(func(j int, values *goquery.Selection) {
				if !strings.HasPrefix(values.Text(), "See full cast") && !strings.HasPrefix(values.Text(), "1 more credit ") && !strings.HasPrefix(values.Text(), "1 more credit") {
					value = append(value, values.Text())
				}
			})
			m[key] = value
			fmt.Println()
		})
		film.Credit = m
		listReturn = append(listReturn, film)

	}
	return listReturn
}
func ExtractDetail(wg sync.WaitGroup, fimIn chan Film, filmOut chan Film) {
	defer wg.Done()
	for {
		select {
		case film, ok := <-fimIn:
			if !ok {
				return
			}
			newFilm := film
			filmDoc, err := GetDocFormURL(newFilm.URL)
			if err != nil {
				return
			}
			newFilm.Description = strings.TrimSpace(filmDoc.Find("div.plot_summary").Find("div.summary_text").Text())
			m := make(map[string][]string)
			filmDoc.Find("div.plot_summary").Find("div.credit_summary_item").Each(func(i int, item *goquery.Selection) {
				key := strings.TrimRight(item.Find("h4.inline").Text(), ":")
				value := []string{}
				item.Find("a").Each(func(i int, values *goquery.Selection) {
					value = append(value, values.Text())
				})
				m[key] = value

			})
			newFilm.Credit = m
			filmOut <- newFilm
		}
	}
}
func NormalizeURL(url string) (fullUrl string) {
	rootURL := "https://www.imdb.com"
	fullUrl = rootURL + url
	return
}
func GetDocFormURL(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36")
	Client := http.Client{}
	body, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(body.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil

}
