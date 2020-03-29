package Model

type User struct {
	NTAccount  string    `bson:"ntaccount"`
	RemindTime string `bson:"remindtime"`
	LineId     string    `bson:"lineid"`
}
