package xlsReader

/////////////////////////
// LDAP Parser
/////////////////////////
// BY Patricio Pérez
// p.perez.bustos@accenture.com
/////////////////////////
// 2. Configuration Functions (Read / Load from json)
/////////////////////////

import (
	"fmt"
	"path"
	"io/ioutil"
	"encoding/json"
)

var config XLSReaderConfig
var mainConfig Configuration

//This function defines the Paths for the Configuration
func SetConfiguration(configFilePath, libraryPath, language string){
	config = XLSReaderConfig{
		ConfigPath: configFilePath,
		LanguagePath: fmt.Sprintf("%s.json",path.Join(libraryPath, language)),
	}
}

//This function defines the XLSReaderConfig
func SetConfigurationObj(xlsConfig XLSReaderConfig){
	config = xlsConfig
}

//This recursive function will navigate through the LDAPOrganizationUnit
//and returns an Array with the LDAP Parsed String Object.
func getParsedLdapOrgName(orgUnits []*LDAPOrganizationUnit) []string{
	var commonNames, res []string
	var prefix string
	
	for _, orgUnit := range orgUnits {
		if(orgUnit.OrgUnit != nil) {
			commonNames = getParsedLdapOrgName(orgUnit.OrgUnit)
			prefix = ""
		} else if(len(orgUnit.CommonName) > 0){
			commonNames = orgUnit.CommonName
			prefix = "cn="
		} else {
			commonNames = []string{""}
			prefix = ""
		}
		
		var unitName string
		unitName = orgUnit.UnitName
		
		for _, name := range commonNames {
			var str string
			if(len(name) > 0){
				str = fmt.Sprintf("%s%s,ou=%s", prefix, name, unitName)
			} else {
				str = fmt.Sprintf("ou=%s", unitName)
			}
			res = append(res, str)
		}
	}
	
	return res
}

//This function adds the Parsed LDAP String with the domain.
func getFullParsedOrganization(orgUnits []*LDAPOrganizationUnit) []string {
	var finalLst []string
	lst := getParsedLdapOrgName(orgUnits)
	
	for _, item := range lst {
		finalLst = append(finalLst, fmt.Sprintf("%s,%s", item, mainConfig.LdapConfig.Domain))
	}
	
	return finalLst
}

//Returns the Parsed LDAP Organization Structure.
func GetParsedLdapOrganizationNames() []string {
	return getFullParsedOrganization(mainConfig.LdapConfig.OrgUnit)
}

//Returns the Parsed LDAP Member Structure.
func GetParsedLdapMemberDirs() []string {
	return getFullParsedOrganization(mainConfig.LdapConfig.MemberUnit)
}

//Returns the Main Config Object.
func GetMainConfiguration() Configuration{
	return mainConfig
}

//Returns the LDAP Config Object.
func GetLdapConfiguration() LDAPConfiguration{
	return mainConfig.LdapConfig
}

//Returns the DB Config Object.
func GetDbConfiguration() DatabaseConfiguration{
	return mainConfig.DbConfig
}

func ReadConfigurations() (error){
	//Read the configuration and the language settings.
	var raw []byte
	var err error
	
	//Reads the Config File
	raw, err = ioutil.ReadFile(config.ConfigPath)
	if err != nil { return err }
	
	err = json.Unmarshal(raw, &mainConfig)
	if err != nil { return err }
	
	//Reads the language file.
	raw, err = ioutil.ReadFile(config.LanguagePath)
	if err != nil { return err }
	
	//You can find the excelConfig var in excel.go
	err = json.Unmarshal(raw, &excelConfig)
	if err != nil { return err }
	
	return nil
}