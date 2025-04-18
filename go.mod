go 1.21

toolchain go1.22.5

module gbox

require (
	go.uber.org/zap v1.27.0
	gorm.io/driver/mysql v1.5.4
	gorm.io/gorm v1.25.7
)

require (
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/lestrrat-go/strftime v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	google.golang.org/protobuf v1.34.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible

require (
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
)

replace github.com/protocolbuffers/protobuf-go => google.golang.org/protobuf v1.34.1
