package main

import (
	"flag"
	"fmt"
	"github.com/hanbao-workspace/get-mzitu/request"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type DownloadInfo struct {
	path    string
	url     string
	referer string
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	var (
		err error
		//successLog = log.New(os.Stdout, "message：", 0)
		errLog = log.New(os.Stdout, "[ERROR]：", 0)
	)
	var _dir string
	_dir, err = os.Getwd()
	if err != nil {
		errLog.Fatal("获取当前工作路径失败")
	}
	nums := flag.Int("n", 10, "下载图集个数")
	dir := flag.String("d", _dir, "保存图片的目录")
	flag.Parse()

	if !request.IsExist(*dir) {
		errLog.Fatal(fmt.Sprintf("路径%s不存在", *dir))
	}

	var list []request.List

	//获取列表信息
	if list, err = request.GetList(*dir, *nums); err != nil {
		errLog.Fatal(err)
	}

	// end
	//获取图集信息
	var picList []*request.PicList = make([]*request.PicList, len(list))
	fmt.Printf("共%d个图集\n正在获取图集信息......\n", len(list))
	q1 := make(chan *request.PicList, 10)
	for i, val := range list {
		if i%4 == 0 { //每4组间隔一秒，防止403
			time.Sleep(1 * time.Second)
		}
		go request.GetPic(val.Url, q1, i)
	}
	successPic := 0

	for range list {
		pic := <-q1
		if pic != nil {
			successPic++
			picList[pic.Index] = pic
		}
	}
	fmt.Printf("获取图集信息成功：%d\n", successPic)
	//end
	//下载图片
	var downloadList []DownloadInfo
	for i, val := range picList {
		//创建目录
		picPath := filepath.Join(*dir, list[i].Title)
		if err := os.Mkdir(picPath, 0766); err != nil {
			fmt.Printf("%s目录创建失败\n", list[i].Title)
		}
		for j := 0; j < val.PageCount; j++ {
			name := strconv.Itoa(int(j)+1) + ".jpg"
			var downInfo DownloadInfo
			if j < 9 {
				name = "0" + name
			}
			downInfo.referer = list[i].Url
			downInfo.path = filepath.Join(picPath, name)
			downInfo.url = val.Path + val.Prefix + name
			downloadList = append(downloadList, downInfo)
		}
	}
	fmt.Println("共", len(downloadList), "张图片")
	fmt.Println("Downloadling......")
	var c chan bool = make(chan bool)
	for i, val := range downloadList {
		if i%4 == 0 {
			time.Sleep(1 * time.Second)
		}
		go request.Download(val.path, val.url, val.referer, c)
	}
	var success, failed int
	for range downloadList {
		ok := <-c
		if ok {
			success++
		} else {
			failed++
		}
	}
	fmt.Println("下载图片总数:", len(downloadList))
	fmt.Println("成功:", success)
	fmt.Println("失败:", failed)
	//end

}
