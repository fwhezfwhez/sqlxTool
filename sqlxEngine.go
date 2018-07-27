package sqlxTool

import (
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"errors"
	"strings"
	"strconv"
	"reflect"
	"time"
)

var Dbs map[string]*sqlx.DB
var LocalSessions map[string]*sqlx.Tx

func init() {
	Dbs = make(map[string]*sqlx.DB)
	LocalSessions = make(map[string]*sqlx.Tx)
}

//config a dataSource to "default" and init a global session
//for example:
/*
import db "sqlxTool"
const (
	host = "1.1.1.1" //database remote addr
	port = 5432    //database port
	user = "ft"    //login user
	password = "mypw"  //login password
	dbName="dbx"    //db name
)

func init() {
	dataSource := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",host,port,user,password,dbName)
	_,er:=db.DataSouce(dataSource)
	//you can also do it straightly like:
	//db.DataSource("postgres","postgres://postgres:123@localhost:5432/test?sslmode=disable")
}
 */
func DataSource(driverName string, dataSource string) (*sqlx.DB, error) {
	if driverName == "" {
		driverName = "postgres"
	}
	var err error
	db, err := sqlx.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}
	LocalSession, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	Dbs["default"] = db
	LocalSessions["default"] = LocalSession

	return db, nil
}
//import a datasource by initialed instance
func DataSourceImport(key string,db *sqlx.DB){
	Dbs[key] = db
	LocalSessions[key],_ = db.Beginx()
}

func NewDataSource(key string, driverName string, dataSource string) (*sqlx.DB, error) {
	if key == "" {
		return nil, errors.New("please specific the key of this db connection")
	}
	if driverName == "" {
		driverName = "postgres"
	}
	var err error
	db, err := sqlx.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}
	sessionNew, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	Dbs[key] = db
	LocalSessions[key] = sessionNew
	return db, nil
}

//config a db's max idle and open connection
func Config(key string, maxIdleConns int, maxOpenConns int) error {
	if key == "" {
		for k, _ := range Dbs {
			Dbs[k].SetMaxIdleConns(maxIdleConns)
			Dbs[k].SetMaxOpenConns(maxOpenConns)
		}
		return nil
	} else {
		if v, ok := Dbs[key]; !ok {
			return errors.New("cannot find key: " + key + " in Dbs,use db.DataSource(key,driverName,dataSource string) to init")
		} else {
			v.SetMaxOpenConns(maxOpenConns)
			v.SetMaxIdleConns(maxIdleConns)
		}
		return nil
	}
}

//default config 
func DefaultConfig() error {
	for k, _ := range Dbs {
		Dbs[k].SetMaxIdleConns(2000)
		Dbs[k].SetMaxOpenConns(2000)
	}
	return nil
}

//get Db instance which type is *sqlx.DB
func GetDb(key string) (*sqlx.DB) {
	if key == "" {
		return Dbs["default"]
	}
	return Dbs[key]
}

//select one object from database,using dataSource whose dbKey is specific and 'default' by empty key
func DynamicSelectOne(dbKey string, dest interface{}, basicSql string, whereMap [][]string, orderBy []string, Asc string, limit int, offset int, args ...interface{}) error {
	var db *sqlx.DB
	if dbKey == "" {
		db = Dbs["default"]
	} else {
		db = Dbs[dbKey]
	}
	sql := RollingSql(basicSql, whereMap, orderBy, Asc, limit, offset)
	args = RemoveZero(args)
	return db.Get(dest, sql, args...)
}

//select objects array from database,using dataSource whose dbKey is specific and 'default' by empty key
func DynamicSelect(dbKey string, dest interface{}, basicSql string, whereMap [][]string, orderBy []string, Asc string, limit int, offset int, args ...interface{}) error {
	var db *sqlx.DB
	if dbKey == "" {
		db = Dbs["default"]
	} else {
		db = Dbs[dbKey]
	}
	sql := RollingSql(basicSql, whereMap, orderBy, Asc, limit, offset)
	args = RemoveZero(args)
	return db.Select(dest, sql, args...)
}

