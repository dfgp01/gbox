package internal

import "fmt"

/**
* 盡可能覆蓋所有結構體的場景
**/

type (
	NormalStruct struct {
		Int  int
		Str  string
		Byt  []byte
		Bool bool
	}

	AliasStruct NormalStruct

	EmbedStruct struct {
		NormalStruct
		Float float64
	}

	EmbedPtrStruct struct {
		*NormalStruct
		Float float64
	}

	EmbedAliasStruct struct {
		AliasStruct
		Name string
	}

	EmbedAliasPtrStruct struct {
		*AliasStruct
		Name string
	}

	AdvancedStruct struct {
		Ptr         *AliasStruct    //has a ptr struct
		Sub         []NormalStruct  //has a slice struct
		SubPtr      []*NormalStruct //has a slice ptr struct
		InnerStruct struct {        //has a inner struct
			Id  int
			Act bool
		}
		EmptyStruct struct{}
	}
)

func (a *AliasStruct) AliasPtrMethod() string {
	return fmt.Sprintf("AliasPtrMethod: %v", a.Str)
}

func (a AliasStruct) AliasMethod() string {
	return fmt.Sprintf("AliasMethod: %v", a.Str)
}

func StructRun() {

}
