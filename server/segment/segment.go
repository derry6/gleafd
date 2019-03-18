package segment

import "time"

type Segment struct {
	BizTag      string
	MaxID       int64
	Step        int32
	Description string
	Updated     time.Time
}
