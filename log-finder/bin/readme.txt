log-finder
v0.1 - 2019.02.14
By Patricio Pérez - p.perez.bustos@accenture.com

Usage
With Go: go run messageLogFinder.go [Options] <AWB>
With linux binaries: ./messageLogFinder [Options] <AWB>
With windows binaries: messageLogFinder.exe [Options] <AWB>

Options:
-v 					Debug mode
-z=<FILE>			Set the zipped file where are located the log files.
					The default zip is named example.zip, located where the script is executed.
-f=<regexp>			Set the filenames to check.
					By default, it will search for all the files matching
					the expression "booking*.log*".
-r=<name>			Set report file name. 
					Parameters: %y year, %M Month (01-12), %d day (01-31), %h hour, %s seconds. 
					Also, %n AWB Number (123-12345678), %p AWB Prefix, %c AWB Code.
					By default, the report file name has the following format: 
					%y-%M-%d_%h-%m-%s_%n-report.txt
-t=<val>			Set the number of parallel workers that are going to verify the files.
					Although more workers means that the process will execute faster, it also means
					that the script will consume more memory, so please take caution.
					The default vale is 10 (tested in computer with 16GB ram).

Example:

go run messageLogFinder.go -v -z="./example.zip" -t=10 -f="bookingEv*.log*" -r="report_%y%M%d_%h%m%d_%n.txt" 45-1234567

----------
Backlog
- v0.1 - 2019.02.14: Script creation.