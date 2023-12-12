go build -o ./build/main.exe .
taskkill -im main.exe -f

powershell ./build/main.exe | tee out.txt