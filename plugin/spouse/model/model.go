package model

type Card struct {
	Name         string
	Source       string
	UploaderName string
	UploaderId   int64
	GroupId      int64
	Hash         string
}

type Type string

const (
	Wife    Type = "老婆"
	Husband Type = "老公"
)

func (g *Type) String() string {
	switch *g {
	case Wife:
		return "老婆"
	case Husband:
		return "老公"
	default:
		return ""
	}
}
