package array_utils

func RemoveItem(array []string, item string) []string {
	var temp []string
	for _, a := range array {
		if item != a {
			temp = append(temp, a)
		}
	}
	return temp
}
