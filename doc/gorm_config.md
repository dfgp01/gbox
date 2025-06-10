# gorm.Config 配置项

| 配置项 | 说明 | 默认值 | 代码示例 | 預設 |
|--------|------|--------|----------|-------|
| `SkipDefaultTransaction` | 禁用 GORM 的默认事务机制 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ SkipDefaultTransaction: true, }) ``` | true |
| `NamingStrategy` | 定义表名和字段的命名策略 | 默认命名策略 | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ NamingStrategy: schema.NamingStrategy{ TablePrefix: "tbl_", SingularTable: true, NameReplacer: strings.NewReplacer("CID", "Cid"), }, }) ``` | NamingStrategy{SingularTable: true} |
| `Logger` | 自定义日志记录器 | 默认日志记录器 | ```go customLogger := logger.New( log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{ SlowThreshold: 3 * time.Second, LogLevel: logger.Warn, }, ) db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ Logger: customLogger, }) ``` | 未定 |
| `NowFunc` | 自定义获取当前时间的函数 | `time.Now` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ NowFunc: func() time.Time { return time.Now().UTC() }, }) ``` |  |
| `DryRun` | 生成 SQL 但不执行 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ DryRun: true, }) ``` | |
| `PrepareStmt` | 是否启用预编译语句 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ PrepareStmt: true, }) ``` | true |
| `DisableNestedTransaction` | 禁用嵌套事务支持 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ DisableNestedTransaction: true, }) ``` | true |
| `AllowGlobalUpdate` | 是否允许全局更新操作（不带条件的更新） | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ AllowGlobalUpdate: true, }) ``` | |
| `DisableAutomaticPing` | 禁用自动连接检测 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ DisableAutomaticPing: true, }) ``` | |
| `DisableForeignKeyConstraintWhenMigrating` | 在迁移时禁用外键约束 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ DisableForeignKeyConstraintWhenMigrating: true, }) ``` | true |
| `FullSaveAssociations` | 在保存关联模型时，是否保存所有关联数据 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ FullSaveAssociations: true, }) ``` | |
| `QueryFields` | 是否在查询时包含所有字段 | `false` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ QueryFields: true, }) ``` | |
| `CreateBatchSize` | 批量创建记录时每次提交的记录数 | `0`（不限制） | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ CreateBatchSize: 100, }) ``` | |
| `ClauseBuilders` | 自定义 SQL 子句构建器 | `nil` | ```go db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{ ClauseBuilders: map[string]clause.ClauseBuilder{ "my_clause": myCustomClauseBuilder, }, }) ``` | |

# mysql.Config 配置项

| 配置项 | 说明 | 默认值 | 代码示例 | 預設 |
|--------|------|--------|----------|------|
| `DSN` | 数据源名称（连接字符串） | 无 | ```go mysqlConfig := mysql.Config{ DSN: "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local", } ``` | charset=utf8mb4&parseTime=True&loc=Local |
| `DefaultStringSize` | 默认字符串字段长度 | `255` | ```go mysqlConfig := mysql.Config{ DefaultStringSize: 256, } ``` | |
| `DisableDatetimePrecision` | 禁用 `datetime` 精度 | `false` | ```go mysqlConfig := mysql.Config{ DisableDatetimePrecision: true, } ``` | true |
| `DontSupportRenameIndex` | 重命名索引时采用删除并新建的方式 | `false` | ```go mysqlConfig := mysql.Config{ DontSupportRenameIndex: true, } ``` | true |
| `DontSupportRenameColumn` | 重命名列时采用 `change` 而不是 `rename` | `false` | ```go mysqlConfig := mysql.Config{ DontSupportRenameColumn: true, } ``` |
| `SkipInitializeWithVersion` | 跳过根据 MySQL 版本自动配置的步骤 | `false` | ```go mysqlConfig := mysql.Config{ SkipInitializeWithVersion: true, } ``` | |
| `Conn` | 直接传递一个已有的 `database/sql` 连接 | 无 | ```go sqlDB, err := sql.Open("mysql", dsn) mysqlConfig := mysql.Config{ Conn: sqlDB, } ``` | | 
