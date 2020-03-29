package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/weizhe0422/GGQuestionnaireReminder/DBUtil"
	"log"
	"net/http"
	"os"
)

var bot *linebot.Client

var groupID string

func main() {
	var err error

	mongo.Connect(contect.)




	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	mongo := &DBUtil.MongoDB{
		URL: "mongodb+srv://gguser:true0422@cluster0-lpy0f.gcp.mongodb.net/test?retryWrites=true&w=majority",
		Database: "GGUser",
		Collection: "QuestionReminder",
	}

	for _, event := range events {
		find, err := mongo.FindRecord("073300")
		mongo.InsertOneRecord("073300")
		bot.PushMessage(event.Source.UserID,linebot.NewTextMessage(fmt.Sprintf("找紀錄: %s / %v", find, err))).Do()
		/*if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Println(message.Text)
				bot.ReplyMessage(event.ReplyToken,
					linebot.NewFlexMessage("你想設定什麼呢?", &linebot.BubbleContainer{
						Type:linebot.FlexContainerTypeCarousel,
						Body:&linebot.BoxComponent{
							Type:     linebot.FlexComponentTypeButton,
							Layout:   linebot.FlexBoxLayoutTypeHorizontal,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:    linebot.FlexComponentTypeText,
									Text: 	"防疫小幫手",
								},
								&linebot.TextComponent{
									Type:    linebot.FlexComponentTypeText,
									Text: 	"其他",
								},
							},
						},
					})).Do()
			}
		}*/
	}
}
