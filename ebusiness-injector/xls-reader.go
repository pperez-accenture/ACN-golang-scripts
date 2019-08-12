package main

import (
    "fmt"
	"os"
	"strings"
	"io/ioutil"
    "github.com/tealeg/xlsx"
	//Local library
	xlsReader "./lib" 
)

//This function ignore the empty cells in a row.
func removeEmptyCells (c []*xlsx.Cell) []*xlsx.Cell{
	var selector string
	var cls []*xlsx.Cell
	
	selector = ""
	
	for _,cl := range c {
		if(strings.TrimSpace(cl.Value) != selector){
			cls = append(cls, cl)
		}
	}
	return cls
}

//This function set an Environment Variable.
func WriteIntoFile(file, str string){
	ioutil.WriteFile(file, []byte(str), 0777)
}

func main() {
	var err error
	
	var debug bool
	debug = true
	
	var configPath, langFolder, language string
	configPath = "./config.json"
	langFolder = "./language/"
	language = "br"
	
	var excelFileName string
	excelFileName = "./example.xlsx"
	
	//Read configurations
	xlsReader.SetConfiguration(configPath, langFolder, language)
	err = xlsReader.ReadConfigurations()

	if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
	
	//Testing
	if(debug){
		fmt.Printf("---\nMain Configuration:\n%#v\n", xlsReader.GetMainConfiguration())
		fmt.Printf("---\nExcel Configuration:\n%#v\n", xlsReader.GetExcelConfiguration())
		fmt.Printf("---\nLDAP Organizations:\n%v\n", xlsReader.GetParsedLdapOrganizationNames())
		fmt.Printf("---\nLDAP Members:\n%v\n", xlsReader.GetParsedLdapMemberDirs())
	}
	
	//Read the Excel File
    xlFile, err := xlsx.OpenFile(excelFileName)
	
	//Check if the file is correct.
    if err != nil {
        fmt.Printf("%v\n", err)
    }
	
	var excelConfig = xlsReader.GetExcelConfiguration()
	
	//Processing files
	var isHeader bool 
	isHeader = true
	xlsReader.ExcelInit()
	
    for _, sheet := range xlFile.Sheets {
		//Searching the sheet by the name.
		if(strings.TrimSpace(sheet.Name) == excelConfig.SheetName){
			for _, row := range sheet.Rows {
				//Row verification
				if(len(removeEmptyCells(row.Cells)) > 0){
					//The row has data.
					for colNum, cell := range row.Cells {
						if(isHeader){
							xlsReader.SetColumnPos(cell.Value, colNum)
						} else {
							xlsReader.AddRowData(cell.Value, colNum)
						}
					}
					
					xlsReader.NewRow()
					//We don't need the header anymore
					isHeader = false
				}
			}
		}
    }
	
	//Testing
	if(debug){
		fmt.Printf("---\nExcel Data:\n%#v\n", xlsReader.GetParsedRows());
	}
	//Now that the data is ready, it's time to insert them in the Database and LDAP.
	
	//First it's neccesary to try the insert the missing data into the Database.
	//TO-DO: Workaround for Production Environment. Working in script for the 160 machine.
	sqlFolder := "tmp/sql"
	filePath := fmt.Sprintf("%s%s", sqlFolder, "/EBUSINESS_SQL_INJECT.sql")
	xlsReader.GenerateSQLFile(filePath)
	
	//Since the connection string comes from the config, 
	//the script needs to set the value for the next steps. 
	//So, the script saves the string into a temp file.
	connectStr := fmt.Sprintf("%s%s", sqlFolder, "/SQL_CONNECT_STRING")
	WriteIntoFile(connectStr, xlsReader.GetOracleConnection())
	
	//Then, insert the permissions in LDAP.
	//xlsReader.AddAllToLdapGroups()
	
	ldapFolder := "tmp/ldap"
	xlsReader.CreateLDAPGroupScriptFiles(ldapFolder)
	
	//Since the SSH connection string comes from the config, 
	//the script needs to set the value for the next steps. 
	//So, the script saves the string into a temp file.
	mainConfig := xlsReader.GetMainConfiguration()
	connectSshHost := fmt.Sprintf("%s%s", ldapFolder, "/SSH_HOST")
	connectSshUser := fmt.Sprintf("%s%s", ldapFolder, "/SSH_USER")
	connectSshPass := fmt.Sprintf("%s%s", ldapFolder, "/SSH_PASS")
	WriteIntoFile(connectSshHost, mainConfig.LdapConfig.Server.Ip)
	WriteIntoFile(connectSshUser, mainConfig.LdapConfig.Server.User)
	WriteIntoFile(connectSshPass, mainConfig.LdapConfig.Server.Password)
}

