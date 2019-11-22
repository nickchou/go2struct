package main

import (
	"bytes"
	"fmt"
	"go2struct/models"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

//Dbo 数据库链接
var Dbo *gorm.DB

func main() {
	tplPath := "./template/gorm.tpl"
	filePath := "./tmodels"
	dbInfo := &models.DbInfo{}
	//初始化数据库连接
	initDB()
	//获取数据库所有的表和列，并初始化好命名空间
	tables := getDataBaseInfo("models")
	dbInfo.DbTables = tables
	//dbInfo.DbColumns = columns
	//fmt.Println(dbInfo.DbTables)
	//fmt.Println("==================================")
	//fmt.Println(dbInfo.DbColumns)

	//读取模板生成模板
	tplBytes, err := ioutil.ReadFile(tplPath)
	if err != nil {
		fmt.Println("读取模板文件错误:", err)
		return
	}
	//模板文件字符串
	tplStr := string(tplBytes)
	//fmt.Println(tplStr)

	//new 模板，命名temHtml
	tplFile, _ := template.New("temHtml").Parse(tplStr)
	//创建多级目录
	os.MkdirAll(filePath, os.ModePerm)
	curPath, _ := os.Getwd()
	fmt.Println(curPath)
	//获取要生成模板的绝对路径
	genDir, _ := filepath.Abs(filePath)
	fmt.Println(genDir)
	genDir = strings.Replace(genDir, "\\", "/", -1)
	//命名空间
	model := path.Base(genDir)
	fmt.Println(model)
	//遍历表
	for i, table := range dbInfo.DbTables {
		if i > 0 {
			break
		}
		fmt.Println("tablename:", table.TableName)
		fmt.Println("tablename:", table.PackageName)
		fmt.Println("==========================end==========================")
		//根据模板生成最终数据
		buf := new(bytes.Buffer)
		tplFile.Execute(buf, table)
		//fmt.Println(buf.String())
		//生成go文件路径
		goFilePath := fmt.Sprintf("%s//%s%s", filePath, table.TableName, ".go")
		fmt.Println(goFilePath)
		f, err := os.Create(filePath + table.TableName + ".go")
		defer f.Close()
		if err == nil {
			//保存文件
			f.Write([]byte(buf.String()))
		}
	}
}
func saveGoFile() {

}

//initdb 初始化数据库连接池
func initDB() {
	var err error
	Dbo, err = gorm.Open("mysql", "root:123456@(127.0.0.1:3306)/scriptdb?charset=utf8")
	if err != nil {
		panic(err)
	} else {
		Dbo.LogMode(true)
		//设置最大连接数（默认=2），如果<=0不保留空闲连接数
		Dbo.DB().SetMaxIdleConns(1)
		//设置最大的数据库连接数，默认值=0（没有限制）
		Dbo.DB().SetMaxOpenConns(1)
	}
}

//getDataBaseInfo 获取数据库的表、列名等
func getDataBaseInfo(spaceName string) []models.Tables {
	//查所有表信息
	var tables []models.Tables
	Dbo.Raw("SELECT TABLE_SCHEMA,TABLE_NAME,TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA=? ORDER BY TABLE_NAME ASC;", "scriptdb").Scan(&tables)
	//查所有列信息
	var columns []models.Columns
	Dbo.Raw("SELECT TABLE_SCHEMA,TABLE_NAME,COLUMN_NAME,COLUMN_KEY,ORDINAL_POSITION,COLUMN_DEFAULT,IS_NULLABLE,DATA_TYPE,COLUMN_TYPE,EXTRA,COLUMN_COMMENT FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME,ORDINAL_POSITION ASC;", "scriptdb").Scan(&columns)
	//循环列信息，映射golang数据类型
	for j := range columns {
		columns[j].GoName = TitleCasedName(columns[j].ColumnName) //格式化列名
		columns[j].GoType = objcTypeStr(columns[j].DataType)      //数据库类型转golang类型
	}
	var k int = 0
	//循环表信息，设置命名空间
	for i := range tables {
		tables[i].PackageName = spaceName
		tables[i].GoName = TitleCasedName(tables[i].TableName) //格式化表名
		tables[i].Imports = make(map[string]string)            //初始化当前表的包
		//这里减少了重复循环，前提是表明、列名里的TABLE_NAME排过序的，否则会有问题
		for ; k < len(columns); k++ {
			if tables[i].TableName == columns[k].TableName {
				//把列名添加到表的数据结构里
				tables[i].TableColumns = append(tables[i].TableColumns, columns[k])
				//====下面是添加包的逻辑
				if _, ok := tables[i].Imports["time"]; !ok && columns[k].GoType == "time.Time" {
					//判断表中的Packages是否已经存在time，不存在就添加time包
					tables[i].Imports["time"] = "time"
				}
			} else {
				break
			}
		}
	}
	return tables
}
func objcTypeStr(columnType string) string {
	goType := ""
	switch columnType {
	case "bit", "tinyint", "smallint", "mediumint", "int", "integer", "serial":
		goType = "int32"
		break
	case "bigint", "bigserial":
		goType = "int64"
		break
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		goType = "string"
		break
	case "date", "datetime", "time", "timestamp":
		goType = "time.Time"
		break
	case "decimal", "numeric":
		goType = "float"
		break
	case "real", "float":
		goType = "float"
		break
	case "double":
		goType = "double"
		break
	case "tinyblob", "blob", "mediumblob", "longblob", "bytea":
		goType = "string"
		break
	case "bool":
		goType = "bool"
		break
	default:
		goType = "string"
	}
	return goType
}

//TitleCasedName 列名表名转大写
func TitleCasedName(name string) string {
	newstr := make([]rune, 0)
	upNextChar := true
	name = strings.ToLower(name)
	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			if 'a' <= chr && chr <= 'z' {
				chr -= ('a' - 'A')
			}
		case chr == '_':
			upNextChar = true
			continue
		}
		newstr = append(newstr, chr)
	}
	return string(newstr)
}
