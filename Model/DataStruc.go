package Model

import "time"

type User struct {
	NTAccount string `bson:"ntaccount"`
	RemindTime time.Time `bson:"remindtime"`
}