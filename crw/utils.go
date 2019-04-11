package crw

import (
	"crypto/tls"
	"net/http"
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
