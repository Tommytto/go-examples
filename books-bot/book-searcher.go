package main

import (
	"fmt"
	"log"
)

type GlobalSearchBookService struct {
	BookSearchers []BookSearcher
}

func NewGlobalSearchBookService(bookSearchers []BookSearcher) *GlobalSearchBookService {
	return &GlobalSearchBookService{BookSearchers: bookSearchers}
}

type FoundBookInfo struct {
	Source       string
	Link         string
	DownloadLink string
	Title        string
	Author       string
	Format       string
}

type DownloadBookInfo struct {
	FilePath string
	FileName string
}

type FindBookOutput struct {
	Books []*FoundBookInfo
}

func (b *GlobalSearchBookService) Find(text string) (*FindBookOutput, error) {
	output := &FindBookOutput{
		Books: []*FoundBookInfo{},
	}
	for _, searcher := range b.BookSearchers {
		result, err := searcher.Find(text)
		if err != nil {
			log.Printf("can't find book: %s with searcher %v: %v", text, searcher, err)
		} else {
			for _, book := range result {
				output.Books = append(output.Books, book)
			}
		}

		if len(output.Books) > 5 {
			break
		}
	}

	return output, nil
}

type GlobalDownloadBookInput struct {
	Source       string
	DownloadLink string
}

func (b *GlobalSearchBookService) Download(input *GlobalDownloadBookInput) (*DownloadBookInfo, error) {
	var bookSearcher BookSearcher

	for _, bs := range b.BookSearchers {
		if bs.GetId() == input.Source {
			bookSearcher = bs
		}
	}

	if bookSearcher == nil {
		return nil, fmt.Errorf("unknown book searcher %s", input.Source)
	}

	return bookSearcher.DownloadByLink(input.DownloadLink)
}

type BookSearcher interface {
	GetId() string
	Find(text string) ([]*FoundBookInfo, error)
	DownloadByLink(text string) (*DownloadBookInfo, error)
}
