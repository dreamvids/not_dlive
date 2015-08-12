package stream

type LiveAccess struct {
	Id         int
	ChannelId  string `sql:"size:6"`
	UserId     int
	Key        string `sql:"size:255"`
	Timestamp  int64
	Online     bool
	StreamName string `sql:"size:255"`
	Viewers    int
}

type Session struct {
	Id         int
	UserId     int
	SessionId  string `sql:"size:32"`
	Expiration int64
	Remember   bool
}

type User struct {
	Id            int
	Username      string `sql:"size:40"`
	Email         string `sql:"size:255"`
	Pass          string `sql:"size:100"`
	Subscriptions string `sql:"size:255"`
	RegTimestamp  int64
	RegIp         string `sql:"size:15"`
	ActualIp      string `sql:"size:15"`
	Rank          int
	Settings      string `sql:"size:255"`
	LastVisit     int64
	LogFail       string `sql:"size:255"`
}
