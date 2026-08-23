package clock

import "time"

type Wall struct{}
func NewWall() Wall { return Wall{} }
func (Wall) Now() time.Time { return time.Now() }
func (Wall) After(d time.Duration) <-chan time.Time { return time.After(d) }
