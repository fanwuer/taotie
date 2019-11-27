package model

import (
	"strings"
	"xorm.io/xorm"
)

type PageHelp struct {
	Limit int64 `json:"limit"`
	Page  int64 `json:"page"`
}

func (page *PageHelp) Build(s *xorm.Session, sort []string, base []string) {
	build(s, sort, base)

	if page.Page == 0 {
		page.Page = 1
	}

	if page.Limit <= 0 {
		page.Limit = 20
	}

	if page.Limit > 200 {
		page.Limit = 200
	}
	s.Limit(int(page.Limit), int((page.Page-1)*page.Limit))
}

func build(s *xorm.Session, sort []string, base []string) {
	nowSort := make([]string, 0, len(sort))
	for _, v := range sort {
		nowSort = append(nowSort, v)
	}

	dict := make(map[string]struct{}, 0)

	for _, v := range base {
		a := v[1:]
		dict[a] = struct{}{}

		useBase := true
		for _, vv := range sort {
			b := vv[1:]
			if a == b {
				useBase = false
			}
		}

		if useBase {
			nowSort = append(nowSort, v)
		}
	}

	for _, v := range nowSort {
		a := v[1:]
		if _, ok := dict[a]; ok {
			if strings.HasPrefix(v, "+") {
				s.Asc(a)
			} else if strings.HasPrefix(v, "-") {
				s.Desc(a)

			}
		}
	}
}
