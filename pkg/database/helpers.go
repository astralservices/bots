package database

func GetPagination(page, size int) (from, to int) {
	if size == 0 {
		size = 3
	}
	if page == 0 {
		page = 1
	}
	from = (page - 1) * size
	to = from + size - 1
	return
}
