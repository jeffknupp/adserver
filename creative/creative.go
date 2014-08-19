package creative

import "time"

type CreativeId int

type Creative struct {
	Id            CreativeId
	Height, Width int
	Path          string
}

type CreativeStat struct {
	Impressions   int
	LastServed    time.Time
	ServedPerHour map[int]int
}

func NewCreativeStat() *CreativeStat {
	return &CreativeStat{0, time.Now(), make(map[int]int, 24)}
}

var universalCreativeId CreativeId = 0

func getId() CreativeId {
	universalCreativeId += 1
	return universalCreativeId
}

func NewCreative(path string, width, height int) *Creative {
	return &Creative{Id: getId(), Height: height, Width: width, Path: path}
}
