package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Tommytto/books-bot/utils"
	"io"
	"log"
	"mime"
	"net/url"
	"os"
	"strings"
)

type SearchCredentials struct {
	Domain string
	Cookie string
}

var searchCredentials = []*SearchCredentials{
	{
		"https://lib-4ieh23xxghrlxscap25txeri.1lib.ch",
		"remix_userid=32902253; remix_userkey=719fc82226667a250acf258754d73e13; pdSwitchToArticlesLink=https%3A%2F%2Farticles-jhc5wbohnea3hab5vg632pjy.1lib.fr; siteLanguageV2=ru; proxiesNotWorking=; domainsNotWorking=",
	},
	{
		"https://lib-ea5bvsn65n3prrfbuunlzej3.booksc.eu",
		"remix_userid=32920217; remix_userkey=12c62de6fa70000df9f8d104dd10d06a; pdSwitchToArticlesLink=https%3A%2F%2Farticles-honrnw4bpp57wgomjdxbv6rk.booksc.eu; siteLanguageV2=ru; proxiesNotWorking=; domainsNotWorking=",
	},
}

var activeCredsIndex = 0

type ZLibrarySearcher struct {
}

func (z *ZLibrarySearcher) Find(searchText string) ([]*FoundBookInfo, error) {
	var err error
	page, err := z.searchRequest(searchText)
	if err != nil {
		return nil, fmt.Errorf("zlibrary searsh request error: %v", err)
	}

	foundBooks := z.findBooksOnPage(searchText, page)
	if foundBooks == nil {
		return nil, fmt.Errorf("book is not found on page")
	}

	var response []*FoundBookInfo
	for _, bookData := range foundBooks {
		response = append(response, bookData)
	}

	return foundBooks, nil
}

func (z *ZLibrarySearcher) DownloadByLink(pageLink string) (*DownloadBookInfo, error) {
	pageLink = z.GetCreds().Domain + pageLink
	downloadPage, err := z.getPage(pageLink)
	if err != nil {
		return nil, fmt.Errorf("can't get book page: %v", err)
	}

	var downloadLink string
	downloadLink, err = z.getDownloadLink(downloadPage)
	if errors.Is(err, LimitError) {
		//for z.UpdateCreds() {
		//	fullResult, err := z.Find(searchText)
		//	if err == nil {
		//		return fullResult, err
		//	} else {
		//		log.Println("retry error", err)
		//	}
		//}
		return nil, LimitError
	}
	if downloadLink == "" || err != nil {
		log.Printf("download link not found on page for %s", pageLink)
	}

	filePath, fileName, err := z.downloadBook(downloadLink)
	if err != nil {
		return nil, fmt.Errorf("can't download book: %v", err)
	}

	return &DownloadBookInfo{
		FilePath: filePath,
		FileName: fileName,
	}, nil
}

func (z *ZLibrarySearcher) GetId() string {
	return "zlibrary"
}

func (z *ZLibrarySearcher) getPage(url string) (string, error) {
	creds := z.GetCreds()
	headers := getHeaders(creds.Cookie, nil)
	response, err := utils.GetRequest(url, headers)
	if err != nil {
		return "", err
	}

	return string(response), nil
}

func (z *ZLibrarySearcher) GetCreds() *SearchCredentials {
	return searchCredentials[activeCredsIndex]
}

func (z *ZLibrarySearcher) UpdateCreds() bool {
	nextIndex := activeCredsIndex + 1
	if len(searchCredentials)-1 < nextIndex {
		return false
	}

	activeCredsIndex = nextIndex
	return true
}

func (z *ZLibrarySearcher) downloadBook(bookUrl string) (string, string, error) {
	creds := z.GetCreds()
	requestUrl := strings.ReplaceAll(creds.Domain+bookUrl, "/book/", "/dl/")

	getBookRes, err := utils.GetRequestRaw(requestUrl, getHeaders(creds.Cookie, nil))
	if err != nil {
		return "", "", fmt.Errorf("can't get download redirect: %v", err)
	}
	defer getBookRes.Body.Close()

	_, params, err := mime.ParseMediaType(getBookRes.Header.Get("Content-Disposition"))
	downloadedFileName := params["filename"]

	urlParts := strings.Split(bookUrl, "/")
	filePath := "__tmp_files_" + strings.Join(urlParts[len(urlParts)-2:], "_")
	out, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("can't create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, getBookRes.Body)
	if err != nil {
		return "", "", fmt.Errorf("can't copy file: %v", err)
	}

	return filePath, downloadedFileName, nil
}

func (z *ZLibrarySearcher) findBooksOnPage(searchText string, page string) []*FoundBookInfo {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Fatal(err)
	}

	type pageBookInfo struct {
		selection           *goquery.Selection
		normHammingDistance float64
		format              string
		name                string
		author              string
		link                string
	}
	var pageBookList []*pageBookInfo
	doc.Find(".bookRow").Each(func(i int, selection *goquery.Selection) {
		if len(pageBookList) == 5 {
			return
		}
		bookFormat := ""
		selection.Find(".property_value").Each(func(i2 int, selection2 *goquery.Selection) {
			someText := strings.ToLower(selection2.Text())
			if strings.Contains(someText, "epub") {
				bookFormat = "epub"
			} else if strings.Contains(someText, "fb2") {
				bookFormat = "fb2"
			} else if strings.Contains(someText, "pdf") {
				bookFormat = "pdf"
			}
		})
		bookLinkElems := selection.Find("a")
		bookNameLinkElem := bookLinkElems.Eq(1)
		bookName := bookNameLinkElem.Text()
		bookAuthor := selection.Find(".authors a").First().Text()
		bookLink, _ := bookNameLinkElem.Attr("href")
		//isGoodBook := utils.CloseEnoughHamming(bookName, searchText, 0.8)
		b := &pageBookInfo{
			selection: selection,
			name:      bookName,
			link:      bookLink,
			author:    bookAuthor,
			format:    bookFormat,
		}
		pageBookList = append(pageBookList, b)
	})

	if len(pageBookList) == 0 {
		return nil
	}

	var response []*FoundBookInfo

	for _, book := range pageBookList {
		response = append(response, &FoundBookInfo{
			Source:       z.GetId(),
			Link:         z.GetCreds().Domain + book.link,
			DownloadLink: book.link,
			Title:        book.name,
			Author:       book.author,
			Format:       book.format,
		})
	}

	return response
}

