go 1.18

module gbox

require (
	google.golang.org/protobuf v1.34.1
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.10
)

require (
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
)

replace github.com/protocolbuffers/protobuf-go => google.golang.org/protobuf v1.34.1
