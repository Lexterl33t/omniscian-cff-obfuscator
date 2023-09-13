package utils

import "os"

func ReadF(file string) (string, error) {
	dat, err := os.ReadFile(file)

	if err != nil {
		return "", err
	}

	return string(dat), nil
}