//dynamic update an object
func DynamicUpdate(dbKey string, basicSql string, whereMap [][]string,  args ...interface{}) error{
	var db *sqlx.DB
	if dbKey == "" {
		db = Dbs["default"]
	} else {
		db = Dbs[dbKey]
	}
	sql := RollingSql(basicSql, whereMap, nil, "", -1, -1)
	args = RemoveZero(args)
	_,er:=db.Exec( sql, args...)
	return er
}

func DynamicInsert(dbKey string,basicSql string,columns []string,args ...interface{}){

}
//Only prepare exec without tx.RollBack() and tx.Commit()
func DynamicSelectOneSpecificTx(tx *sqlx.Tx, dest interface{}, basicSql string, whereMap [][]string, orderBy []string, Asc string, limit int, offset int, args ...interface{}) error {
	if tx==nil {
		return errors.New("tx is nil,please use Db.BeginTran() to create a tx")
	}
	sql := RollingSql(basicSql, whereMap, orderBy, Asc, limit, offset)
	args = RemoveZero(args)
	return tx.Get(dest, sql, args...)
}

//Only prepare exec without tx.RollBack() and tx.Commit()
func DynamicSelectSpecificTx(tx *sqlx.Tx, dest interface{}, basicSql string, whereMap [][]string, orderBy []string, Asc string, limit int, offset int, args ...interface{}) error {
	if tx==nil {
		return errors.New("tx is nil,please use Db.BeginTran() to create a tx")
	}
	sql := RollingSql(basicSql, whereMap, orderBy, Asc, limit, offset)
	args = RemoveZero(args)
	return tx.Select(dest, sql, args...)
}

//Only prepare exec without tx.RollBack() and tx.Commit()
func DynamicUpdateSpecificTx(tx *sqlx.Tx, basicSql string, whereMap [][]string,  args ...interface{}) error{
	if tx==nil {
		return errors.New("tx is nil,please use Db.BeginTran() to create a tx")
	}
	sql := RollingSql(basicSql, whereMap, nil, "", -1, -1)
	args = RemoveZero(args)
	_,er:=tx.Exec( sql, args...)
	return er
}

//normal query one object
func SelectOne(dbKey string, dest interface{},sql string,args...interface{})error{
	var db *sqlx.DB
	if dbKey == "" {
		db = Dbs["default"]
	} else {
		db = Dbs[dbKey]
	}
	return db.Get(dest,sql,args...)
}
//normal query objects array
func Select(dbKey string, dest interface{},sql string,args...interface{})error{
	var db *sqlx.DB
	if dbKey == "" {
		db = Dbs["default"]
	} else {
		db = Dbs[dbKey]
	}
	return db.Select(dest,sql,args...)
}

//normal delete
func Delete(dbKey string,sql string,args...interface{})error{
	return Exec(dbKey,sql ,args...)
}

//normal update
func Update(dbKey string,sql string,args...interface{})error{
	return Exec(dbKey,sql ,args...)
}

//normal exec
func Exec(dbKey string,sql string,args...interface{})error{
	var db *sqlx.DB
	if dbKey == "" {
		db = Dbs["default"]
	} else {
		db = Dbs[dbKey]
	}
	_,er :=db.Exec(sql,args...)
	return er
}


func RollingSql(basicSql string, whereMap [][]string, orderBy []string, Asc string, limit int, offset int) string {
	var sql = basicSql
	//1.处理where
	if len(whereMap) != 0 {
		sql = sql + " where "
		for _, v := range whereMap {
			//v[0]表示性质，and 还是or,v[1]表示field，比如name，age,v[2]表示条件符号,=,>,<,<>,like
			if v[2] == "between" {
				sql = sql + " " + v[0] + " " + v[1] + " " + "between" + " " + "?" + " " + "and" + " " + "?" + " "
				continue
			}
			if v[2] == "in" {
				sql = sql + " " + v[0] + " " + v[1] + " " + "in" + " " + v[3]
				continue
			}
			sql = sql + " " + v[0] + " " + v[1] + " " + v[2] + " " + "?"
		}
	}
	//fmt.Println("处理where完毕:"+sql)

	//2.处理Orderby和asc
	if len(orderBy) != 0 && orderBy != nil {
		sql = sql + " order by " + strings.Join(orderBy, ",") + " " + Asc + " "
	}
	//fmt.Println("处理order,asc完毕:"+sql)

	//3.处理limit,offset
	if limit != -1 && offset != -1 {
		sql = sql + " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)
	}
	//fmt.Println(sql)
	sql,_ = ReplaceQuestionToDollar(sql)
	return sql
}

