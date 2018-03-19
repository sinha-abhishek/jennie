package confighelper

import (
	"encoding/json"
	"log"
	"os"
)

type AutoResponse struct {
	LinkedinResponse string `json:"linkedin"`
}

var auto *AutoResponse

func GetAutoResponseConfig() (*AutoResponse, error) {
	if auto == nil {
		file, err := os.Open("./conf/autoreply.json")
		if err != nil {
			return auto, err
		}
		auto = new(AutoResponse)
		err = json.NewDecoder(file).Decode(auto)
		if err != nil {
			log.Println("Failed parsing ", err)
			return auto, err
		}
	}
	return auto, nil
}
