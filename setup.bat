go get -u github.com/joho/godotenv
go get -u  "github.com/kataras/iris/v12"

go install github.com/joho/godotenv
go install  "github.com/kataras/iris/v12"

go mod tidy
go get

mkdir ./build
mkdir ./logs/daylog
