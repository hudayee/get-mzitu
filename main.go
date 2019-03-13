package main

import (
	"bufio"
	"errors"
	"fmt"
	"get-mzitu/request"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type DownloadInfo struct {
	path    string
	url     string
	referer string
}

var parametersInfo = [2]string{"保存图片的目录（默认当前目录）：", "下载图集的个数（限制255以下，默认10）："}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	var directory string
	var number uint64
	for i, v := range parametersInfo {
		fmt.Print(v)
		read := bufio.NewReader(os.Stdin)
		valByte, _, err := read.ReadLine()
		if err != nil {
			panic(err.Error())
		}
		value := string(valByte)
		switch i {
		case 0: //目录
			if value == "" {
				if defaultPath, err := GetCurrentPath(); err != nil {
					panic(err.Error())
				} else {
					directory = defaultPath
				}
			} else if PathExists(string(valByte)) {
				directory = value
			} else {
				panic("没有找到该目录")
			}
		case 1: //下载个数
			if value == "" {
				number = 10
			} else {
				num, err := strconv.ParseUint(string(valByte), 10, 64)
				if err != nil {
					panic(err.Error())
				}
				number = num
			}
		default:
			panic("invalid parametersInfo[i]")
		}
	}
	fmt.Println(directory)
	fmt.Println(number)
	var list []request.List

	var pageCount int64 = 0
	//获取列表信息
	for i := int64(1); ; i++ {
		arr, _pageCount, _ := request.GetList(i)
		if i == 1 {
			pageCount = _pageCount
		}
		for _, val := range arr {
			filePath := filepath.Join(directory, val.Title)
			if !PathExists(filePath) && uint64(len(list)) < number {
				list = append(list, val)
			}
		}
		if i == pageCount || uint64(len(list)) == number {
			break
		}
	}
	// end
	//获取图集信息
	var picList []request.PicList = make([]request.PicList, len(list))
	q1 := make(chan request.PicList)
	for i, val := range list {
		go request.GetPic(val.Url, q1, i)
	}
	for range list {
		pic := <-q1
		picList[pic.Index] = pic
	}
	fmt.Printf("共%d个图集\n", len(picList))
	//end
	//下载图片
	var downloadList []DownloadInfo
	for i, val := range picList {
		//创建目录
		picPath := filepath.Join(directory, list[i].Title)
		if err := os.Mkdir(picPath, 0766); err != nil {
			fmt.Printf("%s目录创建失败\n", list[i].Title)
		}
		for j := uint64(0); j < val.PageCount; j++ {
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
	for _, val := range downloadList {
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
