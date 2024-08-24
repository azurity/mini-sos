package tree

import (
	"io/fs"
	"strings"
)

const Separator = "/"

type Path []string

func NewPath(path string) Path {
	path = strings.TrimSpace(path)
	if len(path) > 0 && !strings.HasPrefix(path, Separator) && !strings.HasPrefix(path, "."+Separator) && path != "." {
		path = "./" + path
	}
	parts := strings.Split(path, Separator)
	ret := []string{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		ret = append(ret, part)
	}
	return ret
}

func (path Path) Normalize() Path {
	if len(path) == 0 {
		return path
	}
	relative := path[0] == "."
	if relative {
		path = path[1:]
	}
	ret := []string{}
	for _, part := range path {
		if part == ".." {
			ret = ret[:len(ret)-1]
		} else {
			ret = append(ret, part)
		}
	}
	if relative {
		ret = append([]string{"."}, ret...)
	}
	return ret
}

func (path Path) IsRel() bool {
	return len(path) >= 1 && path[0] == "."
}

func (path Path) Str() string {
	ret := strings.Join(path, Separator)
	if !path.IsRel() {
		ret = "/" + ret
	}
	return ret
}

// func (path Path) Parent() Path {
// 	if len(path) == 0 {
// 		return Path{}
// 	} else {
// 		return append(Path{}, path[:len(path)-1]...)
// 	}
// }

func Join(path ...Path) Path {
	ret := Path{}
	for _, item := range path {
		normalized := item.Normalize()
		if normalized.IsRel() {
			ret = append(ret, normalized[1:]...)
		} else {
			ret = normalized
		}
	}
	return ret.Normalize()
}

func Rel(path Path, base Path) (Path, error) {
	normalizedPath := path.Normalize()
	normalizedBase := base.Normalize()
	if base.IsRel() {
		return nil, fs.ErrInvalid
	}
	if normalizedPath.IsRel() {
		return normalizedPath, nil
	}
	maxLen := min(len(normalizedPath), len(normalizedBase))
	sameLen := 0
	for i := 0; i < maxLen; i++ {
		if normalizedPath[i] == normalizedBase[i] {
			sameLen += 1
		} else {
			break
		}
	}
	ret := Path{"."}
	for i := 0; i < len(normalizedBase)-sameLen; i++ {
		ret = append(ret, "..")
	}
	if len(normalizedPath) > sameLen {
		ret = append(ret, normalizedPath[sameLen:]...)
	}
	return ret, nil
}
