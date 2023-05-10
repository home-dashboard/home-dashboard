package utils

import "os"

func FileExist(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
