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
