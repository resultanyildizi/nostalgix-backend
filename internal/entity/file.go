package entity

import "fmt"

type File struct {
	ID  string `json:"id"`
	URL string `json:"url"`

	// These variables are used internally
	UserID      string `json:"-"`
	Subject     string `json:"-"`
	ContentType string `json:"-"`
	Size        int64  `json:"-"`
}

func (f File) GetExtension() string {
	switch f.ContentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	default:
		return ""
	}
}
func (f File) GetName() string {
	return fmt.Sprintf("%s%s", f.ID, f.GetExtension())
}
