package sliceUtils

func RemoveItem(array []string, item string) []string {
	var temp []string
	for _, a := range array {
		if item != a {
			temp = append(temp, a)
		}
	}
	return temp
}

func Contains(array []string, item string) bool {
	for _, a := range array {
		if item == a {
			return true
		}
	}
	return false
}
