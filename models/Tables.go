package models

//Tables 要用户信息表
type Tables struct {
	TableSchema  string            `gorm:"column:TABLE_SCHEMA;type:varchar(64);"`   //数据库名
	TableName    string            `gorm:"column:TABLE_NAME;type:varchar(64);"`     //表名
	TableComment string            `gorm:"column:TABLE_COMMENT;type:varchar(256);"` //表描述
	PackageName  string            //生成模板文件的包名
	GoName       string            //要生成的struct名
	TableColumns []Columns         //表下面的所有列名
	Imports      map[string]string //要导入的包
}
