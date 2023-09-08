package main

import (
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

type FlibustaSearcher struct{}

const flibustaHost = "http://proxy.flibusta.is"

func NewFlibustaSearcher() *FlibustaSearcher {
	return &FlibustaSearcher{}
}

func (f *FlibustaSearcher) GetId() string {
	return "flibusta"
}

func (f *FlibustaSearcher) Find(text string) ([]*FoundBookInfo, error) {
	requestUrl := flibustaHost + "/booksearch?ask=" + url.PathEscape(text)
	response, err := utils.GetRequest(requestUrl, nil)
	if err != nil {
		return nil, err
	}

	pageText := string(response)
	println(requestUrl)

	foundBooksOnPage := f.findBooksOnPage(pageText)

	return foundBooksOnPage, nil
}

func (f *FlibustaSearcher) DownloadByLink(link string) (*DownloadBookInfo, error) {
	res, e := f.downloadBook(flibustaHost + link)
	return res, e
}

func (f *FlibustaSearcher) downloadBook(link string) (*DownloadBookInfo, error) {
	response, err := utils.GetRequest(link, nil)
	if err != nil {
		return nil, err
	}

	bookPageText := string(response)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bookPageText))
	if err != nil {
		log.Fatal(err)
	}

	fb2Link, _ := doc.Find("a:contains('fb2')").Attr("href")
	epubLink, _ := doc.Find("a:contains('epub')").Attr("href")
	pdfLink, _ := doc.Find("a:contains('pdf')").Attr("href")

	var lastErr error
	if fb2Link != "" {
		res, err := f.downloadBookFile(flibustaHost + fb2Link)
		if err == nil {
			return res, err
		}
		lastErr = err
	}
	if epubLink != "" {
		res, err := f.downloadBookFile(flibustaHost + epubLink)
		if err == nil {
			return res, err
		}
		lastErr = err
	}
	if pdfLink != "" {
		res, err := f.downloadBookFile(flibustaHost + pdfLink)
		if err == nil {
			return res, err
		}
		lastErr = err
	}

	return nil, lastErr
}

func (f *FlibustaSearcher) downloadBookFile(link string) (*DownloadBookInfo, error) {
	getBookRes, err := utils.GetRequestRaw(link, nil)
	if err != nil {
		return nil, fmt.Errorf("can't get download redirect: %v", err)
	}
	defer getBookRes.Body.Close()

	_, params, err := mime.ParseMediaType(getBookRes.Header.Get("Content-Disposition"))
	fileName := params["filename"]

	filePath := "__tmp_files_" + fileName
	out, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, getBookRes.Body)
	if err != nil {
		return nil, fmt.Errorf("can't copy file: %v", err)
	}

	return &DownloadBookInfo{
		FilePath: filePath,
		FileName: fileName,
	}, nil
}

func (f *FlibustaSearcher) findBooksOnPage(page string) []*FoundBookInfo {
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
	doc.Find("h3:contains('Найденные книги') + ul > li").Each(func(i int, selection *goquery.Selection) {
		if len(pageBookList) == 5 {
			return
		}

		bookNameLink := selection.Find("a").First()
		bookName := ""
		if bookNameLink == nil {
			return
		}
		bookName = bookNameLink.Text()
		bookFormat := "available after download"
		bookLink, _ := bookNameLink.Attr("href")
		bookAuthor := ""
		bookElementParts := strings.Split(selection.Text(), "-")
		if len(bookElementParts) >= 2 {
			bookAuthor = bookElementParts[1]
		}

		b := &pageBookInfo{
			selection: selection,
			name:      bookName,
			link:      bookLink,
			author:    bookAuthor,
			format:    bookFormat,
		}
		pageBookList = append(pageBookList, b)
	})

	var result []*FoundBookInfo

	for _, book := range pageBookList {
		result = append(result, &FoundBookInfo{
			Source:       f.GetId(),
			Link:         flibustaHost + book.link,
			DownloadLink: book.link,
			Title:        book.name,
			Author:       book.author,
			Format:       book.format,
		})
	}

	return result
}
