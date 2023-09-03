package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"gptbot/plugin/spouse/model"
	"os"
)

const (
	imagePath = ""
	indexPath = ""
)

func main() {
	//fileInfos, err := os.ReadDir(indexPath)
	//if err != nil {
	//	panic(err)
	//}
	buf, err := os.ReadFile(indexPath)
	if err != nil {
		panic(err)
	}
	cards := []model.Card{}
	err = json.Unmarshal(buf, &cards)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(cards); i++ {
		if cards[i].Hash != "" {
			continue
		}
		buf, err := os.ReadFile(imagePath + cards[i].Name + ".jpg")
		if err != nil {
			println(cards[i].Name)
			continue
		}
		sum := md5.Sum(buf)
		cards[i].Hash = hex.EncodeToString(sum[:])
		cards[i].UploaderName = "system"
		cards[i].UploaderId = 0
		os.Rename(imagePath+cards[i].Name+".jpg", imagePath+"1/"+cards[i].Hash+".jpg")
	}
	buf, err = json.MarshalIndent(cards, "", "  ")
	os.WriteFile(indexPath, buf, 0644)
}
