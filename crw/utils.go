package crw

import (
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//All function reuse form voz crawler https://github.com/lnquy/vozer
func getHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 10 * time.Second,
	}
}

// Since 2018/09/01, voz added filter on go client user-agent.
// => Fake valid user-agent to bypass the filter.
func getCustomRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", "vflastvisit=1535954670; vflastactivity=0; vfforum_view=d99e85613f547374e9db4f942bf6192fb611ae2aa-1-%7Bi-17_i-1535954671_%7D; _ga=GA1.2.144936460.1535954673; _gid=GA1.2.1737523081.1535954673; _gat_gtag_UA_351630_1=1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36")
	return req, nil
}

//
func getLastPageNu(url string) (int, error) {
	req, err := getCustomRequest(url)
	if err != nil {
		return -1, fmt.Errorf("failed to init request to first page: %s", err)
	}
	resp, err := getHTTPClient().Do(req)
	if err != nil || resp.StatusCode/200 != 1 {
		return -1, fmt.Errorf("failed to crawl first page from thread: %s, %s", resp.Status, err)
	}
	firstPage, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		return -1, err
	}
	pageControlStr := firstPage.Find("div.neo_column.main table").First().Find("td.vbmenu_control").Text() // Page 1 of 100
	if pageControlStr == "" {                                                                              // Thread with only 1 page
		return 1, nil
	}
	lastPageStr := pageControlStr[strings.LastIndex(pageControlStr, " ")+1:]
	lastPageNu, err := strconv.Atoi(lastPageStr)
	if err != nil {
		return -1, err
	}
	return lastPageNu, nil
}
func MakeDirFormTitle(url string) (string, error) {
	req, err := getCustomRequest(url)
	if err != nil {
		return "", fmt.Errorf("failed to init request to first page: %s", err)
	}
	resp, err := getHTTPClient().Do(req)
	defer resp.Body.Close()
	if err != nil || resp.StatusCode/200 != 1 {
		return "", fmt.Errorf("failed to crawl first page from thread: %s, %s", resp.Status, err)
	}
	page, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	t := page.Find("td.navbar").First().Find("strong").Text()
	err = os.MkdirAll(fmt.Sprintf("%s", strings.TrimSpace(t)), os.ModePerm)
	if err != nil {
		fmt.Print(err)
	}

	if err != nil {
		return "", err
	}
	return strings.TrimSpace(t), nil
}
