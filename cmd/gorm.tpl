package {{.PackageName}}
{{$ilen := len .Imports}}{{if gt $ilen 0}}
import (
	{{range .Imports}}"{{.}}"{{end}}
)
{{end}}
//{{.GoName}} {{.TableComment}}
type {{.GoName}} struct { {{range .TableColumns}}
    {{.GoName}}     {{.GoType}}     `gorm:"type:{{.ColumnType}};column:{{.ColumnName}};{{if eq .ColumnKey "PRI"}}primary_key;{{end}}{{if eq .Extra "auto_increment"}}AUTO_INCREMENT;{{end}}{{if eq .IsNullable "NO"}}not null;{{end}}"`  //{{.ColumnComment}} {{end}}
}

//TableName 实际数据库表名
func ({{.GoName}}) TableName() string {
	return "{{.TableName}}"
}