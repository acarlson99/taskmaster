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

func GetExitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	} else if exiterr, ok := err.(*exec.ExitError); ok {
		return exiterr.ExitCode(), nil
	} else {
		return -1, err
	}
}

func CheckExit(err error, codes []int) (bool, error) {
	code, exErr := GetExitCode(err)
	if exErr != nil {
		return false, err
	} else {
		return InSlice(code, codes), nil
	}
}
