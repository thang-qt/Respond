package store

type pagination struct {
	Page    int
	PerPage int
	Offset  int
}

func normalizePagination(page, perPage, defaultPerPage, maxPerPage int) pagination {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return pagination{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}
