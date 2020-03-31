package Model

import "time"

type User struct {
	NTAccount      string    `bson:"ntaccount"`
	RemindTime     string    `bson:"remindtime"`
	LineId         string    `bson:"lineid"`
	ClaimTime      time.Time `bson:"claimtime`
	LastRemindTime time.Time    `bson:"lastremindtime"`
}

type User2 struct {
	LineId         string `bson:"lineid"`
	RemindTime     string `bson:"remindtime"`
	LastRemindTime time.Time `bson:"lastremindtime"`
}
