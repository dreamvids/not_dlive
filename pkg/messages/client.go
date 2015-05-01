package messages

import (
	"fmt"
	"strings"
)

const (
	RankViewer   int = 0
	RankStreamer int = 1
	RankModo     int = 2
	RankAdmin    int = 3
)

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

func (this Session) TableName() string {
	return "users_sessions"
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

type UserChannel struct {
	Id          string `sql:"size:6"`
	Name        string `sql:"size:255"`
	Description string `sql:"size:255"`
	OwnerId     int
	AdminsIds   string `sql:"size:255"`
	Avatar      string
	Background  string `sql:"size:255"`
	Subscribers int
	SubsList    string `sql:"size:255"`
	Views       int
	Verified    bool
}

func (this UserChannel) TableName() string {
	return "users_channels"
}

func GetRank(sessId string) int {
	var sess Session

	Database.Where(&Session{SessionId: sessId}).First(&sess)
	if len(sess.SessionId) <= 0 {
		return RankViewer
	}

	var user User
	userId := sess.UserId
	Database.Where(&User{Id: userId}).First(&user)
	if len(user.Username) <= 0 {
		return RankViewer
	}

	if user.Rank == AdminId {
		return RankAdmin
	}
	if user.Rank == ModoId {
		return RankModo
	}

	var channels []UserChannel
	var channelsIds string
	Database.Where("admins_ids LIKE ?", fmt.Sprintf(";%d;", user.Id)).Find(&channels)
	if len(channels) <= 0 {
		return RankViewer
	}

	for _, c := range channels {
		channelsIds += fmt.Sprintf("%s; ", c.Id)
	}
	channelsIds = strings.TrimSuffix(channelsIds, "; ")

	var accesses []LiveAccess
	Database.Where("channel_id IN (?)", channelsIds).Find(&accesses)
	if len(accesses) >= 0 {
		return RankStreamer
	}

	return RankViewer
}
