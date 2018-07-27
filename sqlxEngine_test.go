package sqlxTool

import (
	"testing"
	"time"
)

type Class struct{
	Id int `xorm:"id"`
	Age int `xorm:"age"`
	Name string `xorm:"name"`
	Sal float64 `xorm:"sal"`
}
func Init(){
	DataSource("postgres","postgres://postgres:123@localhost:5432/test?sslmode=disable")
	DefaultConfig()
}
func TestRollingSql(t *testing.T) {
	str := "hellowr world"
	for i,_:=range str{
		t.Log(string(str[i]))
	}
}

func TestReplaceQuestionToDollar(t *testing.T) {
	t.Log(ReplaceQuestionToDollar("where   1 = ? and name = ? or name < ? and name > ? and created between ? and ? "))
	t.Log(ReplaceQuestionToDollarInherit("where   1 = ? and name = ? or name < ? and name > ? and created between ? and ?",7))
}

func TestGenWhereByStruct(t *testing.T) {
	type Tmp struct{
		Addr string `column:"and,addr,like*"`
		Desc string `column:"and,desc,like"`
		Job string`column:"and,job,*like"`
		Name string `column:"and,name,="`
		Sal float32 `column:"and,sal,>"`
		AgeMin int`column:"or,age,between"`
		AgeMax int `column:"or,age,between"`
		Start time.Time `column:"and,created,between"`
		Stop time.Time `column:"and,created,between"`
		Jump string `column:"-"`
	}
	var tmp = Tmp{
		Addr:"earth",
		Name:"ft",
		Sal:333,
		AgeMin:9,
		AgeMax:18,
		Desc:"happ",
		Job:"engineer",
		Jump:"jump",
	}

	t.Log(GenWhereByStruct(tmp))
}

func TestSelect(t *testing.T){
	Init()
	var clss = make([]Class,0)
	er:=Select("default",&clss,"select * from class")
	if er!=nil {
		t.Fatal(er.Error())
	}
	t.Log(clss)
}
func TestSelectOne(t *testing.T) {
	Init()
	var cls Class
	er:=SelectOne("default",&cls,"select * from class where id = 2")
	if er!=nil {
		t.Fatal(er.Error())
	}
	t.Log(cls)

	var count =0
	er=SelectOne("default",&count,"select count(id) from class")
	if er!=nil {
		t.Fatal(er.Error())
	}
	t.Log(count)
}

func TestDelete(t *testing.T) {
	Init()
	er:= Delete("default","delete from class where id = 1")
	if er!=nil {
		t.Fatal(er.Error())
	}
}

func TestUpdate(t *testing.T) {
	Init()
	er:=Update("default","update class set age=11 where id =2")
	if er!=nil {
		t.Fatal(er.Error())
	}
}

