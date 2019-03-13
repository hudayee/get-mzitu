package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type List struct {
	Url   string
	Title string
}
type PicList struct {
	Path      string
	Index     int
	Prefix    string
	PageCount uint64
}

func main() {
}
func GetList(number int64) (arr []List, pageCount int64, err error) { //(val string)
	// newestList: "/page/${page}/",
	// hotList: "/hot/page/${page}/"

	doc, err := goquery.NewDocument("http://www.mzitu.com")
	if err != nil {
		panic(err.Error())
	}
	pageDom := doc.Find(".nav-links .page-numbers")
	listDom := doc.Find("#pins li span a")
	listDom.Each(func(i int, s *goquery.Selection) {
		url, bool := s.Attr("href")
		title := s.Text()
		if bool {
			arr = append(arr, List{Url: url, Title: title})
		}
	})
	val := pageDom.Last().Prev().Text()
	pageCount, err = strconv.ParseInt(val, 10, 64)
	return
}
func GetPic(_url string, q chan PicList, index int) {
	doc, err := goquery.NewDocument(_url)
	if err != nil {
		panic(err.Error())
	}
	pageCount, err := strconv.ParseUint(doc.Find(".pagenavi a").Last().Prev().Text(), 10, 64)
	if err != nil {
		panic(err.Error())
	}
	pic, isExist := doc.Find(".main-image p a img").Attr("src")
	if !isExist {
		panic("not found img src")
	}
	prefix := string([]byte(path.Base(pic))[:3])

	path := strings.Replace(pic, path.Base(pic), "", 1)
	q <- PicList{PageCount: pageCount, Prefix: prefix, Index: index, Path: path}
}
func Download(path string, url string, referer string, c chan bool) {
	fmt.Println(url)
	ok := true
	defer func() {
		if err := recover(); err != nil {
			ok = false
			fmt.Println(err)
			c <- ok
		}
	}()
	client := &http.Client{}
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
