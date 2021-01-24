package main

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/kardianos/osext"
)

//Configuration 配置信息
type Configuration struct {
	Connection string `json:"Connection"`
	UserName   string `json:"UserName"`
	Password   string `json:"Password"`
}

//Save 写到文件
func (config *Configuration) Save() error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	folder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(folder, "config.json"), data, 0666)
}

//Load 读取文件
func (config *Configuration) Load() error {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(path.Join(folder, "config.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

//Config 通用配置
var Config = &Configuration{
	Connection: "conn",
	UserName:   "username",
	Password:   "password",
}
