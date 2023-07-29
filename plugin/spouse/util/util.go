package util

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gptbot/plugin/spouse/model"
	"gptbot/store"
	"os"
	"strconv"
)

func GetIndexPath(groupId int64, spouseType model.Type) string {
	return "spouse/" + strconv.FormatInt(groupId, 10) + "/" + spouseType.String() + "/index.json"
}

func GetPicturePath(groupId int64, spouseType model.Type) string {
	return "spouse/" + strconv.FormatInt(groupId, 10) + "/" + spouseType.String() + "/pictures/"
}

func GetWeightPath(groupId int64, spouseType model.Type) string {
	return "spouse/" + strconv.FormatInt(groupId, 10) + "/" + spouseType.String() + "/weight.json"
}

func GetCards(groupId int64, spouseType model.Type) ([]model.Card, error) {
	exist, err := store.GetStoreClient().IsExist(GetIndexPath(groupId, spouseType))
	if err != nil {
		return nil, err
	}
	var cards []model.Card
	if exist {
		buf, err := store.GetStoreClient().GetObjectBytes(GetIndexPath(groupId, spouseType))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(buf, &cards)
		if err != nil {
			return nil, errors.Wrap(err, "marshal index fill error")
		}
	} else {
		cards = []model.Card{}
	}
	return cards, nil
}

func SaveWifeFile(path string, cards []model.Card) error {
	wifeJsonBytes, err := json.Marshal(cards)
	if err != nil {
		return err
	}
	return os.WriteFile(path, wifeJsonBytes, 0644)
}
