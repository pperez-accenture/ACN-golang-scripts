package logbackup

import "path/filepath"

//JSON Data

//LogConfig Contains the process configuration.
type LogConfig struct {
	MaxAttempts  int     `json:"max-attempts"`  //MaxAttempts set the default max attempts per ftp reading
	IntervalTime float64 `json:"interval-time"` //IntervalTime set the default time to wait (TTW) between the calls.
	LogFolder    string  `json:"log-folder"`    //LogFolder set the default folder where the logs are going to be dropped.
	ZippedFolder string  `json:"zipped-folder"` //ZippedFolder set the default folder where the final zip is going to be left.
}

//LogServer Contain the Server connection settings.
type LogServer struct {
	Host     string   `json:"host"`
	User     string   `json:"user"`
	Pass     string   `json:"pass"`
	Paths    []string `json:"paths"`
	SavePath string   `json:"save-in"`
}

//LogServers Contain the Server collections to fetch.
type LogServers struct {
	Servers []LogServer `json:"servers"`
}

//LogBackup Set the environments to fetch.
type LogBackup struct {
	BookingLog    LogServers `json:"system1"`
	CargoResLog   LogServers `json:"system2"`
	CargoPriceLog LogServers `json:"system3"`
	Config        LogConfig  `json:"config"`
}

//GetSavePath will return the composed saving path.
func (ls *LogServer) GetSavePath() string {
	return filepath.Join(Config.LogFolder, ls.SavePath)
}

//GetServerList return the list of existing servers.
func (f *LogBackup) GetServerList() []LogServer {
	var lstServers []LogServer //Server list to handle

	//Pass all the collections into one only variable
	lstServers = append(lstServers, f.BookingLog.Servers...)
	lstServers = append(lstServers, f.CargoResLog.Servers...)
	lstServers = append(lstServers, f.CargoPriceLog.Servers...)

	return lstServers
}

//ZipSavePath return the path to save the named file
func (f *LogBackup) ZipSavePath(filename string) string {
	return filepath.Join(Config.ZippedFolder, filename)
}
