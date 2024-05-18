package tvsocket

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"
)

var NewYork *time.Location

func DateTimeStr(dt time.Time) string {
	if NewYork == nil {
		NewYork, _ = time.LoadLocation("America/New_York")
	}
	t := dt.In(NewYork).Format(time.RFC3339)[0:19]
	t = strings.ReplaceAll(t, "T", " ")
	return t
}

// GetRandomString ...
func GetRandomString(length int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	var characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var letterIdxBits int = 6
	var letterIdxMask int64 = 1<<letterIdxBits - 1
	var letterIdxMax = 63 / letterIdxBits

	requestID := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(characters) {
			requestID[i] = characters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(requestID)
}

// GetStringRepresentation ...
func GetStringRepresentation(data interface{}) string {
	str, _ := json.Marshal(data)
	return string(str)
}
