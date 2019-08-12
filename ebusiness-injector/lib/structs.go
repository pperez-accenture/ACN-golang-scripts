package xlsReader

/////////////////////////
// LDAP Parser
/////////////////////////
// BY Patricio Pérez
// p.perez.bustos@accenture.com
/////////////////////////
// 1. Type Definitions
/////////////////////////

// 1.1 Excel Configuration Definitions
type ExcelHeader struct{
	OperType 		string `json:"type"`
    Email 			string `json:"email"`
    FirstName 		string `json:"firstName"`
    LastName 		string `json:"lastName"`
    BusinessName 	string `json:"businessName"`
    BusinessId 		string `json:"businessId"`
    BusinessAccount string `json:"businessAccount"`
    PersonCode 		string `json:"personCode"`
    ChargeCode 		string `json:"chargeCode"`
    AgentOrigin 	string `json:"agentOrigin"`
    Functionality 	string `json:"functionality"`
}

type ExcelOperType struct{
	CreateName string `json:"create"`
	UpdateName string `json:"update"`
}

type ExcelConfig struct{
	Language 		string `json:"language"`
	SheetName 		string `json:"sheetData"`
	HeaderNames		ExcelHeader `json:"header"`
	OperationType	ExcelOperType `json:"type"`
}

// 1.2 LDAP Configuration Definitions
type LDAPValidation struct {
	IsInLdap			bool `json:"isInLdap"`
	IsInMac				bool `json:"isInMac"`
	HasMCMPermissions	bool `json:"hasMCMPermissions"`
}

type LDAPValidationConfig struct{
	UserValidation	LDAPValidation `json:"user"`
	GroupValidation	LDAPValidation `json:"group"`
}

type LDAPOrganizationUnit struct{
	UnitName 		string `json:"ou"`
	CommonName		[]string `json:"commonNames"`
	Id				[]string `json:"id"`
	OrgUnit			[]*LDAPOrganizationUnit `json:"organizationalUnit"`
}

type LDAPServer struct{
	Ip			string `json:"ip"`
	User		string `json:"user"`
	Password	string `json:"pass"`
}

type LDAPConfiguration struct{
	Server 			LDAPServer `json:"server"`
	Port 			int `json:"port"`
	Domain 			string `json:"domain"`
	User 			string `json:"user"`
	Password 		string `json:"password"`
	Organization 	string `json:"organization"`
	Validations 	LDAPValidationConfig `json:"validations"`
	OrgUnit 		[]*LDAPOrganizationUnit `json:"organizationalUnit"`
	MemberUnit 		[]*LDAPOrganizationUnit `json:"memberUnit"`
}

// 1.3 Database Configurations
type TableData struct {
	Owner	string `json:"owner"`
	Name	string `json:"name"`
}

type MACRecord struct{
	AccountNumber	TableData `json:"accountNumber"`
	AccountContact	TableData `json:"accountContact"`
	BusinessContact	TableData `json:"businessContact"`
}

type DatabaseConfiguration struct {
	Server 		string `json:"server"`
	Port 		string `json:"port"`
	SID 		string `json:"sid"`
	User 		string `json:"user"`
	Password 	string `json:"password"`
	MACTables 	MACRecord `json:"tables"`
}

type Configuration struct{
	LdapConfig LDAPConfiguration `json:"LDAP"`
	DbConfig DatabaseConfiguration `json:"database"`
}

// 1.3 Library Configuration Definitions
type XLSReaderConfig struct {
	ConfigPath string
	LanguagePath string
}

// 1.4 Column Position
type ExcelColPos map[int]string
// 1.5 Parsing Row Data
type ExcelRowData map[string]string

// 1.5 ExcelRow
type MACExcelRow struct {
	OperType 		string
    Email 			string
    FirstName 		string
    LastName 		string
    BusinessName 	string
    BusinessId 		string
    BusinessAccount string
    PersonCode 		string
    ChargeCode 		string
    AgentOrigin 	string
    Functionality 	string
}