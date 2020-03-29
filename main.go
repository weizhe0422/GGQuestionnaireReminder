package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/weizhe0422/GGQuestionnaireReminder/DBUtil"
	"github.com/weizhe0422/GGQuestionnaireReminder/Model"
	"log"
	"net/http"
	"os"
)

var bot *linebot.Client

var groupID string

const (
	mongoAtlas = "mongodb+srv://gguser:true0422@cluster0-lpy0f.gcp.mongodb.net/test?retryWrites=true&w=majority"
	surveycakeURL = "https://zh.surveymonkey.com/r/EmployeeHealthCheck?fbclid=IwAR2fKoFAYPEHxwhNpxIcgFXzWXylYGcVfVRuNPS88VpKwwKi_40cavQZYFU"
)
func main() {
	var err error

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
		URL: mongoAtlas,
		Database: "GGUser",
		Collection: "QuestionReminder",
	}

	for _, event := range events {
		log.Println("EVENT TYPE:",event.Type)
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Println(message.Text)
				bot.ReplyMessage(event.ReplyToken,
					linebot.NewFlexMessage("請問你想做什麼?",
						&linebot.CarouselContainer{
							Type: linebot.FlexContainerTypeCarousel,
							Contents: []*linebot.BubbleContainer{
								{
									Type:      linebot.FlexContainerTypeCarousel,
									Size:      linebot.FlexBubbleSizeTypeGiga,
									Direction: linebot.FlexBubbleDirectionTypeLTR,
									Header:    nil,
									Hero: &linebot.ImageComponent{
										Type:  linebot.FlexComponentTypeImage,
										URL:   "https://images.pexels.com/photos/3073037/pexels-photo-3073037.jpeg?auto=compress&cs=tinysrgb&dpr=3&h=750&w=1260",
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeButton,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeButton,
												Action: linebot.NewDatetimePickerAction("設定提醒時間", "remindTime", "time",
													"07:00", "23:59", "00:00"),
											},
											&linebot.ButtonComponent{
												Type:   linebot.FlexComponentTypeButton,
												Action: linebot.NewURIAction("防疫問卷", surveycakeURL),
											},
										},
									},
									Footer: nil,
									Styles: &linebot.BubbleStyle{},
								},
								{
									Type:      linebot.FlexContainerTypeCarousel,
									Size:      linebot.FlexBubbleSizeTypeGiga,
									Direction: linebot.FlexBubbleDirectionTypeLTR,
									Header:    nil,
									Hero: &linebot.ImageComponent{
										Type:  linebot.FlexComponentTypeImage,
										URL:   "https://images.pexels.com/photos/981150/pexels-photo-981150.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeButton,
										Layout: linebot.FlexBoxLayoutTypeHorizontal,
										Contents: []linebot.FlexComponent{
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "Hello,",
											},
										},
									},
									Footer: nil,
									Styles: &linebot.BubbleStyle{},
								},
							},
						})).Do()
			}
		case linebot.EventTypePostback:
			switch event.Postback.Data{
			case "remindTime":
				registInfo := Model.User{LineId: event.Source.UserID, RemindTime: event.Postback.Params.Time}
				record, err := mongo.InsertOneRecord(registInfo)
				if err != nil {
					bot.PushMessage(event.Source.UserID,linebot.NewTextMessage("設定時間失敗，請重新嘗試!"))
				}
				bot.PushMessage(event.Source.UserID, linebot.NewTextMessage(event.Source.UserID+"預定了每天"+event.Postback.Params.Time+"提醒填寫問卷!")).Do()
				log.Println("Mongo insert info: ",record)
			}
		default:
			bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("DEFAULT")).Do()
		}
	}
}
