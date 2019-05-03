package imdb

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	giv := "/title/tt0111161/?pf_rd_m=A2FGELUUNOQJNL&pf_rd_p=e31d89dd-322d-4646-8962-327b42fe94b1&pf_rd_r=D70AGXBVMEB6Q7B64YY1&pf_rd_s=center-1&pf_rd_t=15506&pf_rd_i=top&ref_=chttp_tt_1"
	fmt.Println(NormalizeURL(giv))
}
func TestCrawler(t *testing.T) {
	start := time.Now()
	Crawler()
	duration := time.Since(start)
	fmt.Println(duration)
}
func TestExtractDetailNonChan(t *testing.T) {
	tr := "https://www.imdb.com/title/tt0111161/?pf_rd_m=A2FGELUUNOQJNL&pf_rd_p=e31d89dd-322d-4646-8962-327b42fe94b1&pf_rd_r=9BSACD527YX11PKNTZXR&pf_rd_s=center-1&pf_rd_t=15506&pf_rd_i=top&ref_=chttp_tt_1"
	f := Film{
		Title: "fdsds",
		URL:   tr,
	}
	ll := ExtractDetailNonChan([]Film{f})
	jstring, err := json.MarshalIndent(ll, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", jstring)
}
