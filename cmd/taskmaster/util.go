package main

import (
	"fmt"
	"os/exec"
	"sort"
)

// return true, false if num in sorted slice
// NOTE: slice MUST be sorted
func SearchSlice(num int, nums []int) bool {
	idx := sort.Search(len(nums),
		func(ii int) bool { return nums[ii] >= 0 })
	if idx < len(nums) && nums[idx] == 0 {
		return true
	} else {
		return false
	}
}

func CheckExit(err error, codes []int) (bool, error) {
	// handle exit code 0
	if err == nil {
		return SearchSlice(0, codes), nil
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		code := exiterr.ExitCode()
		fmt.Println(code)
		ok := SearchSlice(code, codes)
		return ok, nil
	}
	return false, err
}
