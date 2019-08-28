package request

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type List struct {
	Url   string
	Title string
}
type PicList struct {
	Path      string
	Index     int
	Prefix    string
	PageCount int
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
func GetList(dir string, nums int) (arr []List, err error) {
	page := 1
	pageCount := 1
	url := "http://www.mzitu.com"
	for ; page <= pageCount; page++ {

		if page > 1 {
			url = fmt.Sprintf("http://www.mzitu.com/page/%d", page)
		}
		if page%4 == 0 { //每4组间隔一秒，防止403
			time.Sleep(1 * time.Second)
		}
		var doc *goquery.Document
		doc, err = goquery.NewDocument(url)
		if err != nil {
			return
		}
		pageDom := doc.Find(".nav-links .page-numbers")
		listDom := doc.Find("#pins li span a")
		listDom.Each(func(i int, s *goquery.Selection) {
			url, bool := s.Attr("href")
			title := s.Text()
			filePath := filepath.Join(dir, title)
			if IsExist(filePath) || !bool {
				return
			}
			arr = append(arr, List{Url: url, Title: title})
		})
		val := pageDom.Last().Prev().Text()
		if pageCount, err = strconv.Atoi(val); err != nil {
			return
		}
		if len(arr) >= nums {
			break
		}
	}
	return arr[0:nums], nil
}

func GetPic(_url string, q chan *PicList, index int) {
	doc, err := goquery.NewDocument(_url)
	if err != nil {
		q <- nil
		return
	}
	pageInfo := doc.Find(".pagenavi a").Last().Prev().Text()
	if pageInfo == "" {
		q <- nil
		return
	}
	pageCount, err := strconv.Atoi(pageInfo)
	if err != nil {
		q <- nil
		return
	}
	pic, isExist := doc.Find(".main-image p a img").Attr("src")
	if !isExist {
		q <- nil
		return
	}
	prefix := string([]byte(path.Base(pic))[:3])

	s := strings.Replace(pic, path.Base(pic), "", 1)
	q <- &PicList{PageCount: pageCount, Prefix: prefix, Index: index, Path: s}
}

func Download(path string, url string, referer string, c chan bool) {
	ok := true
	defer func() {
		if err := recover(); err != nil {
			ok = false
			fmt.Println(err)
			c <- ok
		}
	}()
	client := &http.Client{}
	fmt.Println("Downloading：" + url)
	req, _ := http.NewRequest("GET", url, strings.NewReader(""))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36")
	req.Header.Set("Host", "www.mzitu.com")
	req.Header.Set("Referer", referer)
	res, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	if res.StatusCode != 200 {
		statusCode := strconv.Itoa(res.StatusCode)
		str := url + ":" + "http StatusCode:" + statusCode
		panic(str)
	}
	file, err := os.Create(path)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}
	if _, err := file.Write(fileBytes); err != nil {
		panic(err.Error())
	}
	c <- ok
}
