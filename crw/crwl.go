package crw

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type (
	// To fan-in model
	// Cary data over goroutine
	PagesChanel struct {
		PageNumber int
		PageData   *goquery.Selection
	}
	PostsChanel struct {
		PostCount int
		PostDate  string
		PostData  *goquery.Selection
	}
	UrlMetadata struct {
		Url     string
		pageNum int
	}
	//Cary comment and save it
	Comment struct {
		User      User   `json:"user"`
		PostCount int    `json:"post_count"`
		PostDate  string `json:"post_date"`
		Cmd       string `json:"cmd"`
	}
	//User information who post this comment
	User struct {
		UserName    string `json:"user_name"\`
		UserPage    string `json:"user_page"`
		Description string `json:"description"`
		JoinDate    string `json:"join_date"`
	}
	CommenttoWrite struct {
		cmd []Comment `json:"cmd"`
	}
)

var (
	mu   = sync.Mutex{}
	cmtw = CommenttoWrite{}
)

func Crawler(ctx context.Context, cf VozConfig) {
	lastNuPage, err := getLastPageNu(cf.TheadUrl)
	if err != nil {
		logrus.Errorf("can't get page number all crawler is finish %v", err)
	}
	pageschanel := make(chan PagesChanel, lastNuPage)
	UrlChanel := make(chan UrlMetadata, lastNuPage)
	go func(ctx context.Context) {
		var pgwg sync.WaitGroup
		for i := 1; i <= cf.NumWorker; i++ {
			pgwg.Add(1)
			go PageCrawler(i, ctx, pgwg, pageschanel, UrlChanel)
		}
		pgwg.Wait()
		logrus.Infof("all page crawler is done extracting data")
		close(pageschanel)
	}(ctx)
	postChanel := make(chan PostsChanel, 10*lastNuPage)
	// restriction worker to avoid goroutine grown without control
	go func(ctx context.Context) {
		var postwg sync.WaitGroup
		for j := 1; j <= 10; j++ {
			postwg.Add(1)
			go PostCrawler(ctx, j, postwg, pageschanel, postChanel)
		}
		postwg.Wait()
		close(postChanel)
	}(ctx)

	cmdchanel := make(chan Comment)
	urlchanel := make(chan string)
	go func(ctx context.Context) {
		var postwg sync.WaitGroup
		for j := 1; j <= cf.NumWorker; j++ {
			postwg.Add(1)
			go DataExtraction(ctx, postwg, j, postChanel, cmdchanel, urlchanel)
		}
		postwg.Wait()
		close(postChanel)
	}(ctx)
	go func(ctx context.Context) {
		var cmdwg sync.WaitGroup
		for j := 1; j <= cf.NumWorker; j++ {
			cmdwg.Add(1)
			go Save(j, cmdwg, urlchanel, cmdchanel, cf)
		}
		cmdwg.Wait()
	}(ctx)

	for pageidx := 1; pageidx <= lastNuPage; pageidx++ {
		logrus.Infof("throwing %s&page=%d into chanel to process ", cf.TheadUrl, pageidx)
		UrlChanel <- UrlMetadata{
			Url:     fmt.Sprintf("%s&page=%d", cf.TheadUrl, pageidx),
			pageNum: pageidx,
		}
	}
	close(UrlChanel)
}
func PostCrawler(ctx context.Context, idx int, pg sync.WaitGroup, pagesin <-chan PagesChanel, postOut chan<- PostsChanel) {
	defer pg.Done()

	logrus.Infof("post crawler #%d : crawling post data ", idx)
	for {
		select {
		case <-ctx.Done():
			logrus.Info("all crawler is terminated by user ")
			return
		case p, ok := <-pagesin:
			if !ok {
				logrus.Infof("post crawler done")
				return
			}
			logrus.Infof("post crawler #%d : getting post from page %d", idx, p.PageNumber)
			page := p.PageData
			page.Each(func(i int, post *goquery.Selection) {
				postcountstr, _ := post.Find("a").First().Attr("name")
				postcountint, _ := strconv.Atoi(postcountstr)
				date := strings.TrimSpace(post.First().Find("td div.normal").Last().Text())
				data := post
				out := PostsChanel{
					PostCount: postcountint,
					PostDate:  date,
					PostData:  data,
				}
				postOut <- out
			})
			logrus.Infof("post crawler #%d : getting post from page %d done", idx, p.PageNumber)
		}
	}
}
func PageCrawler(idx int, ctx context.Context, wg sync.WaitGroup, pageout chan<- PagesChanel, urlin <-chan UrlMetadata) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("all crawler is terminated by user ")
			return
		case url, ok := <-urlin:
			if !ok {
				logrus.Info("page crawler all done")
				return
			}
			logrus.Infof("page crawler #%d extracting page:%d", idx, url.pageNum)
			pageout <- DataCrawler(url.Url, url.pageNum)
			logrus.Infof("page :%s done", url.Url)

		}
	}
}
func DataCrawler(url string, numpage int) PagesChanel {
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
	pageDoc := Doc.Find("table.tborder.voz-postbit")
	page := PagesChanel{PageNumber: numpage, PageData: pageDoc}
	return page
}

//func CheckResult(cmd chan Comment) {
//	for {
//		select {
//		case c, ok := <-cmd:
//			if !ok {
//				logrus.Infof("done")
//			}
//			logrus.Info(c.PostCount)
//			logrus.Infof("-------")
//		}
//	}
//}
func DataExtraction(ctx context.Context, wg sync.WaitGroup, idx int, posts <-chan PostsChanel, cmtout chan<- Comment, imageURl chan<- string) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("data extraction is terminated")
			return
		case post, ok := <-posts:
			if !ok {
				logrus.Infof("all crawler is done")
				return
			}
			logrus.Infof("data extraction :crawl %d cmt at #%d", idx, post.PostCount)
			postdata := post.PostData
			usrif := postdata.First().Find("td.alt2").First().Find("a.bigusername")
			userUrl, ok := usrif.Attr("href")
			userName := usrif.Text()
			role := postdata.First().Find("td.alt2").First().Find("div.smallfont").First().Text()
			joindDate := postdata.Find("td.alt2 table tbody tr td").Last().Find("div.smallfont").Find("div").First().Text()
			user := User{
				UserName:    userName,
				UserPage:    fmt.Sprintf("https://forums.voz.vn/%s", userUrl),
				Description: role,
				JoinDate:    strings.Trim(joindDate, ("Join Date: ")),
			}
			cmdstr := postdata.Find("div.voz-post-message").Text()
			comnent := Comment{
				User:      user,
				PostCount: post.PostCount,
				PostDate:  post.PostDate,
				Cmd:       cmdstr,
			}
			cmtout <- comnent
			postdata.Find("div.voz-post-message").Find("img").Each(func(i int, imglink *goquery.Selection) {
				url, ok := imglink.Attr("src")
				if ok {
					if strings.HasPrefix(url, "http") {
						imageURl <- url
					}
				}
			})
			logrus.Infof(" data extraction crawl cmt at #%d is done", post.PostCount)
		}

	}
}

func Save(idx int, wg sync.WaitGroup, imgchanel chan string, cmdchan chan Comment, cf VozConfig) {
	defer wg.Done()
	dir, _ := MakeDirFormTitle(cf.TheadUrl)
	logrus.Infof("save %d runging ", idx)
	client := http.Client{}
	for {
		select {
		case url, ok := <-imgchanel:
			if !ok {
				logrus.Infof("save done")
			}
			logrus.Infof("image form url : %s", url)
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			if resp.StatusCode/200 != 1 {

				continue
			}
			b, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			fp := filepath.Join(dir, url[strings.LastIndex(url, "/")+1:])
			err = ioutil.WriteFile(fp, b, 0644)
			if err != nil {
				continue
			}
		case cmd, ok := <-cmdchan:
			if !ok {
				logrus.Infof("save done")
			}
			mu.Lock()
			cmtw := append(cmtw.cmd, cmd)
			mu.Unlock()
			j, _ := json.Marshal(cmtw)
			fp := filepath.Join(dir, fmt.Sprintf("%scmd.json", dir))
			err := ioutil.WriteFile(fp, j, 0644)
			if err != nil {
				continue
			}
		}
	}
}
