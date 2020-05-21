package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//var FileLogger *log.Logger
var File *os.File
var WorkDir string

func init() {
	var err error
	File, err = os.OpenFile("links.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	WorkDir, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}
	WorkDir, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println("WorkDir: " + WorkDir)
	//defer f.Close()

	//TODO: create folders - data, logs, downloads
}

func appendToFile(text string) {
	if _, err := File.WriteString(text + "\n"); err != nil {
		log.Println(err)
	}
}

func main() {
	//TODO: downloads pages with pagination
	//TODO:fix link codes to full link
	//TODO:add workers for downloads
	//TODO:add logger
	//TODO:remove connection error panic
	file, err := os.Open("codes.txt")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	file.Close()
	for _, eachline := range txtlines {
		downloadPage(eachline)
		fmt.Println(eachline)
	}

	defer File.Close()

	/*
		codeList := []string{"37330", "27624"}
		for _, code := range codeList {
			downloadPage (code)
		}
	*/
}

func downloadPage(code string) {
	url := "http://www.itmtu.net/mm/" + code + "/"
	log.Printf("Download of %s\n", url)
	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		//code := resp.StatusCode
		log.Printf("Status: %v Error: %v \n", err.Error())
		//panic(err)
	}
	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read body error: %v \n", err.Error())
	}
	getPageData(string(html), code)
}

func getPageData(body string, code string) {
	validLink := ""
	linkCode := ""
	var validLinkRE = regexp.MustCompile(`id\=\"image_div\".*?img src=\"([^\"]*)"`)
	//validLink := string(validLinkRE.Find(html))
	validLinkMatches := validLinkRE.FindAllStringSubmatch(body, -1)
	for _, validLinkMatch := range validLinkMatches {
		for _, validLinkOfMatch := range validLinkMatch {
			validLink = validLinkOfMatch
		}
	}
	//validLink := validLinkRE.FindAllStringSubmatch(body, -1)[0][1]
	if validLink != "" {

		linkCodeRE := regexp.MustCompile(`mm\/([^\/]?\/?[^\/]*\/[^\/]*\/[^\/]*)\/[^\/]*\.jpg`)
		linkCodeMatches := linkCodeRE.FindAllStringSubmatch(validLink, -1)
		for _, linkCodeMatch := range linkCodeMatches {
			for _, linkCodeOfMatch := range linkCodeMatch {
				linkCode = linkCodeOfMatch
			}
		}
		//linkCode = linkCodeRE.FindAllStringSubmatch(body, -1)[0][1]
		if linkCode != "" {
			var lastPageRE = regexp.MustCompile(`em><a class="page-numbers" href=".mm.` + code + `.(\d+)"`)
			lastPage := lastPageRE.FindAllStringSubmatch(body, -1)[0][1]
			log.Printf("validLink %v lastPage %v linkCode %v \n", validLink, lastPage, linkCode)
			lastPageInt, _ := strconv.Atoi(lastPage)
			//link := fmt.Sprintf("http://img.itmtu.cc/mm/s/slct/%s/%04d.jpg \n", linkCode, lastPageInt)
			//log.Printf("Last link lastPage %v \n", link)
			generateLinks(linkCode, lastPageInt)
		} else {
			fmt.Println("linkCode not found!")
		}
	} else {
		fmt.Println("Link not found!")
	}
	//fmt.Println(validLinkRE.FindAllString(body, -1))
	//fmt.Println(validLinkRE.MatchString(body))
	// show the HTML code as a string %s
	//fmt.Printf("%s\n", html)
}

func generateLinks(linkCode string, lastPageInt int) []string {
	var links []string
	folderName := strings.ReplaceAll(linkCode, "/", "_")
	for i := 1; i <= lastPageInt; i++ {
		var link = fmt.Sprintf("http://img.itmtu.cc/mm/%s/%04d.jpg", linkCode, i)
		if i == 1 {
			fmt.Print(link + " ")
		}
		//fmt.Println(link + " ")
		links = append(links, link)
		appendToFile(link)
		dir := WorkDir + string(os.PathSeparator) + folderName
		////time.Sleep(time.Second * 5)
		err := DownloadFile(dir, dir+string(os.PathSeparator)+folderName+"-"+strconv.Itoa(i)+".jpg", link)
		////var err error
		////panic(err)
		fmt.Print(".")
		if err != nil {
			log.Printf("DL err: %v  %v\n", err, dir)
		}
	}
	log.Printf("\nSaved links\n")
	return links
}

func DownloadFile(folderPath string, filepath string, url string) error {
	//fmt.Printf("%s %s %s \n", folderPath, filepath, url )
	os.MkdirAll(folderPath, os.ModePerm)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	fmt.Print("[" + strconv.Itoa(resp.StatusCode) + "] ")
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
