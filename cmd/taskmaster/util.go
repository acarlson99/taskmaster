package main

import (
	"os/exec"
	"sort"
)

func InSlice(num int, nums []int) bool {
	idx := sort.Search(len(nums),
		func(ii int) bool { return nums[ii] >= num })
	return idx < len(nums) && nums[idx] == num
}

func CheckExit(err error, codes []int) (bool, error) {
	if err == nil {
		return InSlice(0, codes), nil
	} else if exiterr, ok := err.(*exec.ExitError); ok {
		code := exiterr.ExitCode()
		return InSlice(code, codes), nil
	}
	return false, err
}
