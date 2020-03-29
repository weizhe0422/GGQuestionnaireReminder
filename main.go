package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
)

var bot *linebot.Client

var groupID string

const mongoAtlas = "mongodb+srv://gguser:true0422@cluster0-lpy0f.gcp.mongodb.net/test?retryWrites=true&w=majority"

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

	/*mongo := &DBUtil.MongoDB{
		URL: mongoAtlas,
		Database: "GGUser",
		Collection: "QuestionReminder",
	}*/

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Println(message.Text)
				/*bot.ReplyMessage(event.ReplyToken,
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
				})).Do()*/
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
										Type: linebot.FlexComponentTypeImage,
										URL:  "https://images.pexels.com/photos/2587467/pexels-photo-2587467.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
										Align: linebot.FlexComponentAlignTypeCenter,
										Size: linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeButton,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeText,
												Action: linebot.NewURIAction("yahoo","https://tw.yahoo.com/"),
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
										Type: linebot.FlexComponentTypeImage,
										URL:  "https://images.pexels.com/photos/981150/pexels-photo-981150.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeBox,
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
		}
	}
}
