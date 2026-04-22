package web

type PaginationItem struct {
	IsActive    bool
	IsSeperator bool
	TargetPage  int
}

// BuildPagination generates pagination items with a fixex size of 10
func BuildPagination(
	total, pageSize, currentPage int,
) []PaginationItem {
	if total <= 0 || pageSize <= 0 {
		return nil
	}

	totalPages := (total + pageSize - 1) / pageSize
	currentPage = max(1, min(currentPage, totalPages))

	items := make([]PaginationItem, 0, min(totalPages, 10))
	addPage := func(page int) {
		items = append(items, PaginationItem{
			IsActive:   page == currentPage,
			TargetPage: page,
		})
	}
	addSeparator := func(targetPage int) {
		items = append(items, PaginationItem{
			IsSeperator: true,
			TargetPage:  targetPage,
		})
	}

	if totalPages <= 10 {
		for page := 1; page <= totalPages; page++ {
			addPage(page)
		}
		return items
	}

	// Keep the control width stable at 10 items while still showing the
	// current page, the boundaries, and separators when pages are skipped.
	if currentPage <= 5 {
		for page := 1; page <= 8; page++ {
			addPage(page)
		}
		addSeparator(9)
		addPage(totalPages)
		return items
	}

	if currentPage >= totalPages-4 {
		addPage(1)
		addSeparator(totalPages - 8)
		for page := totalPages - 7; page <= totalPages; page++ {
			addPage(page)
		}
		return items
	}

	addPage(1)
	addSeparator(currentPage - 3)
	for page := currentPage - 2; page <= currentPage+3; page++ {
		addPage(page)
	}
	addSeparator(currentPage + 4)
	addPage(totalPages)
	return items
}
