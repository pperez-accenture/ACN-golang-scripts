package logbackup

//Config can be defined in the XML reading,
//so it's not a constant.
var Config = SetDefaultConfig()

//SetDefaultConfig will initialize the Configuration.
func SetDefaultConfig() LogConfig {
	return LogConfig{
		MaxAttempts:  10,
		IntervalTime: 0.5,
		LogFolder:    "log/",
		ZippedFolder: "./build",
	}
}
