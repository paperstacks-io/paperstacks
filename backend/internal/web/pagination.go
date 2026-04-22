package web

type PaginationItem struct {
	Page       int
	IsActive   bool
	IsEllipsis bool
	TargetPage int
}

func BuildPagination(
	total, pageSize, currentPage int,
) []PaginationItem {
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	if totalPages <= 0 {
		return nil
	}

	const visiblePagesCount = 7

	var items []PaginationItem

	firstPage := 1
	lastPage := totalPages
	windowSize := visiblePagesCount
	lastWindowStart := totalPages - windowSize + 1
	nextWindowTarget := windowSize + 1

	// Show all pages if they fit within the visible window
	if totalPages <= windowSize {
		for page := firstPage; page <= lastPage; page++ {
			items = appendPage(items, page, currentPage, totalPages)
		}
		return items
	}

	/*
		Beginning:
		Show the first window the last page
		Example: 1 2 3 4 5 6 7 ... last
	*/
	if currentPage < windowSize {
		for page := firstPage; page <= windowSize; page++ {
			items = appendPage(items, page, currentPage, totalPages)
		}
		items = appendEllipsis(items, nextWindowTarget, totalPages)
		items = appendPage(items, lastPage, currentPage, totalPages)
		return items
	}

	/*
		End:
		Show first page, ellipsis, and the last window
		Example: 1 ... last
	*/
	if currentPage > totalPages-(windowSize-1) {
		items = appendPage(items, firstPage, currentPage, totalPages)
		items = appendEllipsis(items, totalPages-windowSize, totalPages)

		for page := lastWindowStart; page <= lastPage; page++ {
			items = appendPage(items, page, currentPage, totalPages)
		}
		return items
	}

	/*
		Middle:
		Show first page, ellipsis, current page, ellipsis, and last page
		Example: 1 ... [window] ... last
	*/
	halfWindowSize := windowSize / 2
	windowStart := currentPage - halfWindowSize
	windowEnd := currentPage + halfWindowSize

	items = appendPage(items, firstPage, currentPage, totalPages)
	items = appendEllipsis(items, currentPage-windowSize, totalPages)

	for page := windowStart; page <= windowEnd; page++ {
		items = appendPage(items, page, currentPage, totalPages)
	}

	items = appendEllipsis(items, currentPage+windowSize, totalPages)
	items = appendPage(items, lastPage, currentPage, totalPages)

	return items
}

func appendPage(items []PaginationItem, page, currentPage, totalPages int) []PaginationItem {
	if page < 1 || page > totalPages {
		return items
	}

	return append(items, PaginationItem{
		Page:     page,
		IsActive: page == currentPage,
	})
}

func appendEllipsis(items []PaginationItem, targetPage, totalPages int) []PaginationItem {
	if targetPage < 1 {
		targetPage = 1
	}
	if targetPage > totalPages {
		targetPage = totalPages
	}

	return append(items, PaginationItem{
		IsEllipsis: true,
		TargetPage: targetPage,
	})
}
