package dto

import "time"

type Genre uint8

const (
	PROGRAMMING Genre = iota
	DHBW
	OTHER
)

type MehmDTO struct {
	Id          int       `json:"id"`
	AuthorName  string    `json:"authorName"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageSource string    `json:"imageSource"`
	CreatedDate time.Time `json:"createdDate"`
	Genre       Genre     `json:"genre"`
	Likes       int       `json:"likes"`
}

type CommentDTO struct {
	Comment  string    `json:"id"`
	Author   string    `json:"author"`
	DateTime time.Time `json:"dateTime"`
}

type CommentInput struct {
	Id      int64  `json:"id" minimum:"1"`
	Comment string `json:"text" minlength:"1" maxlength:"256"`
}

type Comment struct {
	MehmId  int64  `json:"mehmId" min:"1"`
	Comment string `json:"comment" minlength:"1" maxlength:"256"`
}

type MehmInput struct {
	Description string `json:"description" minlength:"1" maxlength:"128"`
	Title       string `json:"title" minlength:"1" maxlength:"32"`
}
