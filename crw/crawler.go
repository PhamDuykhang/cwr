package crw

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

func CrawlerPageStage1(url string) {
	rq, err := getCustomRequest(url)
	if err != nil {
		fmt.Print(err)
	}
	respond, err := getHTTPClient().Do(rq)
	if err != nil {
		fmt.Print(err)
	}
	Doc, err := goquery.NewDocumentFromReader(respond.Body)
	if err != nil {
		fmt.Print(err)
	}
	//Get postcount
	page := Doc.Find("table.tborder.voz-postbit")
	//for each commnet
	page.Each(func(i int, s *goquery.Selection) {
		postcount, ok := s.Find("a").First().Attr("name") //1
		//get date form page
		date := page.First().Find("td div.normal").Last().Text()
		if ok {
			fmt.Println(postcount) //1
		}
		fmt.Println(strings.TrimSpace(date)) //Yesterday, 11:37
		//get com form page
		usrif := s.First().Find("td.alt2").First().Find("a.bigusername")
		userurl, ok := usrif.Attr("href")
		if ok {
			fmt.Println(userurl) //member.php?u=1125
		}
		username := usrif.Text() //tamvatam

		fmt.Println(username)
		//get role Junior Member
		role := page.First().Find("td.alt2").First().Find("div.smallfont").First().Text()
		fmt.Println(role)
		fmt.Println(s.Find("div.voz-post-message").Text())
		fmt.Println("---------------")

	})

}
