package main

//
//import (
//	"fmt"
//	"github.com/PuerkitoBio/goquery"
//	"log"
//	"os"
//	"strings"
//)
//
//func main() {
//	dat, err := os.Open("data.txt")
//
//	doc, err := goquery.NewDocumentFromReader(dat)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	goodFormatBooks := doc.Find(".bookRow").FilterFunction(func(i int, selection *goquery.Selection) bool {
//		foundSelection := selection.Find(".property_value").FilterFunction(func(i2 int, selection2 *goquery.Selection) bool {
//			someText := strings.ToLower(selection2.Text())
//			return strings.Contains(someText, "epub") || strings.Contains(someText, "fb2")
//		})
//		return len(foundSelection.Nodes) != 0
//	})
//
//	var book *goquery.Selection
//	if len(goodFormatBooks.Nodes) != 0 {
//		book = goodFormatBooks.First()
//	} else {
//		book = doc.Find(".bookRow").First()
//	}
//
//	bookLinkElems := book.Find("a")
//	bookNameLinkElem := bookLinkElems.Eq(1)
//	bookName := bookNameLinkElem.Text()
//	bookAuthor := bookLinkElems.Eq(2).Text()
//	bookLink, _ := bookNameLinkElem.Attr("href")
//
//	fmt.Println(bookName, bookAuthor, bookLink)
//}
