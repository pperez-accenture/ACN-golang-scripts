package main

import (
	"fmt"
	"log"
	"flag"
	"gopkg.in/ldap.v2"
)

type LDAPServer struct{
	Ip			string
	Port		int
	User		string
	Password	string
}

type UserGroup struct{
	Group		string
	User		string
}

var debug = false

func AddToGroups(server LDAPServer, data UserGroup){
	if(debug){
		log.Println("Connecting to LDAP");
	}
	//First, connect to LDAP
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		log.Fatal(err)
	}
	
	if(debug){
		log.Println("Validating credentials");
	}
	
	// Access with credentials
	err = l.Bind(server.User, server.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	
	if(debug){
		log.Println("Adding user to group");
	}

	//Add Accout to Group
	modify := ldap.NewModifyRequest(data.Group)
	modify.Add("member", []string{data.User}) //To remove a group, change Add to Remove.

	err = l.Modify(modify)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Added user in group")
		log.Println(fmt.Sprintf("User: %s", data.User))
		log.Println(fmt.Sprintf("Group: %s", data.Group))
	}
}

func main(){
	//A. Input Variables
	//1. Debug
	debugFlag := flag.Bool("v", false, "Set debug")
	
	//2. LDAP Server
	ldapHost := flag.String("lsrv", "", "LDAP Server Host")
	ldapPort := flag.Int("lprt", 389, "LDAP Server Port")
	ldapUser := flag.String("lusr", "", "LDAP Server Username")
	ldapPass := flag.String("lpwd", "", "LDAP Server Password")
	
	//3. User & Group
	usrGroup := flag.String("g", "", "Group Name you want to add the user. Example: cn=INVOICE,ou=BUSINESS,ou=aplications,dc=example,dc=com")
	usrEmail := flag.String("u", "", "LDAP Member you want to add. Example: uid=TEST@MAIL.com,ou=users,dc=example,dc=com")
	
	//B. Reading flag and arguments
	flag.Parse()
	//args := flag.Args()
	
	//C. Passing parameters
	var ldapServer LDAPServer
	var userGroup UserGroup
	
	ldapServer.Ip = *ldapHost
	ldapServer.Port = *ldapPort
	ldapServer.User = *ldapUser
	ldapServer.Password = *ldapPass
	
	userGroup.Group = *usrGroup
	userGroup.User = *usrEmail
	
	debug = *debugFlag
	
	//D. Procedure
	AddToGroups(ldapServer, userGroup)
}