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
	driver := ""  //Mysql驱动，目前只支持mysql
	sqlconn := "" //数据库链接
	tplPath := "" //模板文件位置  ./template/gorm.tpl
	goPath := ""  //生成出来的go代码存放位置，不传的话生成在当前文件夹下
	dbInfo := &models.DbInfo{}
	//curPath, _ := os.Getwd()
	//fmt.Println(curPath)

	//打印参数,第0个参数是exe的路径
	if len(os.Args) < 4 {
		fmt.Println("参数不全！！")
		return
	} else {
		driver = os.Args[1]
		sqlconn = os.Args[2]
		tplPath = os.Args[3]
		fmt.Println("driver:", driver)
		fmt.Println("sqlconn:", sqlconn)
		fmt.Println("tplPath:", tplPath)
	}
	if len(os.Args) > 4 {
		goPath = os.Args[4]
		fmt.Println("goPath:", goPath)
	}

	//获取要生成模板的绝对路径
	genDir, _ := filepath.Abs(goPath)
	fmt.Println("curr_dir:", genDir)
	//return
	//获取实体类的命名空间,注意：path.Base不能解析windows下的“\”，需要替换为“/”
	goNameSapce := path.Base(strings.Replace(genDir, "\\", "/", -1))
	//fmt.Println("goNameSapce:", goNameSapce)

	//setp 1、初始化数据库连接
	initDB(driver, sqlconn)
	//setp 2、获取数据库所有的表和列，并初始化好命名空间
	tables := getDataBaseInfo(goNameSapce)

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
	os.MkdirAll(genDir, os.ModePerm)
	//遍历表,生成实体
	for i, table := range dbInfo.DbTables {
		fmt.Println(i+1, "/", len(dbInfo.DbTables), table.TableName)
		// if i > 0 {
		// 	break
		// }
		//根据模板生成最终数据
		buf := new(bytes.Buffer)
		tplFile.Execute(buf, table)
		//fmt.Println(buf.String())
		//生成go文件路径
		currFile := path.Join(genDir, fmt.Sprintf("%s%s", table.TableName, ".go"))
		f, err := os.Create(currFile)
		defer f.Close()
		if err == nil {
			//保存文件
			f.Write([]byte(buf.String()))
		} else {
			fmt.Println(err)
		}
	}
}
func saveStructFile() {

}

//initdb 初始化数据库连接池
func initDB(driver string, sqlconn string) {
	var err error
	Dbo, err = gorm.Open("mysql", "root:123456@(127.0.0.1:3306)/scriptdb?charset=utf8")
	if err != nil {
		panic(err)
	} else {
		//Dbo.LogMode(true)
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
	//循环表信息，设置命名空间，并把表的列名都添加进去方便template使用
	for i := range tables {
		tables[i].PackageName = spaceName
		tables[i].GoName = TitleCasedName(tables[i].TableName) //格式化表名
		tables[i].Imports = make(map[string]string)            //初始化当前表的包
		//这里减少了重复循环，前提是表名、列名里的TABLE_NAME排过序的，否则会有问题
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
		goType = "float64"
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
