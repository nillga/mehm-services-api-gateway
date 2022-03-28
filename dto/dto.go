package dto

type CommentInput struct {
	UserId  int64 `json:"userId"`
	Id      int64  `json:"id"`
	Comment string `json:"comment"`
	Admin   bool   `json:"isAdmin"`
}

type MehmInput struct {
	UserId      int64  `json:"userId"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Admin       bool   `json:"isAdmin"`
}