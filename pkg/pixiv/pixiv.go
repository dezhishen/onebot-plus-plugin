package pixiv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	illusts_url            = "https://app-api.pixiv.net/v1/walkthrough/illusts"
	pixiv_cat_generate_url = "https://api.pixiv.cat/v1/generate"
)

type imageUrls struct {
	SquareMedium string `json:"square_medium"`
	Medium       string `json:"medium"`
	Large        string `json:"large"`
}
type illust struct {
	Id        int64     `json:"id"`
	ImageUrls imageUrls `json:"image_urls"`
}

type illustsResp struct {
	Illusts []*illust `json:"illusts"`
}

func RandomImgsWithRetry() ([]string, error) {
	var r []string
	var e error
	for i := 0; i < 3; i++ {
		r, e = RandomImgs()
		if e == nil {
			return r, e
		}
		logrus.Warnf("获取图片列表发生错误,%v", e)
	}
	return r, e
}

func DownloadImage(url string) ([]byte, error) {
	imageResp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	imageBytes, err := ioutil.ReadAll(imageResp.Body)
	if err != nil {
		return nil, err
	}
	imageResp.Body.Close()
	return imageBytes, nil
}

func getProxy() string {
	// return "socks5://127.0.0.1:10080"
	return os.Getenv("PIXIV_PROXY")
}

func RandomImgs() ([]string, error) {
	cli := newHttpCli()
	r, err := cli.Get(illusts_url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var resp illustsResp
	err = json.Unmarshal(robots, &resp)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, v := range resp.Illusts {
		s, e := generatePixivCat(v.Id)
		if e != nil {
			logrus.Warnf("生成代理图片失败,%v", e)
			continue
		}
		result = append(result, s)
		if len(result) > 10 {
			break
		}
	}
	return result, nil
}

type artist struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
type generateResp struct {
	Artist           artist `json:"artist"`
	OriginalUrl      string `json:"original_url"`
	OriginalUrlProxy string `json:"original_url_proxy"`
	Title            string `json:"title"`
	Success          bool   `json:"success"`
	Error            string `json:"error"`
}

func generatePixivCat(id int64) (string, error) {
	cli := newHttpCli()
	postData := url.Values{}
	postData.Set("p", fmt.Sprintf("%v", id))
	r, err := cli.Post(
		pixiv_cat_generate_url,
		"application/x-www-form-urlencoded; charset=UTF-8",
		strings.NewReader(postData.Encode()),
	)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	var resp generateResp
	err = json.Unmarshal(robots, &resp)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", errors.New(resp.Error)
	}
	return strings.ReplaceAll(resp.OriginalUrlProxy, "pixiv.cat", "pixiv.re"), nil
}

func newHttpCli() *http.Client {
	var cli *http.Client
	proxy := getProxy()
	if proxy != "" {
		logrus.Infof("使用代理,%v", proxy)
		proxyUrl, err := url.Parse(proxy)
		if err == nil {
			cli = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		}
	}
	if cli == nil {
		cli = http.DefaultClient
	}
	return cli
}
