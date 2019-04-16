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
		date := s.First().Find("td div.normal").Last().Text()
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
		role := s.First().Find("td.alt2").First().Find("div.smallfont").First().Text()
		fmt.Println(role)
		cmds := s.Find("div.voz-post-message")
		jo := s.Find("td.alt2 table tbody tr td").Last().Find("div.smallfont").Find("div").First().Text()
		fmt.Println(strings.Trim(jo, ("Join Date: ")))
		fmt.Println(strings.TrimSpace(cmds.Text()))
		cmds.Find("img ").Each(func(i int, alink *goquery.Selection) {
			url, ok := alink.Attr("src")
			if ok {
				fmt.Println(url)
			}
		})
		fmt.Println("---------------")

	})

}
