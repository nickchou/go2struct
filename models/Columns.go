package models

//Columns 列信息
type Columns struct {
	TableSchema     string `gorm:"column:TABLE_SCHEMA;"`     //数据库名
	TableName       string `gorm:"column:TABLE_NAME;"`       //表名
	ColumnName      string `gorm:"column:COLUMN_NAME;"`      //列名
	ColumnKey       string `gorm:"column:COLUMN_KEY;"`       //主键，PRI=主键
	OrdinalPosition string `gorm:"column:ORDINAL_POSITION;"` //列顺序，升序
	ColumnDefault   string `gorm:"column:COLUMN_DEFAULT;"`   //列默认值
	IsNullable      string `gorm:"column:IS_NULLABLE;"`      //是否允许为弄 NO/YES
	DataType        string `gorm:"column:DATA_TYPE;"`        //列数据类型
	ColumnType      string `gorm:"column:COLUMN_TYPE;"`      //列数据类型，含长度
	Extra           string `gorm:"column:EXTRA;"`            //是否自增 auto_increment=自增
	ColumnComment   string `gorm:"column:COLUMN_COMMENT;"`   //列注释
	GoName          string //struct的属性名
	GoType          string //数据库类型对应的GO语言类型
}
