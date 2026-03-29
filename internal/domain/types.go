package domain

type ModID string
type CategoryID string
type PlaysetIndex int

const CategoryPrefix = "category:"

func IsCategoryID(raw string) bool {
	return len(raw) > len(CategoryPrefix) && raw[:len(CategoryPrefix)] == CategoryPrefix
}
