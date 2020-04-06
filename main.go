package main

import (
	"context"
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/weizhe0422/GGQuestionnaireReminder/Controller"
	"github.com/weizhe0422/GGQuestionnaireReminder/DBUtil"
	"github.com/weizhe0422/GGQuestionnaireReminder/Model"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var bot *linebot.Client

var groupID string

const (
	mongoAtlas        = "mongodb+srv://gguser:true0422@cluster0-lpy0f.gcp.mongodb.net/test?retryWrites=true&w=majority"
	surveycakeURL     = "https://zh.surveymonkey.com/r/EmployeeHealthCheck?fbclid=IwAR2fKoFAYPEHxwhNpxIcgFXzWXylYGcVfVRuNPS88VpKwwKi_40cavQZYFU"
	healthydeclareURL = "https://zh.surveymonkey.com/r/EmployeeHealthDeclarationForm"
	PharmacyInfoURL   = "https://raw.githubusercontent.com/kiang/pharmacies/master/json/points.json"
	MapURL 			  = "https://fysh711426.github.io/iMaskMap/index.html?longitude=%f&latitude=%f"
)

var imageURLPage1, imageURLPage2, pharmacyImgURL []string
var tempPharmacyList []Model.PharmacyInfo
var pharmacyInfoUtil Controller.PharmacyInfo

func main() {
	var err error

	imageURLPage1 = []string{
		"https://images.pexels.com/photos/307008/pexels-photo-307008.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		"https://images.pexels.com/photos/3992952/pexels-photo-3992952.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		"https://images.pexels.com/photos/3873197/pexels-photo-3873197.jpeg?auto=compress&cs=tinysrgb&dpr=2&w=500",
	}
	imageURLPage2 = []string{
		"https://images.pexels.com/photos/3952231/pexels-photo-3952231.jpeg?auto=compress&cs=tinysrgb&dpr=2&w=500",
		"https://images.pexels.com/photos/3902732/pexels-photo-3902732.jpeg?auto=compress&cs=tinysrgb&dpr=3&h=750&w=1260",
	}
	pharmacyImgURL = []string {
		"https://images.pexels.com/photos/419585/chemist-pharmacy-singer-man-419585.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		"https://images.pexels.com/photos/3943882/pexels-photo-3943882.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		"https://images.pexels.com/photos/3962264/pexels-photo-3962264.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		"https://images.pexels.com/photos/3962285/pexels-photo-3962285.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
		"https://images.pexels.com/photos/3958125/pexels-photo-3958125.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260",
	}

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

	pharmacyInfoUtil = Controller.PharmacyInfo{URL: PharmacyInfoURL}
	log.Println("開始更新藥局資訊")
	tempPharmacyList = []Model.PharmacyInfo{}
	tempPharmacyList, _ = RefreshPharmacyList()
	log.Println("結束更新藥局資訊")
	go func() {
		for {
			log.Println("開始更新藥局資訊")
			tempPharmacyList = []Model.PharmacyInfo{}
			tempPharmacyList, _ = RefreshPharmacyList()
			log.Println("結束更新藥局資訊")
			time.Sleep(3 * time.Minute)
		}
	}()

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func RefreshPharmacyList() ([]Model.PharmacyInfo, error) {
	resp, err := pharmacyInfoUtil.GetPharmacyResp()
	if err != nil {
		log.Printf("failed to get pharmacy service response: %v", err)
		return nil, err
	}
	pharmacyInfoList, err := pharmacyInfoUtil.GetPharmacyInfoList(resp)
	if err != nil {
		log.Printf("failed to get pharmacy parsing result: %v", err)
		return nil, err
	}
	return pharmacyInfoList, nil
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
		log.Println("ID:", user.LineId, "設定時間:", user.SettingRemindTime)
		if err != nil {
			log.Printf("failed to decode: %v", err)
			continue
		}
		log.Println("提醒時間:", user.NextRemindTime.In(timeLoc))
		log.Println("現在時間(上海):", time.Now().In(timeLoc))

		log.Printf("上次提醒時間:%v", user.LastRemindTime.In(timeLoc))

		if user.NextRemindTime.In(timeLoc).Before(time.Now().In(timeLoc)) {
			log.Println("開始發送提醒")
			_, err := bot.PushMessage(user.LineId, linebot.NewTextMessage("記得去填問卷啊！"+surveycakeURL)).Do()
			if err != nil {
				log.Printf("推送提提醒給%s失敗:%v", user.LineId, err)
				continue
			}
			msgCosum, _ := bot.GetMessageConsumption().Do()
			log.Printf("推送提提醒給%s成功, 訊息已使用量: %d", user.LineId,msgCosum.TotalUsage)
			tomorrow := time.Now().AddDate(0, 0, 1)
			setHour, _ := strconv.Atoi(user.SettingRemindTime[0:2])
			setMin, _ := strconv.Atoi(user.SettingRemindTime[3:5])
			_, err = mongo.UpdateRecord(bson.M{"lineid": user.LineId},
				bson.M{"$set": bson.M{"lastremindtime": time.Now().In(timeLoc),
					"nextremindtime": time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
						setHour, setMin, 0, 0, timeLoc)}})
			log.Printf("下次提醒時間: %v", time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), setHour, setMin, 0, 0, timeLoc))
			if err != nil {
				log.Printf("更新提醒時間失敗:%v", err)
			}
		} else {
			log.Printf("%s尚未到提醒時間%s", user.LineId, user.NextRemindTime.In(timeLoc))
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
		rand.Seed(time.Now().UnixNano())
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.LocationMessage:
				log.Printf("[查詢附近藥局] 目前所在: %f, %f", message.Latitude, message.Longitude)
				lastShortDistpharmacy, err := pharmacyInfoUtil.GetLastShortDistancePharmacy(tempPharmacyList, [2]float64{message.Latitude, message.Longitude})
				if err != nil {
					log.Printf("failed to get last short distance pharmacy info: %v",err)
					return
				}
				_, err = bot.ReplyMessage(event.ReplyToken,
					linebot.NewFlexMessage("距離你最近的五家藥局",
						&linebot.CarouselContainer{
							Type: linebot.FlexContainerTypeCarousel,
							Contents: []*linebot.BubbleContainer{
								//第一個旋轉選單
								{
									Type:      linebot.FlexContainerTypeCarousel,
									Size:      linebot.FlexBubbleSizeTypeGiga,
									Direction: linebot.FlexBubbleDirectionTypeLTR,
									Header:    nil,
									Hero: &linebot.ImageComponent{
										Type:  linebot.FlexComponentTypeImage,
										URL:   pharmacyImgURL[0],
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeText,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: lastShortDistpharmacy[0].Properties.Name,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "地址: " + lastShortDistpharmacy[0].Properties.Address,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "電話: " + lastShortDistpharmacy[0].Properties.Phone,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "[剩餘口罩] 大人：" + strconv.Itoa(lastShortDistpharmacy[0].Properties.Mask_adult) +
													"、小孩：" + strconv.Itoa(lastShortDistpharmacy[0].Properties.Mask_child),
											},
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeText,
												Action: linebot.NewURIAction("地圖",
													fmt.Sprintf(MapURL, lastShortDistpharmacy[0].Geometry.Coordinates[0],lastShortDistpharmacy[0].Geometry.Coordinates[1])),
											},
										},
									},
									Footer: nil,
									Styles: &linebot.BubbleStyle{},
								},
								//第二個旋轉選單
								{
									Type:      linebot.FlexContainerTypeCarousel,
									Size:      linebot.FlexBubbleSizeTypeGiga,
									Direction: linebot.FlexBubbleDirectionTypeLTR,
									Header:    nil,
									Hero: &linebot.ImageComponent{
										Type:  linebot.FlexComponentTypeImage,
										URL:   pharmacyImgURL[1],
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeText,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: lastShortDistpharmacy[1].Properties.Name,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "地址:" + lastShortDistpharmacy[1].Properties.Address,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "電話: " + lastShortDistpharmacy[1].Properties.Phone,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "[剩餘口罩] 大人：" + strconv.Itoa(lastShortDistpharmacy[1].Properties.Mask_adult) +
													"、小孩：" + strconv.Itoa(lastShortDistpharmacy[1].Properties.Mask_child),
											},
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeText,
												Action: linebot.NewURIAction("地圖",
													fmt.Sprintf(MapURL, lastShortDistpharmacy[1].Geometry.Coordinates[0],lastShortDistpharmacy[1].Geometry.Coordinates[1])),
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
										URL:   pharmacyImgURL[2],
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeText,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: lastShortDistpharmacy[2].Properties.Name,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "地址: " + lastShortDistpharmacy[2].Properties.Address,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "電話: " + lastShortDistpharmacy[2].Properties.Phone,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "[剩餘口罩] 大人：" + strconv.Itoa(lastShortDistpharmacy[2].Properties.Mask_adult) +
													"、小孩：" + strconv.Itoa(lastShortDistpharmacy[2].Properties.Mask_child),
											},
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeText,
												Action: linebot.NewURIAction("地圖",
													fmt.Sprintf(MapURL, lastShortDistpharmacy[2].Geometry.Coordinates[0],lastShortDistpharmacy[2].Geometry.Coordinates[1])),
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
										URL:   pharmacyImgURL[3],
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeText,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: lastShortDistpharmacy[3].Properties.Name,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "地址: " + lastShortDistpharmacy[3].Properties.Address,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "電話 " + lastShortDistpharmacy[3].Properties.Phone,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "[剩餘口罩] 大人：" + strconv.Itoa(lastShortDistpharmacy[3].Properties.Mask_adult) +
													"、小孩：" + strconv.Itoa(lastShortDistpharmacy[3].Properties.Mask_child),
											},
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeText,
												Action: linebot.NewURIAction("地圖",
													fmt.Sprintf(MapURL, lastShortDistpharmacy[3].Geometry.Coordinates[0],lastShortDistpharmacy[3].Geometry.Coordinates[1])),
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
										URL:   pharmacyImgURL[4],
										Align: linebot.FlexComponentAlignTypeCenter,
										Size:  linebot.FlexImageSizeTypeFull,
									},
									Body: &linebot.BoxComponent{
										Type:   linebot.FlexComponentTypeText,
										Layout: linebot.FlexBoxLayoutTypeVertical,
										Contents: []linebot.FlexComponent{
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: lastShortDistpharmacy[4].Properties.Name,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "地址: " + lastShortDistpharmacy[4].Properties.Address,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "電話: " + lastShortDistpharmacy[4].Properties.Phone,
											},
											&linebot.TextComponent{
												Type: linebot.FlexComponentTypeText,
												Text: "[剩餘口罩] 大人：" + strconv.Itoa(lastShortDistpharmacy[4].Properties.Mask_adult) +
													"、小孩：" + strconv.Itoa(lastShortDistpharmacy[4].Properties.Mask_child),
											},
											&linebot.ButtonComponent{
												Type: linebot.FlexComponentTypeText,
												Action: linebot.NewURIAction("地圖",
													fmt.Sprintf(MapURL, lastShortDistpharmacy[4].Geometry.Coordinates[0],lastShortDistpharmacy[4].Geometry.Coordinates[1])),
											},
										},
									},
									Footer: nil,
									Styles: &linebot.BubbleStyle{},
								},
							},
						})).Do()
				if err != nil {
					log.Printf("failed to send pharmacy info: %v",err)
					return
				}
				msgCosum, _ := bot.GetMessageConsumption().Do()
				log.Printf("推送附近藥局資訊成功，目前已使用訊息量:%d",msgCosum.TotalUsage)
				//bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("被你找到新功能了！ 之後才會支援喔！")).Do()
			case *linebot.TextMessage:
				log.Println(message.Text)
				switch message.Text {
				case "開始查詢":
					log.Println("跳出位置視窗")
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請選擇動作").WithQuickReplies(
						linebot.NewQuickReplyItems(
							linebot.NewQuickReplyButton("https://i.dlpng.com/static/png/6543501_preview.png",
								linebot.NewLocationAction("選擇你所在位置"))))).Do()
					if err != nil {
						log.Printf("error to get quick replay menu:%v", err)
					}
					msgCosum, _ := bot.GetMessageConsumption().Do()
					log.Printf("推送位置視窗成功，目前已使用訊息量:%d",msgCosum.TotalUsage)
				default:
					bot.ReplyMessage(event.ReplyToken,
						linebot.NewFlexMessage("請問你想做什麼?",
							&linebot.CarouselContainer{
								Type: linebot.FlexContainerTypeCarousel,
								Contents: []*linebot.BubbleContainer{
									//第二個旋轉選單
									{
										Type:      linebot.FlexContainerTypeCarousel,
										Size:      linebot.FlexBubbleSizeTypeGiga,
										Direction: linebot.FlexBubbleDirectionTypeLTR,
										Header:    nil,
										Hero: &linebot.ImageComponent{
											Type:  linebot.FlexComponentTypeImage,
											URL:   imageURLPage1[rand.Intn(len(imageURLPage1))],
											Align: linebot.FlexComponentAlignTypeCenter,
											Size:  linebot.FlexImageSizeTypeFull,
										},
										Body: &linebot.BoxComponent{
											Type:   linebot.FlexComponentTypeButton,
											Layout: linebot.FlexBoxLayoutTypeVertical,
											Contents: []linebot.FlexComponent{
												&linebot.TextComponent{
													Type: linebot.FlexComponentTypeText,
													Text: "溫馨提醒：",
												},
												&linebot.TextComponent{
													Type: linebot.FlexComponentTypeText,
													Text: "04/01: 新版‘員工自主健康聲明書'",
												},
												&linebot.ButtonComponent{
													Type: linebot.FlexComponentTypeButton,
													Action: linebot.NewDatetimePickerAction("設定提醒時間", "remindTime", "time",
														"07:00", "23:59", "00:00"),
												},
												&linebot.ButtonComponent{
													Type:   linebot.FlexComponentTypeButton,
													Action: linebot.NewURIAction("防疫問卷", surveycakeURL),
												},
												&linebot.ButtonComponent{
													Type:   linebot.FlexComponentTypeButton,
													Action: linebot.NewURIAction("填寫 '員工自主健康聲明書(ver. 20200319)' ", healthydeclareURL),
												},
												&linebot.ButtonComponent{
													Type:   linebot.FlexComponentTypeButton,
													Action: linebot.NewMessageAction("查詢附近藥局", "開始查詢"),
												},
											},
										},
										Footer: nil,
										Styles: &linebot.BubbleStyle{},
									},
								},
							})).Do()
					msgCosum, _ := bot.GetMessageConsumption().Do()
					log.Printf("推送選單成功，目前已使用訊息量:%d",msgCosum.TotalUsage)
				}
			}
		case linebot.EventTypePostback:
			switch event.Postback.Data {
			case "remindTime": //event.Postback.Params.Time
				tomorrow := time.Now().AddDate(0, 0, 0)
				setHour, _ := strconv.Atoi(event.Postback.Params.Time[0:2])
				setMin, _ := strconv.Atoi(event.Postback.Params.Time[3:5])
				timeLoc, _ := time.LoadLocation("Asia/Shanghai")
				registInfo := Model.User{LineId: event.Source.UserID,
					SettingRemindTime: event.Postback.Params.Time,
					NextRemindTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
						setHour, setMin, 0, 0, timeLoc)}
				log.Println("設定時間:", time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
					setHour, setMin, 0, 0, time.Now().Location()))
				record, err := mongo.InsertOneRecord(registInfo)
				if err != nil {
					bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("設定時間失敗，請重新嘗試!"))
				}
				bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("預定了每天"+event.Postback.Params.Time+"提醒填寫問卷!")).Do()
				msgCosum, _ := bot.GetMessageConsumption().Do()
				log.Printf("推送設定提醒時間成功，目前已使用訊息量:%d",msgCosum.TotalUsage)
				log.Println("Mongo insert info: ", record)
			}
		default:
			bot.PushMessage(event.Source.UserID, linebot.NewTextMessage("DEFAULT")).Do()
			msgCosum, _ := bot.GetMessageConsumption().Do()
			log.Printf("推送Default訊息成功，目前已使用訊息量:%d",msgCosum.TotalUsage)
		}
	}
}
