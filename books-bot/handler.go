package main

import (
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/url"
	"os"
	"strings"
)

type Handler struct {
	bot          *tgbotapi.BotAPI
	bookSearcher *GlobalSearchBookService
}

func NewHandler(bot *tgbotapi.BotAPI, bookSearcher *GlobalSearchBookService) *Handler {
	return &Handler{bot: bot, bookSearcher: bookSearcher}
}

var usersWhiteList = [][]string{
	{"tommytoo", "Timur"},
}

func isUserAllowed(update *tgbotapi.Update) bool {
	for _, userInfo := range usersWhiteList {
		if updateFrom(update, userInfo[0]) {
			return true
		}
	}
	return false
}

func updateFrom(update *tgbotapi.Update, username string) bool {
	if update.Message != nil {
		return update.Message.From.UserName == username
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.UserName == username
	}

	return false
}

func (h *Handler) BotProblem(update *tgbotapi.Update) {
	h.SendText(update.Message.Chat.ID, "bot problem")
}

func (h *Handler) HandleUpdate(update *tgbotapi.Update) error {
	if !isUserAllowed(update) {
		return fmt.Errorf("NOT ALLOWED")
	}
	if update.CallbackQuery != nil {
		return h.HandleCallback(update)
	}
	if update.Message.Text == "" {
		return fmt.Errorf("empty message")
	}

	chatId := update.Message.Chat.ID
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "status":
			h.SendText(chatId, "ok")
		case "state":
			h.SendState(chatId)
		default:
			h.SendText(chatId, "Try /help")
		}

		return fmt.Errorf("ukndown command")
	}

	err := h.SearchBook(update)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) SearchBook(update *tgbotapi.Update) error {
	chatId := update.Message.Chat.ID
	loadingMsg, _ := h.SendText(chatId, "_searching started..._")
	findBookResult, err := h.bookSearcher.Find(update.Message.Text)
	if err != nil {
		return fmt.Errorf("can't find book: %v", err)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(
		loadingMsg.Chat.ID,
		loadingMsg.MessageID,
	)
	h.bot.Send(deleteMsg)

	if len(findBookResult.Books) == 0 {
		h.bot.Send(tgbotapi.NewMessage(
			loadingMsg.Chat.ID,
			"Unfortunately books are not found",
		))
		return nil
	}

	for _, option := range findBookResult.Books {
		resultText := ""
		resultText += fmt.Sprintf("**%s**\n", option.Title)
		resultText += fmt.Sprintf("author: %s\n", option.Author)
		resultText += fmt.Sprintf("format: %s\n", option.Format)
		resultText += fmt.Sprintf("source: [%s](%s)\n", option.Source, option.Link)
		println("link", fmt.Sprintf("source: [%s](%s)\n", option.Source, option.Link))
		resultText += fmt.Sprintf("[google info](%s)\n", fmt.Sprintf("https://www.google.com/search?q=book short summary %s", url.PathEscape(option.Title)))

		bookMsg := tgbotapi.NewMessage(chatId, "")
		bookMsg.Text = resultText
		bookMsg.ParseMode = tgbotapi.ModeMarkdown
		downloadLinkEncodedB64 := base64.StdEncoding.EncodeToString([]byte(option.DownloadLink))
		bookMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"Download", fmt.Sprintf("download|_|%s|_|%s", option.Source, downloadLinkEncodedB64),
				),
			),
		)
		h.bot.Send(bookMsg)
	}

	return nil
}

func (h *Handler) HandleCallback(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	data := update.CallbackData()
	if data == "" {
		return fmt.Errorf("empty data chatId: %v", chatId)
	}

	dataParts := strings.Split(data, "|_|")
	action := dataParts[0]

	switch action {
	case "download":
		source := dataParts[1]
		linkB64 := dataParts[2]
		linkBytes, _ := base64.StdEncoding.DecodeString(linkB64)
		link := string(linkBytes)

		downloadBookInfo, err := h.bookSearcher.Download(&GlobalDownloadBookInput{
			Source:       source,
			DownloadLink: link,
		})
		if err != nil {
			e := fmt.Errorf("can't download book by button: %v", err)
			h.SendText(chatId, e.Error())
			return e
		}
		if downloadBookInfo == nil {
			e := fmt.Errorf("don't know why, but book can't be downloaded and no error happened")
			h.SendText(chatId, e.Error())
			return e
		}
		f, err := os.Open(downloadBookInfo.FilePath)
		if err != nil {
			return fmt.Errorf("can't open downloaded file %s: %v", downloadBookInfo.FilePath, err)
		}
		reader := tgbotapi.FileReader{Name: downloadBookInfo.FileName, Reader: f}

		h.bot.Send(tgbotapi.NewDocument(chatId, reader))
		e := os.Remove(downloadBookInfo.FilePath)
		if e != nil {
			log.Fatal(e)
		}
	default:
		h.bot.Send(tgbotapi.NewMessage(chatId, "Unknown button"))
		return nil
	}

	return nil
}

func (h *Handler) SendState(chatId int64) {
	h.SendText(chatId, "state")
}

func (h *Handler) SendText(chatId int64, text string) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"

	if sentMsg, err := h.bot.Send(msg); err != nil {
		log.Panic(err)
		return nil, err
	} else {
		return &sentMsg, nil
	}
}
