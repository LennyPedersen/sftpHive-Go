package utils

import "strings"

func Contains(s []string, str string) bool {
    for _, v := range s {
        if strings.TrimSpace(v) == str {
            return true
        }
    }
    return false
}