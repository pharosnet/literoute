package literoute

type Validator interface {
	Validate(string) bool
	OnFail(Context)
}

type validatorInfo struct {
	start int
	end   int
	name  string
}

func containsValidators(path string) []validatorInfo {
	var index []int
	for i, c := range path {
		if c == '|' {
			index = append(index, i)
		}
	}

	if len(index) > 0 {
		var validators []validatorInfo
		for i, pos := range index {
			if i+1 == len(index) {
				validators = append(validators, validatorInfo{
					start: pos,
					end:   len(path),
					name:  path[pos:],
				})
			} else {
				validators = append(validators, validatorInfo{
					start: pos,
					end:   index[i+1],
					name:  path[pos:index[i+1]],
				})
			}
		}
		return validators
	}
	return nil
}
