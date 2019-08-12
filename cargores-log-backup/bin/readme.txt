cargores-log-fetcher
v0.3 - 2018.06.21
By Patricio Pérez - p.perez.bustos@accenture.com

Usage
With Go: go run findInCargo.go [Options] <AWB>
With linux binaries: ./findInCargo [Options] <AWB>
With windows binaries: findInCargo.exe [Options] <AWB>

Options:
-v 					Debug mode
-d=<DIR>			Set the directory where are located the log files.
					The default folder is the one where the script is located.
-f=<regexp>			Set the filenames to check.
					By default, it will search for all the files matching
					the expression "booking*.log*".
-r=<name>			Set report file name. 
					Parameters: %y year, %M Month (01-12), %d day (01-31), %h hour, %s seconds. 
					Also, %n AWB Number (123-12345678), %p AWB Prefix, %c AWB Code.
					By default, the report file name has the following format: 
					%y-%M-%d_%h-%m-%s_%n-report.txt
<AWB>				The document to search.

Example:

go run findInCargo.go -v -d="./" -f="bookingEv*.log*" -r="report_%y%M%d_%h%m%d_%n.txt" 45-1234567

----------
Backlog
- v0.3 - 2018.06.21: Added support for Group Move Messages.
- v0.2 - 2018.06.21: Added Report filename replacement and flag argument.
- v0.1 - 2018.06.20: Script creation.