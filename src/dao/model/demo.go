package model

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

/**
	這裏用來放草稿和構思
**/

type UserQueryParams struct {
	Page      Page       `json:"page"`
	ID        *int       `json:"id" query:"unique;sort=desc,2"`
	LoginTime *time.Time `json:"login_time" query:"sort=desc,2"`
	Score     *int       `json:"score" query:"aggr=sum;sort=desc,1"`
}

type QueryParam struct {
	Page      Page        `json:"page"`
	Sort      []Sort      `json:"sort"`
	Condition []Condition `json:"condition"`
}

func (q *QueryParam) build() {
	// 根據sort.weitht排序，從大到小升序
	sort.Slice(q.Sort, func(i, j int) bool {
		return q.Sort[i].Weight < q.Sort[j].Weight
	})

	// 構建condition
	for _, cond := range q.Condition {
		cond.build()
	}
}

type Aggr string

const (
	AggrNone  Aggr = "NONE"
	AggrCount Aggr = "COUNT"
	AggrSum   Aggr = "SUM"
	AggrAvg   Aggr = "AVG"
	AggrMax   Aggr = "MAX"
	AggrMin   Aggr = "MIN"
)

type Field struct {
	Name  string      `json:"name"`
	Aggr  Aggr        `json:"aggr"`
	Value interface{} `json:"value"` //指針指向的值
}

func buildField(structField string, value interface{}, aggrTag, fieldTag string) *Field {

}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

type Sort struct {
	Field  *Field    `json:"field"`
	Order  SortOrder `json:"order"`
	Weight int       `json:"weight"`
}

func buildSort(orderTag string) *Sort {
	sort := strings.TrimSpace(orderTag)
	if sort == "" {
		return &Sort{
			Order: SortOrderAsc,
		}
	}
	sort = strings.ReplaceAll(sort, " ", "")
	sort = strings.ReplaceAll(sort, "\t", "")

	split := strings.Split(sort, ",")
	order := strings.ToUpper(split[0])
	weight := 0
	if len(split) > 1 {
		weight, _ = strconv.Atoi(split[1])
	}
	od := SortOrder(order)

	if od != SortOrderAsc && od != SortOrderDesc {
		od = SortOrderAsc
	}

	return &Sort{
		Order:  od,
		Weight: weight,
	}
}

type ConditionExp string

const (
	ConditionExpEq         ConditionExp = "EQUAL"         //equal
	ConditionExpNe         ConditionExp = "NOT_EQUAL"     //not equal
	ConditionExpGt         ConditionExp = "GREATER"       //greater than
	ConditionExpGte        ConditionExp = "GREATER_EQUAL" //greater than or equal
	ConditionExpLt         ConditionExp = "LESS"          //less than
	ConditionExpLte        ConditionExp = "LESS_EQUAL"    //less than or equal
	ConditionExpLikeLeft   ConditionExp = "%LIKE"         //prefix like
	ConditionExpLikeRight  ConditionExp = "LIKE%"         //suffix like
	ConditionExpLikeBoth   ConditionExp = "%LIKE%"        //both prefix and suffix like
	ConditionExpIn         ConditionExp = "IN"            //in
	ConditionExpNotIn      ConditionExp = "NOT_IN"        //not in
	ConditionExpBetween    ConditionExp = "BETWEEN"       //between
	ConditionExpNotBetween ConditionExp = "NOT_BETWEEN"   //not between
	ConditionExpIsNull     ConditionExp = "NULL"          //is null
	ConditionExpIsNotNull  ConditionExp = "NOT_NULL"      //is not null
)

type Condition struct {
	Field *Field
	Exp   ConditionExp
}

func buildCondition(expTag string) *Condition {
	exp := strings.TrimSpace(expTag)
	if exp == "" {
		return &Condition{
			Exp: ConditionExpEq,
		}
	}
	exp = strings.ReplaceAll(exp, " ", "")
	exp = strings.ReplaceAll(exp, "\t", "")
	exp = strings.ToLower(exp)

	condExp := ConditionExp(exp)
	if condExp != ConditionExpEq && condExp != ConditionExpNe && condExp != ConditionExpGt && condExp != ConditionExpGte && condExp != ConditionExpLt && condExp != ConditionExpLte && condExp != ConditionExpLikeLeft && condExp != ConditionExpLikeRight && condExp != ConditionExpLikeBoth && condExp != ConditionExpIn && condExp != ConditionExpNotIn && condExp != ConditionExpBetween && condExp != ConditionExpNotBetween && condExp != ConditionExpIsNull && condExp != ConditionExpIsNotNull {
		condExp = ConditionExpEq
	}

	return &Condition{
		Exp: condExp,
	}
}
