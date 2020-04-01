package Model

import "time"

type User struct {
	NTAccount         string    `bson:"ntaccount"`
	NextRemindTime    time.Time `bson:"nextremindtime"`
	SettingRemindTime string    `bson:"settingremindtime"`
	LineId            string    `bson:"lineid"`
	ClaimTime         time.Time `bson:"claimtime`
	LastRemindTime    time.Time `bson:"lastremindtime"`
}

type User2 struct {
	LineId            string    `bson:"lineid"`
	LastRemindTime    time.Time `bson:"lastremindtime"`
	SettingRemindTime string    `bson:"settingremindtime"`
	NextRemindTime    time.Time `bson:"nextremindtime"`
}
