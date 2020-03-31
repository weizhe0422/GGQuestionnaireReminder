package main

import (
	"context"
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/weizhe0422/GGQuestionnaireReminder/DBUtil"
	"github.com/weizhe0422/GGQuestionnaireReminder/Model"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var bot *linebot.Client

var groupID string

const (
	mongoAtlas    = "mongodb+srv://gguser:true0422@cluster0-lpy0f.gcp.mongodb.net/test?retryWrites=true&w=majority"
	surveycakeURL = "https://zh.surveymonkey.com/r/EmployeeHealthCheck?fbclid=IwAR2fKoFAYPEHxwhNpxIcgFXzWXylYGcVfVRuNPS88VpKwwKi_40cavQZYFU"
)

func main() {
	var err error

	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)

	go func() {
		for {
			log.Println("開始提醒")
			PushAlarmMessage()
			log.Println("結束提醒")
			time.Sleep(3 * time.Minute)
		}
	}()

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func PushAlarmMessage() {
	mongo := &DBUtil.MongoDB{
		URL:        mongoAtlas,
		Database:   "GGUser",
		Collection: "QuestionReminder",
	}

	allRecord, err := mongo.FindAllRecord()
	if err != nil {
		log.Printf("failed to get all record: %v", err)
		return
	}

	timeLoc, _ := time.LoadLocation("Asia/Shanghai")
	for allRecord.Next(context.TODO()) {
		log.Println("開始檢查")
		var user Model.User2
		err := allRecord.Decode(&user)
		log.Println("ID:", user.LineId, "設定時間:", user.RemindTime)
		if err != nil {
			log.Printf("failed to decode: %v", err)
			continue
		}
		log.Println("提醒時間:", user.RemindTime.In(timeLoc))
		log.Println("現在時間(上海):", time.Now().In(timeLoc))

		log.Printf("上次提醒時間:%v", user.LastRemindTime.In(timeLoc))

		if user.RemindTime.In(timeLoc).Before(time.Now().In(timeLoc)) {
			log.Println("開始發送提醒")
			_, err := bot.PushMessage(user.LineId, linebot.NewTextMessage("記得去填問卷啊！"+surveycakeURL)).Do()
			if err != nil {
				log.Printf("推送提提醒給%s失敗:%v", user.LineId, err)
				continue
			}

			log.Printf("推送提提醒給%s成功", user.LineId)
			tomorrow := time.Now().AddDate(0,0,1)
			_, err = mongo.UpdateRecord(bson.M{"lineid": user.LineId},
										bson.M{"$set": bson.M{"lastremindtime": time.Now().In(timeLoc),
											                   "remindtime": time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
																   user.RemindTime.Hour()+8,user.RemindTime.Minute(),0,0,timeLoc)}})
			if err != nil {
				log.Printf("更新提醒時間失敗:%v", err)
			}
		} else {
			log.Printf("%s尚未到提醒時間%s", user.LineId, user.RemindTime.In(timeLoc))
		}
		log.Println("結束檢查")
	}
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
		URL:        mongoAtlas,
		Database:   "GGUser",
		Collection: "QuestionReminder",
	}

	for _, event := range events {
		log.Println("EVENT TYPE:", event.Type)
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
			switch event.Postback.Data {
			case "remindTime"://event.Postback.Params.Time
				tomorrow := time.Now().AddDate(0,0,0)
				setHour,_ := strconv.Atoi(event.Postback.Params.Time[0:2])
				setMin,_ := strconv.Atoi(event.Postback.Params.Time[3:5])
				timeLoc, _ := time.LoadLocation("Asia/Shanghai")
				registInfo := Model.User{LineId: event.Source.UserID,
										 RemindTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
										 						setHour,setMin,0,0,timeLoc)}
				log.Println("設定時間:",time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
					setHour,setMin,0,0,time.Now().Location()))
				record, err := mongo.InsertOneRecord(registInfo)
				if err != nil {
					bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("設定時間失敗，請重新嘗試!"))
				}
				bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("預定了每天"+event.Postback.Params.Time+"提醒填寫問卷!")).Do()
				log.Println("Mongo insert info: ", record)
			}
		default:
			bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("DEFAULT")).Do()
		}
	}
}
