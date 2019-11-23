package literoute

func cleanURL(url *string) {
	_url := *url
	urlLength := len(_url)
	if urlLength > 1 {
		if (*url)[urlLength-1:] == "/" {
			*url = (*url)[:urlLength-1]
			cleanURL(url)
		}
	}
}

func valid(path string) bool {
	pathLength := len(path)
	if pathLength > 1 && path[pathLength-1:] == "/" {
		return false
	}
	return true
}
