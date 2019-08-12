package xlsReader

/////////////////////////
// LDAP Parser
/////////////////////////
// BY Patricio Pérez
// p.perez.bustos@accenture.com
/////////////////////////
// 5. LDAP Injection
/////////////////////////

import (
	"fmt"
	"os"
	"log"
	"regexp"
	"gopkg.in/ldap.v2"
)

func getMember(row MACExcelRow) string{
	//uid=<EMAIL>,ou=users,ou=internet,dc=lancargo,dc=com
	var member string
	member = GetParsedLdapMemberDirs()[0]
	//log.Printf("UserGroup: uid=%s,%s\n", row.Email, member)
	return fmt.Sprintf("uid=%s,%s", row.Email, member)
}

func AddAllToLdapGroups() {
	//First, recover the rows which are going to be added.
	var rows []MACExcelRow
	rows = GetParsedRows()
	
	//Now, recover the groups from the configuration.
	var groups []string
	groups = GetParsedLdapOrganizationNames()
	
	//After, connect to the LDAP
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", mainConfig.LdapConfig.Server.Ip, mainConfig.LdapConfig.Port))
	if err != nil {
		log.Fatal(err)
	}
	
	// Access with credentials
	err = l.Bind(mainConfig.LdapConfig.User, mainConfig.LdapConfig.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close() //This will close the connection at the end of the function
	
	// Add a description, and replace the mail attributes
	for _, row := range rows {
		for _, grp := range groups {
			var usr string
			log.Println("======================")
			log.Println("Adding User to Group")
			
			//Get the valid user string for LDAP.
			usr = getMember(row)
			
			//To add the user into the group, it's required to modify
			//the group rules to add a new row with the user info.
			modify := ldap.NewModifyRequest(grp)
			modify.Add("member", []string{usr})
			
			//This will add the user into the corresponding group.
			err = l.Modify(modify)
			if err != nil {
				//If an error is found, it will inform to the user, but
				//it will not close the operation.
				log.Printf("Error trying to add %s into %s", row.Email, grp)
				log.Println(err) //This will not close the connection.
			} else {
				//If no error is found, it will inform it.
				log.Printf("Added %s into %s", row.Email, grp)
			}
			log.Println("======================")
		}
	}
}

func AddToGroups(row MACExcelRow){
	//First, recover the groups from the configuration.
	var groups []string
	groups = GetParsedLdapOrganizationNames()

	//Now, connect to the LDAP
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", mainConfig.LdapConfig.Server.Ip, mainConfig.LdapConfig.Port))
	if err != nil {
		log.Fatal(err)
	}
	
	// Access with credentials
	err = l.Bind(mainConfig.LdapConfig.User, mainConfig.LdapConfig.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Add a description, and replace the mail attributes
	for _, grp := range groups {
		var usr string
		usr = getMember(row)
		fmt.Println(usr)
		_ = grp
		modify := ldap.NewModifyRequest(grp)
		modify.Add("member", []string{usr})

		err = l.Modify(modify)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getCommonName(parsedGroup string) string{
	//Matchs string like: cn=<GROUP>,ou=ebusiness,ou=aplications,ou=internet,dc=lancargo,dc=com
	var re = regexp.MustCompile(`(?m)^cn=(?P<CommonName>\w*),.*`)
	
	//Now, transforms that into a regular search
	n1 := re.SubexpNames()
	r2 := re.FindAllStringSubmatch(parsedGroup, -1)[0]

	md := map[string]string{}
	for i, n := range r2 {
		md[n1[i]] = n
	}
	
	//Return the wanted expression.
	return md["CommonName"]
}

//File Header
func setLDAPFileHeader(f *os.File, orgName string){
	writeLine(f, fmt.Sprintf("dn: %s", orgName))
	writeLine(f, "changetype: modify") //This will skip user prompts for the script
	writeLine(f, "add: member")
}

//This method act as a replacement for the LDAP Injection.
//Since the direct LDAP access is blocked, the function
//is going to generate an LDAP Script file with the group association rules.
//Later, the Jenkins Job will execute the file into the remote server.
func CreateLDAPGroupScriptFiles(filePath string){
	var lstInstructions []string
	var rows []MACExcelRow
	var groups []string
		
	//First, recover the rows which are going to be added.
	rows = GetParsedRows()
	
	//Now, recover the groups from the configuration.
	groups = GetParsedLdapOrganizationNames()
	
	//REDO: Now there is a standalone project which injects the LDAP. So there is only 1 file to inject.
	//Since Jenkins is transfering the StandAlone client, it's required to assign permissions.
	lstInstructions = append(lstInstructions, "chmod +x LDAPGroupAdd")
	
	//Inveting the order for better debugging.
	for _, row := range rows {
		lstInstructions = append(lstInstructions, `echo "===================================="`)
		for _, grp := range groups {
			var usr string
			usr = getMember(row)

			instr := `./LDAPGroupAdd -lsrv '%s' -lprt %d -lusr '%s' -lpwd '%s' -u '%s' -g '%s'`
			instrPrint := `echo "Executing LDAPGroupAdd: -u '%s' -g '%s'"`
			
			lstInstructions = append(lstInstructions, 
				fmt.Sprintf(instrPrint,
					usr,
					grp, //For new line
				), //For new line
			)
			
			lstInstructions = append(lstInstructions, 
				fmt.Sprintf(instr,
					mainConfig.LdapConfig.Server.Ip,
					mainConfig.LdapConfig.Port,
					mainConfig.LdapConfig.User, 
					mainConfig.LdapConfig.Password,
					usr,
					grp, //For new line
				), //For new line
			)
			
			lstInstructions = append(lstInstructions, `echo "------------------------------------"`)
		}
	}
	
	//Add the last instruction (to prevent errors).
	lstInstructions = append(lstInstructions, `echo "All LDAP files executed."`)
	
	//Finally, it's important to create an execution file for all the files.
	f, err := os.Create(fmt.Sprintf("%s/LDAP_EXECUTE.sh", filePath))
	
	if err != nil {
		defer f.Close() //Close the file when the function ends.
		log.Println("Error creating the File")
		log.Fatal(err)
	}
	
	for _, instr := range lstInstructions{
		writeLine(f, instr)
	}
	
	f.Sync()//Save everything.
	f.Close()

}