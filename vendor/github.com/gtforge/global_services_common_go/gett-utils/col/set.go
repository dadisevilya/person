package col

type Set map[string]bool

func NewSet(values []string) *Set {
	s := &Set{}
	s.Add(values...)
	return s
}

func (s *Set) Contains(other string) bool {
	_, ok := (*s)[other]
	return ok
}

func (s Set) Intersects(other []string) bool {
	if len(s) == 0 {
		return false
	}

	for i := range other {
		if _, ok := s[other[i]]; ok {
			return true
		}
	}

	return false
}

func (s *Set) Intersection(other []string) *Set {
	if len(*s) == 0 {
		return &Set{}
	}

	res := &Set{}

	for i := range other {
		if s.Contains(other[i]) {
			res.Add(other[i])
		}
	}

	return res
}

func (s *Set) Add(vals ...string) {
	for i := range vals {
		(*s)[vals[i]] = true
	}
}

func (s *Set) Remove(vals ...string) *Set {
	for _, v := range vals {
		delete(*s, v)
	}
	return s
}

func (s *Set) ToList() []string {
	res := []string{}
	for v := range *s {
		res = append(res, v)
	}
	return res
}
