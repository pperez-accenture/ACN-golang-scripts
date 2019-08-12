file-uploader
v0.1 - 2019.02.28
By Patricio Pérez - p.perez.bustos@accenture.com

Usage
With Go: go run file-uploader.go [Options] FILE_1 FILE_2 ... FILE_N
With linux binaries: ./fileUploader [Options] FILE_1 FILE_2 ... FILE_N
With windows binaries: fileUploader.exe [Options] FILE_1 FILE_2 ... FILE_N

Options:
  -H HOST
        Set the remote server to connect. (default "localhost")
  -O UPLOAD_DIR
        Set the path to put the files. (default "/")
  -P PORT
        Set the remote port to connect. (default 22)
  -U USER
        Set the username of the remote server. (default "username")
  -W PASS
        Set the password of the remote server. (default "password")
  -k KEY
        Set the key path. (default "/path/to/rsa/key"). This value is considered, only if you need to work with SSL access.
  -s    Check if you need to work with SSL

Example:

go run sftp.go -H "10.10.10.1" -U "usr" -W "pwd" -O "bea/user_projects/domains/X" XML1.xml XML2.xml

----------
Backlog
- v0.1 - 2019.02.28: Script creation.