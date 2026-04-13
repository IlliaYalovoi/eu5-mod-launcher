package domain

type (
	ModID        string
	CategoryID   string
	PlaysetIndex int
)

const CategoryPrefix = "category:"

func IsCategoryID(raw string) bool {
	return len(raw) > len(CategoryPrefix) && raw[:len(CategoryPrefix)] == CategoryPrefix
}
