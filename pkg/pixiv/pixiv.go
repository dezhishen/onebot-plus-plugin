package pixiv

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	imageSourceList = "https://api.lolicon.app/setu/v2?r18=1&num=20&size=regular"
	// pixiv_cat_generate_url = "https://api.pixiv.cat/v1/generate"
)

type PixivUrl struct {
	Original string `json:"original"`
	Regular  string `json:"regular"`
	Small    string `json:"small"`
	Thumb    string `json:"thumb"`
	Mini     string `json:"mini"`
}

func (p *PixivUrl) GetUrl() string {
	if p.Original != "" {
		return p.Original
	}
	if p.Regular != "" {
		return p.Regular
	}
	if p.Small != "" {
		return p.Small
	}
	if p.Thumb != "" {
		return p.Thumb
	}
	return p.Mini
}

type PixivImage struct {
	Pid    int64     `json:"pid"`
	UId    int64     `json:"uid"`
	Title  string    `json:"title"`
	Author string    `json:"author"`
	R18    bool      `json:"r18"`
	Tags   []string  `json:"tags"`
	Urls   *PixivUrl `json:"urls"`
}
type imageResp struct {
	Error string        `json:"error"`
	Data  []*PixivImage `json:"data"`
}

func RandomImgsWithRetry() ([]*PixivImage, error) {
	var r []*PixivImage
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

func RandomImgs() ([]*PixivImage, error) {
	cli := newHttpCli()
	r, err := cli.Get(imageSourceList)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	robots, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var resp imageResp
	err = json.Unmarshal(robots, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Data, nil
}

var cli *http.Client

func newHttpCli() *http.Client {
	if cli == nil {
		cli = http.DefaultClient
	}
	return cli
}
