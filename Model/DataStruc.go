package Model

import "time"

type User struct {
	NTAccount  string    `bson:"ntaccount"`
	RemindTime string    `bson:"remindtime"`
	LineId     string    `bson:"lineid"`
	ClaimTime  time.Time `bson:"claimtime`
}
