package xlsReader

/////////////////////////
// LDAP Parser
/////////////////////////
// BY Patricio Pérez
// p.perez.bustos@accenture.com
/////////////////////////
// 4. Database File Generation
/////////////////////////

import (
	"fmt"
	"os"
	"log"
	"strings"
)

var dbConf DatabaseConfiguration

//This function will set the Connection String
func GetOracleConnection() string{ 
	return fmt.Sprintf("%s/%s@%s:%s/%s", dbConf.User, dbConf.Password, dbConf.Server, dbConf.Port, dbConf.SID) 
}

func replaceQuote(str string) string{
	return strings.Replace(str, "'", "''", -1);
}

func removeQuote(str string) string{
	return strings.Replace(str, "'", "", -1);
}

//This function creates the insert for the AccountNumber Struct
func setAccountNumber(row MACExcelRow) string{
	return fmt.Sprintf("INSERT INTO %s.%s (ACCOUNT_NAME, ACCOUNT_NUMBER, ACCOUNT_COUNTRY_ID) VALUES (%s, '%s', %s);", 
			dbConf.MACTables.AccountNumber.Owner, 
			dbConf.MACTables.AccountNumber.Name, 
			row.BusinessAccount, 
			replaceQuote(row.BusinessName), 
			removeQuote(row.PersonCode))
}

//This function creates the insert for the AccountContact Struct
func setAccountContact(row MACExcelRow) string{
	return fmt.Sprintf("INSERT INTO %s.%s (ACCOUNT_NAME, ACCOUNT_EMAIL) VALUES (%s, '%s');", 
			dbConf.MACTables.AccountContact.Owner, 
			dbConf.MACTables.AccountContact.Name, 
			removeQuote(row.BusinessAccount), 
			removeQuote(row.Email))
}

//This function creates the insert for the BusinessContact Struct
func setBusinessContact(row MACExcelRow) string{
	return fmt.Sprintf("INSERT INTO %s.%s (ACCOUNT_AIR_STATION, ACCOUNT_ENTERPRISE_ID, ACCOUNT_EMAIL) VALUES ('%s', %s, '%s');", 
			dbConf.MACTables.BusinessContact.Owner, 
			dbConf.MACTables.BusinessContact.Name, 
			removeQuote(row.AgentOrigin), 
			removeQuote(row.ChargeCode), 
			removeQuote(row.Email))	
}

//The new function reads the array and prepare a SQL File.
//Later, the job will upload the file in NEXUS and is going 
//to call SQLDeveloper to inject the file.
func GenerateSQLFile(filePath string){
	//Reads the configuration
	dbConf = GetDbConfiguration() //From config.go
	
	//Getting the data to inject.
	var macRows []MACExcelRow
	macRows = GetParsedRows()
	
	//Create new File
	f, err := os.Create(filePath)
	
	if err != nil {
		log.Println("Error creating the File")
		log.Fatal(err)
	}
	
	defer f.Close() //Close the file when the function ends.
	
	//File Header
	writeLine(f, "-----------------------------------------------")
	writeLine(f, "SET DEFINE OFF;") //This will skip user prompts for the script
	writeLine(f, "-----------------------------------------------")
	
	//File Body
	for _,row := range macRows {
		//Insert the data
		writeLine(f, "-----------------------------------------------")
		writeLine(f, fmt.Sprintf("-- Adding data from user: %s, account: %s", row.Email, row.BusinessAccount))
		writeLine(f, "-----------------------------------------------")
		writeLine(f, setAccountNumber(row))
		writeLine(f, setAccountContact(row))
		writeLine(f, setBusinessContact(row))
		
	}

	//File Footer
	writeLine(f, "-----------------------------------------------")
	writeLine(f, "SET DEFINE ON;") //This will reenable user prompts for the script
	writeLine(f, "-----------------------------------------------")
	writeLine(f, "COMMIT;") //This will save the data of the script.
	writeLine(f, "-----------------------------------------------")
	writeLine(f, "EXIT;") //This will end the execution.
	
	f.Sync()//Save everything.
}