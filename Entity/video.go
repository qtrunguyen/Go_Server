package entity

type Video struct {
	ID          string `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" binding:"min=2,max=10"`
	Description string `json:"description" binding:"max=20"`
	URL         string `json:"url" binding:"required,url"`
}
