module example.com/service

go 1.13

require (
	example.com/functions v0.0.0
	github.com/joho/godotenv v1.3.0
)

replace example.com/functions => ./functions
