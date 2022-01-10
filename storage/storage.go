package storage

import "context"

var linkMap map[string]string

func Initialize() {
	linkMap = make(map[string]string)
}

func AddRecord(key string, data string, ctx context.Context) {
	linkMap[key] = data
}

func FindRecord(key string, ctx context.Context) (res string) {
	value, found := linkMap[key]
	if found {
		return value
	} else {
		return ""
	}
}
