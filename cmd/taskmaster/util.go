package main

import (
	"fmt"
	"os/exec"
	"sort"
)

func InSlice(num int, nums []int) bool {
	idx := sort.Search(len(nums),
		func(ii int) bool { return nums[ii] >= 0 })
	if idx < len(nums) && nums[idx] == 0 {
		return true
	} else {
		return false
	}
}

func CheckExit(err error, codes []int) (bool, error) {
	if err == nil {
		return InSlice(0, codes), nil
	} else if exiterr, ok := err.(*exec.ExitError); ok {
		code := exiterr.ExitCode()
		fmt.Println(code)
		return InSlice(code, codes), nil
	}
	return false, err
}
