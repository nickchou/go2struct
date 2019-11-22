package models

//DbInfo 数据库信息，含表明和列名
type DbInfo struct {
	DbTables []Tables //表信息 + 列信息
	//DbColumns []Columns //列名信息
}
