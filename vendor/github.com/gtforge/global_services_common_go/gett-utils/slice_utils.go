package gettUtils

import (
	"strconv"

	"github.com/gtforge/global_services_common_go/gett-utils/col"
)

func UniqueStrings(sourceArray []string) []string {
	tempMap := map[string]bool{}
	for _, key := range sourceArray {
		tempMap[key] = true
	}
	resultArray := []string{}
	for key := range tempMap {
		resultArray = append(resultArray, key)
	}
	return resultArray
}

func Contains(sourceArray []string, obj string) bool {
	for _, item := range sourceArray {
		if obj == item {
			return true
		}
	}
	return false
}

func ParseIntArray(sourceArray []string) []int {
	res := []int{}
	for _, s := range sourceArray {
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			res = append(res, int(i))
		}
	}
	return res
}

func StringArrayIntersection(s1, s2 []string) []string {

	return col.NewSet(s1).Intersection(s2).ToList()
}
