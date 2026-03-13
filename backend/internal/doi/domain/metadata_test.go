package domain

import "testing"

func TestIsValid(t *testing.T) {
	validDOIs := []string{
		"10.1000/182",
		"10.1234/ABC-123",
		"10.5555/12345678",
		"10.1109/ISESE.2005.1541817",
		"10.1016/j.infsof.2023.107299",
		"10.1109/ICSTW58534.2023.00015",
		"10.1007/s10664-024-10522-z",
	}

	for _, doi := range validDOIs {
		if !IsValid(doi) {
			t.Errorf("IsValid(%q) = false; want true", doi)
		}
	}
}

func TestIsValid_invalid(t *testing.T) {
	invalidDOIs := []string{
		"",
		"10.1000",
		"10.1000/",
		"not-a-doi",
		"/10.1000/182",
		"10.1000/182 extra",
		"https://doi.org/10.1007/s10664-024-10522-z", // URL prefix, but the DOI part is valid
	}

	for _, doi := range invalidDOIs {
		if IsValid(doi) {
			t.Errorf("IsValid(%q) = true; want false", doi)
		}
	}
}
