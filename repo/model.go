package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/btschwartz12/forum/repo/db"
)

var (
	EstTimezone *time.Location
)

func init() {
	var err error
	EstTimezone, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(fmt.Errorf("failed to load timezone: %w", err))
	}
}

type EstTime struct {
	time.Time
}

func (t EstTime) String() string {
	return fmt.Sprintf("%s EST", t.In(EstTimezone).Format("2006-01-02 15:04:05"))
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type Post struct {
	ID            int64
	Author        User
	Content       string
	Timestamp     EstTime
	Ip            string
	ImageFilename string
}

func zulu(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (p *Post) fromDb(row *db.GetAllPostsRow) {
	p.ID = row.PostID
	p.Author = User{
		ID:       row.UserID,
		Username: row.UserUsername,
		Password: row.UserPassword,
		IsAdmin:  row.UserIsAdmin == 1,
	}
	p.Content = row.PostContent
	t, _ := time.Parse(time.RFC3339, row.PostTimestamp)
	p.Timestamp = EstTime{t}
	p.Ip = row.PostIp
	p.ImageFilename = row.PostImageFilename.String
}

func (p *Post) toDb() db.InsertPostParams {
	return db.InsertPostParams{
		AuthorID:      p.Author.ID,
		Content:       p.Content,
		Timestamp:     zulu(p.Timestamp.Time),
		Ip:            p.Ip,
		ImageFilename: sql.NullString{String: p.ImageFilename, Valid: p.ImageFilename != ""},
	}
}

func (u *User) toDb() db.InsertUserParams {
	return db.InsertUserParams{
		Username: u.Username,
		Password: u.Password,
		IsAdmin:  boolToInt(u.IsAdmin),
	}
}

func (u *User) fromDb(row *db.User) {
	u.ID = row.ID
	u.Username = row.Username
	u.Password = row.Password
	u.IsAdmin = row.IsAdmin == 1
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
