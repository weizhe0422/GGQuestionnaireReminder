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
	URL string
	Database string
	Collection string
}

func (m *MongoDB) getMongoDB() (*mongo.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	return mongo.Connect(ctx, options.Client().ApplyURI(m.URL))
}

func (m *MongoDB) InsertOneRecord(userInfo Model.User) (*mongo.InsertOneResult, error){
	if m.URL  == "" {
		return nil, fmt.Errorf("not set mongodb URL")
	}
	dbUtil, _ := m.getMongoDB()
	err := dbUtil.Ping(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongoDB: %v", err)
	}

	collection := dbUtil.Database(m.Database).Collection(m.Collection)

	if notExist, _ := m.FindRecord(userInfo.NTAccount); notExist {
		return nil, fmt.Errorf("already registed")
	}

	insertResult, err := collection.InsertOne(context.TODO(), bson.M{
		"ntaccount": userInfo.NTAccount,
		"remindtime": userInfo.RemindTime,
	})
	if err != nil {
		log.Printf("failed to regist NT account: %v",err)
		return nil, err
	}
	log.Printf("ok to register: %v", insertResult.InsertedID)
	return insertResult, nil
}

func (m *MongoDB) FindRecord(value string) (bool, error) {
	dbUtil, _ := m.getMongoDB()
	err := dbUtil.Ping(context.TODO(), nil)
	if err != nil {
		return false, fmt.Errorf("failed to ping mongoDB: %v", err)
	}
	collection := dbUtil.Database(m.Database).Collection(m.Collection)
	ctx, _ := context.WithTimeout(context.Background(), 30 *time.Second)
	filter := bson.M{"ntaccount": value}
	err = collection.FindOne(ctx, filter).Decode(&Model.User{})
	if err != nil {
		return false, fmt.Errorf("failed to find: %v", err)
	}
	return true, nil
}