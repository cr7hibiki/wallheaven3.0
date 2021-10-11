package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	apikey      = ""
	rowUrl      = "https://wallhaven.cc/api/v1/search"
	q           = flag.String("q", "", "关键词")
	categories  = flag.String("categories", "", "分类(一般/动漫/人物)")                                                     // (一般/动漫/人物)
	purity      = flag.String("purity", "", "纯度(sfw/sketchy/nsfw)")                                                 // (sfw/sketchy/nsfw)
	sorting     = flag.String("sorting", "random", "排序(date_added*, relevance, random, views, favorites, toplist)") // date_added*, relevance, random, views, favorites, toplist
	order       = flag.String("order", "", "升序降序")
	topRange    = flag.String("topRange", "", "排名范围(1d, 3d, 1w,1M*, 3M, 6M, 1y)") // 1d, 3d, 1w,1M*, 3M, 6M, 1y
	atleast     = flag.String("atleast", "", "最小尺寸(1920 * 1080)")                 // 1920 * 1080
	resolutions = flag.String("resolutions", "", "指定尺寸, 逗号隔开")
	ratios      = flag.String("ratios", "", "长宽比例")
	seed        = flag.String("seed", "", "随机种子(翻页时带上 确保不会重复)")
)

func main() {
	fmt.Println("输入CTRL+C停止程序")
	flag.Parse()
	v := url.Values{}
	if apikey != "" {
		v.Add("apikey", apikey)
	}
	if *q != "" {
		v.Add("q", *q)
	}
	if *categories != "" {
		v.Add("categories", *categories)
	}
	if *purity != "" {
		v.Add("purity", *purity)
	}
	if *sorting != "" {
		v.Add("sorting", *sorting)
	}
	if *order != "" {
		v.Add("order", *order)
	}
	if *topRange != "" {
		v.Add("topRange", *topRange)
	}
	if *atleast != "" {
		v.Add("atleast", *atleast)
	}
	if *resolutions != "" {
		v.Add("resolutions", *resolutions)
	}
	if *ratios != "" {
		v.Add("ratios", *ratios)
	}

	if *seed != "" {
		v.Add("seed", *seed)
	}
	rowReq := v.Encode()
	req, _ := url.QueryUnescape(rowReq)
	fmt.Println(req)
	url := rowUrl
	if req != "" {
		url = rowUrl + "?" + req
	}
	go Download(url, 1)
}

// 下载直到最后一页
func Download(url string, index int) {
	pUrl := url + "&page=" + strconv.Itoa(index)
	resp, err := http.Get(pUrl)
	if err != nil {
		panic(err.Error())
	}
	var data ImageList
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	_ = json.Unmarshal(body, &data)
	path := "image/" + *q
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(path, 0755)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*2)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 5))
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 5,
		},
	}
	for _, v := range data.Data {
		resp, err := client.Get(v.Path)
		if err == nil {
			name := v.Path[strings.LastIndex(v.Path, "/")+1:]
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				fmt.Println("正在下载:" + name)
				_ = ioutil.WriteFile(path+"/"+name, body, 0755)
			}
		} else {
			fmt.Println(err.Error())
		}
		resp.Body.Close()
		time.Sleep(time.Second * 2)
	}
	if len(data.Data) == 0 {
		return
	}
	Download(url, int(data.Meta.CurrentPage+1))
	return
}

type ImageList struct {
	Data []struct {
		Category   string   `json:"category"`
		Colors     []string `json:"colors"`
		CreatedAt  string   `json:"created_at"`
		DimensionX int64    `json:"dimension_x"`
		DimensionY int64    `json:"dimension_y"`
		Favorites  int64    `json:"favorites"`
		FileSize   int64    `json:"file_size"`
		FileType   string   `json:"file_type"`
		ID         string   `json:"id"`
		Path       string   `json:"path"`
		Purity     string   `json:"purity"`
		Ratio      string   `json:"ratio"`
		Resolution string   `json:"resolution"`
		ShortURL   string   `json:"short_url"`
		Source     string   `json:"source"`
		Thumbs     struct {
			Large    string `json:"large"`
			Original string `json:"original"`
			Small    string `json:"small"`
		} `json:"thumbs"`
		URL   string `json:"url"`
		Views int64  `json:"views"`
	} `json:"data"`
	Meta struct {
		CurrentPage int64       `json:"current_page"`
		LastPage    int64       `json:"last_page"`
		PerPage     int64       `json:"per_page"`
		Query       interface{} `json:"query"`
		Seed        interface{} `json:"seed"`
		Total       int64       `json:"total"`
	} `json:"meta"`
}