var LimitError = errors.New("zlibrary limit")

func (z *ZLibrarySearcher) getDownloadLink(page string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return "", fmt.Errorf("read page problem: %v", err)
	}

	limitText := doc.Find(".user-data__stats-value").Eq(0).Text()
	limitText = strings.TrimSpace(limitText)
	limitTextParts := strings.Split(limitText, "/")

	if len(limitTextParts) == 2 {
		if limitTextParts[0] == limitTextParts[1] {
			return "", LimitError
		}
	}

	downloadBtnContainer := doc.Find(".book-details-button").First()
	downloadLink, _ := downloadBtnContainer.Find("a").First().Attr("href")

	if downloadLink == "" {
		return "", fmt.Errorf("empty download link for")
	}

	return downloadLink, nil
}

func getHeaders(cookie string, customHeaders *map[string]string) map[string]string {
	result := map[string]string{
		"cookie": cookie,
	}

	if customHeaders != nil {
		for k, v := range *customHeaders {
			result[k] = v
		}
	}

	return result
}

func (z *ZLibrarySearcher) searchRequest(text string) (string, error) {
	creds := z.GetCreds()
	requestUrl := creds.Domain + "/s/" + url.PathEscape(text)

	return z.getPage(requestUrl)
}

var curlReq = "curl 'https://lib-4ieh23xxghrlxscap25txeri.1lib.ch/s/Don%20Quixote%20by%20Miguel%20de%20Cervantes' \\\n  -H 'authority: lib-4ieh23xxghrlxscap25txeri.1lib.ch' \\\n  -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \\\n  -H 'accept-language: ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7' \\\n  -H 'cache-control: no-cache' \\\n  -H 'cookie: remix_userid=32902253; remix_userkey=719fc82226667a250acf258754d73e13; pdSwitchToArticlesLink=https%3A%2F%2Farticles-vvyhpgcvbhmsfciilgv3qa5p.1lib.ch; siteLanguageV2=ru' \\\n  -H 'pragma: no-cache' \\\n  -H 'referer: https://lib-4ieh23xxghrlxscap25txeri.1lib.ch/' \\\n  -H 'sec-ch-ua: \"Google Chrome\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24\"' \\\n  -H 'sec-ch-ua-mobile: ?0' \\\n  -H 'sec-ch-ua-platform: \"macOS\"' \\\n  -H 'sec-fetch-dest: document' \\\n  -H 'sec-fetch-mode: navigate' \\\n  -H 'sec-fetch-site: same-origin' \\\n  -H 'sec-fetch-user: ?1' \\\n  -H 'upgrade-insecure-requests: 1' \\\n  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36' \\\n  --compressed"
var curlReq2 = "curl 'https://lib-ea5bvsn65n3prrfbuunlzej3.booksc.eu/book/2041681/4c425c?dsource=mostpopular' \\\n  -H 'authority: lib-ea5bvsn65n3prrfbuunlzej3.booksc.eu' \\\n  -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \\\n  -H 'accept-language: ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7' \\\n  -H 'cache-control: no-cache' \\\n  -H 'cookie: remix_userid=32920217; remix_userkey=12c62de6fa70000df9f8d104dd10d06a; pdSwitchToArticlesLink=https%3A%2F%2Farticles-honrnw4bpp57wgomjdxbv6rk.booksc.eu; siteLanguageV2=ru' \\\n  -H 'pragma: no-cache' \\\n  -H 'referer: https://lib-ea5bvsn65n3prrfbuunlzej3.booksc.eu/' \\\n  -H 'sec-ch-ua: \"Google Chrome\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24\"' \\\n  -H 'sec-ch-ua-mobile: ?0' \\\n  -H 'sec-ch-ua-platform: \"macOS\"' \\\n  -H 'sec-fetch-dest: document' \\\n  -H 'sec-fetch-mode: navigate' \\\n  -H 'sec-fetch-site: same-origin' \\\n  -H 'sec-fetch-user: ?1' \\\n  -H 'upgrade-insecure-requests: 1' \\\n  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36' \\\n  --compressed"

func NewZLibrarySearcher() *ZLibrarySearcher {
	return &ZLibrarySearcher{}
}
