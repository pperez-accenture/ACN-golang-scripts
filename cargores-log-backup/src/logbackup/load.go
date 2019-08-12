package logbackup

import (
	"encoding/json"
	"io/ioutil"
)

//ReadConfig reads the JSON file and put it into LogFetching
func ReadConfig(path string) (LogBackup, error) {
	var fetch LogBackup

	//Read the configuration and the language settings.
	var raw []byte
	var err error

	//Reads the Config File
	raw, err = ioutil.ReadFile(path)
	if err != nil {
		return fetch, err
	}

	err = json.Unmarshal(raw, &fetch)
	if err != nil {
		return fetch, err
	}

	initConfig(fetch.Config)

	return fetch, nil
}

func initConfig(c LogConfig) {
	var emptyConfig LogConfig

	//Check if config has some data
	if c != emptyConfig {
		if c.IntervalTime > 0 {
			Config.IntervalTime = c.IntervalTime
		}
		if c.MaxAttempts > 0 {
			Config.MaxAttempts = c.MaxAttempts
		}
		if c.LogFolder != "" {
			Config.LogFolder = c.LogFolder
		}
		if c.ZippedFolder != "" {
			Config.ZippedFolder = c.ZippedFolder
		}
	}
}
