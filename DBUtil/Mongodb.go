package DBUtil

import (
	"context"
	"fmt"
	"github.com/weizhe0422/GGQuestionnaireReminder/Model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDB struct {
	URL        string
	Database   string
	Collection string
}

func (m *MongoDB) getMongoDB() (*mongo.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	return mongo.Connect(ctx, options.Client().ApplyURI(m.URL))
}

func (m *MongoDB) InsertOneRecord(userInfo Model.User) (*mongo.InsertOneResult, error) {
	if m.URL == "" {
		return nil, fmt.Errorf("not set mongodb URL")
	}
	dbUtil, _ := m.getMongoDB()
	err := dbUtil.Ping(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongoDB: %v", err)
	}
	defer dbUtil.Disconnect(context.TODO())

	collection := dbUtil.Database(m.Database).Collection(m.Collection)

	timeLoc, _ := time.LoadLocation("Asia/Shanghai")
	if Exist,_, _ := m.FindRecord(userInfo.LineId);Exist {
		//return nil, fmt.Errorf("already registed")
		log.Println("already registed:",userInfo)
		m.UpdateRecord(bson.M{"lineid": userInfo.LineId},
			bson.M{"$set": bson.M{"settingremindtime": userInfo.SettingRemindTime,
				"nextremindtime": userInfo.NextRemindTime,
				"claimtime":      time.Now().In(timeLoc)}})
		return nil, nil
	}

	insertResult, err := collection.InsertOne(context.TODO(), bson.M{
		"ntaccount":         userInfo.NTAccount,
		"settingremindtime": userInfo.SettingRemindTime,
		"nextremindtime":    userInfo.NextRemindTime,
		"lineid":            userInfo.LineId,
		"claimtime":         time.Now().In(timeLoc),
		"lastremindtime":    time.Now().In(timeLoc),
	})
	if err != nil {
		log.Printf("failed to regist NT account: %v", err)
		return nil, err
	}
	log.Printf("ok to register: %v", insertResult.InsertedID)
	return insertResult, nil
}

func (m *MongoDB) FindAllRecord() (*mongo.Cursor, error) {
	dbUtil, _ := m.getMongoDB()
	err := dbUtil.Ping(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongoDB: %v", err)
	}
	defer dbUtil.Disconnect(context.TODO())

	collection := dbUtil.Database(m.Database).Collection(m.Collection)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	findAllResult, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("failed to find: %v", err)
		return nil, fmt.Errorf("failed to find: %v", err)
	}

	return findAllResult, nil
}

func (m *MongoDB) FindRecord(value string) (bool, *Model.User2, error) {
	dbUtil, _ := m.getMongoDB()
	err := dbUtil.Ping(context.TODO(), nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to ping mongoDB: %v", err)
	}
	defer dbUtil.Disconnect(context.TODO())
	collection := dbUtil.Database(m.Database).Collection(m.Collection)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	filter := bson.M{"lineid": value}
	findData := &Model.User2{}
	err = collection.FindOne(ctx, filter).Decode(findData)
	if err != nil {
		log.Printf("failed to find: %v, value: %s", err, value)
		return false, nil, fmt.Errorf("failed to find: %v", err)
	}
	return true, findData, nil
}

func (m *MongoDB) UpdateRecord(filterInfo bson.M, newInfo bson.M) (bool, error) {
	dbUtil, _ := m.getMongoDB()
	err := dbUtil.Ping(context.TODO(), nil)
	if err != nil {
		log.Printf("failed to ping mongoDB: %v", err)
		return false, fmt.Errorf("failed to ping mongoDB: %v", err)
	}
	defer dbUtil.Disconnect(context.TODO())
	collection := dbUtil.Database(m.Database).Collection(m.Collection)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	updateResult, err := collection.UpdateOne(ctx, filterInfo, newInfo)
	if err != nil {
		log.Printf("failed to update: %v", err)
		return false, fmt.Errorf("failed to update: %v", err)
	}
	if updateResult.MatchedCount > 0 && updateResult.ModifiedCount > 0 {
		log.Printf("update ok!")
		return true, nil
	}
	log.Printf("can not found record")
	return false, fmt.Errorf("can not found record")
}
