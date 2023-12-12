@REM go install github.com/joho/godotenv
@REM go install  "github.com/kataras/iris/v12"
kill

start sass --style=compressed --error-css --update --watch ../public/assets/css/scss/input.scss:./public/assets/css/scss.css

start ./tailwindcss --watch -i ./public/assets/css/tailwind/input.css -o ./public/assets/css/tailwind.css --minify

run
