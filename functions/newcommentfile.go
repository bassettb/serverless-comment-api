package functions

type NewCommentFile struct {
	Filename string  `json:"file"`
	Comment  Comment `json:"comment"`
}