//将sql语句中的?转换成$i
func ReplaceQuestionToDollar(sql string) (string,int) {
	var temp = 1
	start := 0
	var i = 0
L:
	for i = start; i < len(sql); i++ {
		if string(sql[i]) == "?" {
			sql = string(sql[:i]) + "$" + strconv.Itoa(temp) + string(sql[i+1:])
			temp++
			start = i + 2
			goto L
		}

		if i == len(sql)-1 {
			return sql,temp
		}
	}
	return sql,temp
}

//将sql语句中的?转换成$i, i存在初始值offset
func ReplaceQuestionToDollarInherit(sql string,offset int) (string,int) {
	if offset <1 {
		return ReplaceQuestionToDollar(sql)
	}
	temp := offset

	start := 0
	var i = 0
L:
	for i = start; i < len(sql); i++ {
		if string(sql[i]) == "?" {
			sql = string(sql[:i]) + "$" + strconv.Itoa(temp) + string(sql[i+1:])
			temp++
			start = i + 2
			goto L
		}

		if i == len(sql)-1 {
			return sql,temp
		}
	}
	return sql,temp
}

func RemoveZero(slice []interface{}) []interface{} {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if IfZero(v) {
			slice = append(slice[:i], slice[i+1:]...)
			return RemoveZero(slice)
			break
		}
	}
	return slice
}

func IfZero(arg interface{}) bool {
	if arg == nil {
		return true
	}
	switch v := arg.(type) {
	case int, float64, int32, int16, int64, float32:
		if v == 0 {
			return true
		}
	case string:
		if v == "" || v == "%%" || v == "%" {
			return true
		}
	case *string, *int, *int64, *int32, *int16, *int8, *float32, *float64:
		if v == nil {
			return true
		}
	case time.Time:
		return v.IsZero()
	default:
		return false
	}
	return false
}

//generate where through a where [][]string
func GenWhere(whereMap [][]string)string {
	rs:=""
	if len(whereMap) != 0 {
		rs = rs + " where "
		for _, v := range whereMap {
			//v[0]表示性质，and 还是or,v[1]表示field，比如name，age,v[2]表示条件符号,=,>,<,<>,like
			if v[2] == "between" {
				rs = rs + " " + v[0] + " " + v[1] + " " + "between" + " " + "?" + " " + "and" + " " + "?" + " "
				continue
			}
			if v[2] == "in" {
				rs = rs + " " + v[0] + " " + v[1] + " " + "in" + " " +v[3]
				continue
			}
			rs = rs + " " + v[0] + " " + v[1] + " " + v[2] + " " + "?"
		}
	}
	return rs
}

//generate  a where through a struct
func GenWhereByStruct(in interface{})(string,[]interface{}){
	vValue :=reflect.ValueOf(in)
	vType :=reflect.TypeOf(in)
	var tagTmp =""
	var whereMap = make([][]string,0)
	var args = make([]interface{},0)

	for i:=0;i<vValue.NumField();i++{
		tagTmp = vType.Field(i).Tag.Get("column")
		if tagTmp =="-"||tagTmp==""{
			continue
		}
		cons :=strings.Split(tagTmp,",")
		if !IfZero(vValue.Field(i).Interface()) {
			if cons[2]=="*like"{
				cons[2] = "like"
				args = append(args, "%"+vValue.Field(i).Interface().(string))
			}else if cons[2]=="like*"{
				cons[2] = "like"
				args = append(args, vValue.Field(i).Interface().(string)+"%")
			}else if cons[2]=="*like*" || cons[2]=="like"{
				cons[2] = "like"
				args = append(args, "%"+vValue.Field(i).Interface().(string)+"%")
			}else{
				args = append(args, vValue.Field(i).Interface())
			}

			if len(whereMap)==0 {
				whereMap = append(whereMap,[]string{
					"",cons[1],cons[2],
				})
			}else{
				whereMap = append(whereMap,[]string{
					cons[0],cons[1],cons[2],
				})
			}

			if cons[2] == "between"{
				i++
				args = append(args,vValue.Field(i).Interface())
			}
		}
	}
	where :=GenWhere(whereMap)
	return where,args
}