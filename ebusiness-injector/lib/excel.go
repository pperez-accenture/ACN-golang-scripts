package xlsReader

/////////////////////////
// LDAP Parser
/////////////////////////
// BY Patricio Pérez
// p.perez.bustos@accenture.com
/////////////////////////
// 3. Excel functions
/////////////////////////

import (
	"strings"
)

var excelConfig ExcelConfig
var excelColMap ExcelColPos
var excelRow 	ExcelRowData
var excelRows 	[]ExcelRowData

func GetExcelConfiguration() ExcelConfig{
	return excelConfig
}

func equalStringNoCase(str1, str2 string) bool{
	return strings.EqualFold(strings.TrimSpace(str1), strings.TrimSpace(str2))
}

func IsValidColumn(col string) bool{
	var eh ExcelHeader
	eh = excelConfig.HeaderNames
	
	//This validates if the cell is from a valid column.
	switch {
		case equalStringNoCase(col, eh.OperType),
				equalStringNoCase(col, eh.Email),
				equalStringNoCase(col, eh.FirstName),
				equalStringNoCase(col, eh.LastName),
				equalStringNoCase(col, eh.BusinessName),
				equalStringNoCase(col, eh.BusinessId),
				equalStringNoCase(col, eh.BusinessAccount),
				equalStringNoCase(col, eh.PersonCode),
				equalStringNoCase(col, eh.ChargeCode),
				equalStringNoCase(col, eh.AgentOrigin),
				equalStringNoCase(col, eh.Functionality):
			return true
	}
	
	return false
}

func ExcelInit(){
	excelColMap = make(ExcelColPos)
	excelRows = make([]ExcelRowData, 0)
}

func NewRow(){
	//Create new Row to add data.
	if(len(excelRow) > 0) {
		//If the row has data, before to continue
		//is required to add the existing row to the array.
		excelRows = append(excelRows, excelRow)
	}
	excelRow = make(ExcelRowData)
}

func SetColumnPos(col string, pos int){
	//Removing spaces
	col = strings.TrimSpace(col)
	if(IsValidColumn(col)){
		excelColMap[pos] = col
	}
}

//CRITICAL: This function will add the cell value.
func AddRowData(data string, pos int){
	var col string
	col = excelColMap[pos]
	excelRow[col] = strings.TrimSpace(data)
}

func getParsedRow(row ExcelRowData) MACExcelRow{
	var ty, em, fn, ln, bn, bi, ba, pc, cc, ao, fu string
	
	var eh ExcelHeader
	eh = excelConfig.HeaderNames

	//The function will verify which field ts the cell and 
	//add the value from the column into the corresponding field.
	for col, val := range row {
		switch {
			case equalStringNoCase(col, eh.OperType):
				ty = val
			case equalStringNoCase(col, eh.Email):
				//Since this is critital, if the user email has a comma for a mistake, it will be replaced with a dot.
				em = strings.Replace(val, ",", ".", -1)
			case equalStringNoCase(col, eh.FirstName):
				fn = val
			case equalStringNoCase(col, eh.LastName):
				ln = val
			case equalStringNoCase(col, eh.BusinessName):
				bn = val
			case equalStringNoCase(col, eh.BusinessId):
				bi = val
			case equalStringNoCase(col, eh.BusinessAccount):
				ba = val
			case equalStringNoCase(col, eh.PersonCode):
				pc = val
			case equalStringNoCase(col, eh.ChargeCode):
				cc = val
			case equalStringNoCase(col, eh.AgentOrigin):
				ao = val
			case equalStringNoCase(col, eh.Functionality):
				fu = val
		}
	}
	
	return MACExcelRow{ty, em, fn, ln, bn, bi, ba, pc, cc, ao, fu}
}

func GetParsedRows() []MACExcelRow{
	var parsedRows []MACExcelRow
	
	for _,row := range excelRows {
		parsedRows = append(parsedRows, getParsedRow(row))
	}
	
	return parsedRows
}