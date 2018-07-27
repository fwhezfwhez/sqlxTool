package sqlxTool

import (
	"testing"
	"time"
)

func Init(){
	DataSource("postgres","postgres://postgres:123@localhost:5432/postgres?sslmode=disable")
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