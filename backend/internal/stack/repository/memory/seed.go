package memory

import (
	"time"

	domainPaper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func seedData() []domain.Stack {
	return []domain.Stack{
		{
			UUID:        "1",
			Owner:       "Will",
			Name:        "Agile & Software Quality Stack",
			Description: "Collection of research on agile software development, software quality and verification, and academic research methodologies.",
			Tags: []string{
				"agile",
				"software quality",
				"software engineering",
				"research methods",
				"empirical software engineering",
			},
			Visibility: domain.VisibilityPublic,
			Papers: []domainPaper.Paper{
				{UUID: "36583bb4-8cdc-554e-bcf5-f67b60d0b290", DOI: "10.48550/ARXIV.1709.08439", Title: "Agile Software Development Methods: Review and Analysis", TitleShort: "Agile Software Development Methods", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Pekka", NameLast: "Abrahamsson"}, domainPaper.Author{NameFirst: "Outi", NameLast: "Salo"}, domainPaper.Author{NameFirst: "Jussi", NameLast: "Ronkainen"}, domainPaper.Author{NameFirst: "Juhani", NameLast: "Warsta"}}, PublicationYear: "2017", PublicationStatus: "preprint", PublicationStatusTimestamp: "2017-01-01T00:00:00Z", Keywords: []string{"FOS: Computer and information sciences", "Software Engineering (cs.SE)"}, Type: "unpublished", PDFs: []string{"/Users/andi/Zotero/storage/77KPQCG6/Abrahamsson et al. - 2017 - Agile Software Development Methods Review and Ana.pdf"}, Metadata: domainPaper.Metadata{Publisher: "arXiv", References: []string{"https://arxiv.org/abs/1709.08439"}, DataSource: "references.bib"}},
				{UUID: "0f324174-926b-585d-b121-3a1e3f7fee0b", DOI: "10.1109/52.28121", Title: "Software Inspections: An Effective Verification Process", TitleShort: "Software Inspections", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "A.F.", NameLast: "Ackerman"}, domainPaper.Author{NameFirst: "L.S.", NameLast: "Buchwald"}, domainPaper.Author{NameFirst: "F.H.", NameLast: "Lewski"}}, PublicationYear: "1989", PublicationStatus: "published", PublicationStatusTimestamp: "1989-05-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/8MIILJDX/Ackerman et al. - 1989 - Software inspections an effective verification process.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "IEEE Software", Pages: "31--36", Volume: "6", Issue: "3", References: []string{"http://ieeexplore.ieee.org/document/28121/"}, DataSource: "references.bib"}},
				{UUID: "fee26da6-4f03-5665-905e-27bfc1815803", DOI: "10.1111/ijmr.12102", Title: "Shades of Grey: Guidelines for Working with the Grey Literature in Systematic Reviews for Management and Organizational Studies: Shades of Grey", TitleShort: "Shades of Grey", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Richard", NameMiddle: "J.", NameLast: "Adams"}, domainPaper.Author{NameFirst: "Palie", NameLast: "Smart"}, domainPaper.Author{NameFirst: "Anne", NameMiddle: "Sigismund", NameLast: "Huff"}}, PublicationYear: "2017", PublicationStatus: "published", PublicationStatusTimestamp: "2017-10-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/9MDJWMWZ/Adams et al. - 2017 - Shades of Grey Guidelines for Working with the Gr.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "International Journal of Management Reviews", Pages: "432--454", Volume: "19", Issue: "4", References: []string{"https://onlinelibrary.wiley.com/doi/10.1111/ijmr.12102"}, DataSource: "references.bib"}},
				{UUID: "8ca3b8ca-50f0-536a-8a75-9bef2033f39b", DOI: "10.1000/empty", Title: "Using Causal Inference and Bayesian Statistics to Explain the Capability of a Test Suite in Exposing Software Faults", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Alireza", NameLast: "Aghamohammadi"}, domainPaper.Author{NameFirst: "Seyed-Hassan", NameLast: "Mirian-Hosseinabadi"}}, PublicationYear: "2023", PublicationStatus: "prepublished", PublicationStatusTimestamp: "2023-03-17T00:00:00Z", Keywords: []string{"Computer Science - Software Engineering"}, Type: "online", PDFs: []string{"/Users/andi/Zotero/storage/KQWH8ZX6/Aghamohammadi and Mirian-Hosseinabadi - 2023 - Using causal inference and Bayesian statistics to .pdf"}, Metadata: domainPaper.Metadata{References: []string{"http://arxiv.org/abs/2303.09968"}, DataSource: "references.bib"}},
				{UUID: "050e4957-1948-5e6b-bc93-e1fa5045879d", DOI: "10.2139/ssrn.4857923", Title: "Distinguishing between Method and Methodology in Academic Research", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Gustavo", NameMiddle: "Jorge Martins De", NameLast: "Aguiar"}}, PublicationYear: "2024", PublicationStatus: "prepublished", PublicationStatusTimestamp: "2024-01-01T00:00:00Z", Type: "online", PDFs: []string{"/Users/andi/Zotero/storage/E26RV8W3/Aguiar - 2024 - Distinguishing between Method and Methodology in Academic Research.pdf"}, Metadata: domainPaper.Metadata{References: []string{"https://www.ssrn.com/abstract=4857923"}, DataSource: "references.bib"}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:        "2",
			Owner:       "Andy",
			Name:        "Code Review & Testing Stack",
			Description: "Collection of research on code review, software testing, and empirical software engineering.",
			Tags: []string{
				"code review",
				"software testing",
				"gui testing",
				"empirical software engineering",
				"open source",
			},
			Visibility: domain.VisibilityPrivate,
			Papers: []domainPaper.Paper{
				{UUID: "96ceaf05-06e6-599e-82b1-0e9e9e82993a", DOI: "10.1109/ESEM56168.2023.10304792", Title: "Exploring the Advances in Identifying Useful Code Review Comments", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Sharif", NameLast: "Ahmed"}, domainPaper.Author{NameFirst: "Nasir", NameMiddle: "U.", NameLast: "Eisty"}}, PublicationYear: "2023", PublicationStatus: "published", PublicationStatusTimestamp: "2023-10-26T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/RFXG2GC6/Ahmed and Eisty - 2023 - Exploring the Advances in Identifying Useful Code Review Comments.pdf"}, Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2023 ACM/IEEE International Symposium on Empirical Software Engineering and Measurement (ESEM)", Pages: "1--7", ISBN: []string{"978-1-6654-5223-6"}, References: []string{"https://ieeexplore.ieee.org/document/10304792/"}, DataSource: "references.bib"}},
				{UUID: "d00304a1-39d5-5cfe-9671-c5ac14a7567f", DOI: "10.1109/ICSTW.2018.00038", Title: "Challenges in Automated Testing Through Graphical User Interface", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Pekka", NameLast: "Aho"}, domainPaper.Author{NameFirst: "Tanja", NameLast: "Vos"}}, PublicationYear: "2018", PublicationStatus: "published", PublicationStatusTimestamp: "2018-04-01T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/7IBEIKSI/Aho and Vos - 2018 - Challenges in Automated Testing Through Graphical User Interface.pdf"}, Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2018 IEEE International Conference on Software Testing, Verification and Validation Workshops (ICSTW)", Pages: "118--121", ISBN: []string{"978-1-5386-6352-3"}, References: []string{"https://ieeexplore.ieee.org/document/8411741/"}, DataSource: "references.bib"}},
				{UUID: "8dce5e51-4e48-5096-a17b-1bd4442b22b4", DOI: "10.1109/ICSE.2019.00111", Title: "Why Does Code Review Work for Open Source Software Communities?", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Adam", NameLast: "Alami"}, domainPaper.Author{NameFirst: "Marisa", NameLast: "Leavitt Cohn"}, domainPaper.Author{NameFirst: "Andrzej", NameLast: "Wasowski"}}, PublicationYear: "2019", PublicationStatus: "published", PublicationStatusTimestamp: "2019-05-01T00:00:00Z", Type: "inproceedings", Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2019 IEEE/ACM 41st International Conference on Software Engineering (ICSE)", Pages: "1073--1083", ISBN: []string{"978-1-7281-0869-8"}, References: []string{"https://ieeexplore.ieee.org/document/8812037/"}, DataSource: "references.bib"}},
				{UUID: "ce127349-93e0-57ab-b9ed-c4037a1dc6bd", DOI: "10.1000/empty", Title: "On the Industrial Applicability of Visual Gui Testing", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Emil", NameLast: "Alégroth"}}, PublicationYear: "2013", PublicationStatus: "published", PublicationStatusTimestamp: "2013-01-01T00:00:00Z", Type: "thesis", Metadata: domainPaper.Metadata{PublishedIn: "Chalmers University of Technology", DataSource: "references.bib"}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:        "3",
			Owner:       "Andy",
			Name:        "UI Testing & Code Review Stack",
			Description: "Research on GUI-based testing, code review practices for testing artifacts, open-source dependency management, and software process improvement.",
			Tags: []string{
				"gui testing",
				"code review",
				"testing artifacts",
				"open source dependencies",
				"software process improvement",
				"software engineering",
			},
			Visibility: domain.VisibilityPublic,
			Papers: []domainPaper.Paper{
				{UUID: "dd0d9ce8-04e8-599c-846f-9acf05dd8e06", DOI: "10.1109/32.6156", Title: "The TAME Project: Towards Improvement-Oriented Software Environments", TitleShort: "The TAME Project", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "V.R.", NameLast: "Basili"}, domainPaper.Author{NameFirst: "H.D.", NameLast: "Rombach"}}, PublicationYear: "1988", PublicationStatus: "published", PublicationStatusTimestamp: "1988-06-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/5TBAD287/Basili and Rombach - 1988 - The TAME project towards improvement-oriented sof.pdf", "/Users/andi/Zotero/storage/ZBE5RVWL/Basili and Rombach - 1988 - The TAME project towards improvement-oriented sof.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "IEEE Transactions on Software Engineering", Pages: "758--773", Volume: "14", Issue: "6", References: []string{"http://ieeexplore.ieee.org/document/6156/"}, DataSource: "references.bib"}},
				{UUID: "e464998a-f0f4-5abb-8683-1a416e8ea27d", DOI: "10.1007/978-3-030-47240-5_3", Title: "Challenges of Tracking and Documenting Open Source Dependencies in Products: A Case Study", TitleShort: "Challenges of Tracking and Documenting Open Source Dependencies in Products", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Andreas", NameLast: "Bauer"}, domainPaper.Author{NameFirst: "Nikolay", NameLast: "Harutyunyan"}, domainPaper.Author{NameFirst: "Dirk", NameLast: "Riehle"}, domainPaper.Author{NameFirst: "Georg-Daniel", NameLast: "Schwarz"}}, PublicationYear: "2020", PublicationStatus: "published", PublicationStatusTimestamp: "2020-01-01T00:00:00Z", Type: "incollection", PDFs: []string{"/Users/andi/Zotero/storage/3ZKU3DXP/Bauer et al. - 2020 - Challenges of Tracking and Documenting Open Source.pdf"}, Metadata: domainPaper.Metadata{Publisher: "Springer International Publishing", PublishedIn: "Open Source Systems", Pages: "25--35", Volume: "582", ISBN: []string{"978-3-030-47239-9", "978-3-030-47240-5"}, References: []string{"http://link.springer.com/10.1007/978-3-030-47240-5_3"}, DataSource: "references.bib"}},
				{UUID: "349462da-0860-52f5-b7ce-a489e26b6f57", DOI: "10.1016/j.infsof.2023.107299", Title: "Code Review Guidelines for GUI-based Testing Artifacts", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Andreas", NameLast: "Bauer"}, domainPaper.Author{NameFirst: "Riccardo", NameLast: "Coppola"}, domainPaper.Author{NameFirst: "Emil", NameLast: "Alégroth"}, domainPaper.Author{NameFirst: "Tony", NameLast: "Gorschek"}}, PublicationYear: "2023", PublicationStatus: "published", PublicationStatusTimestamp: "2023-11-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/VMXMH6WR/Bauer et al. - 2023 - Code review guidelines for GUI-based testing artif.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "Information and Software Technology", Pages: "107299", Volume: "163", References: []string{"https://linkinghub.elsevier.com/retrieve/pii/S0950584923001532"}, DataSource: "references.bib"}},
				{UUID: "3df8adca-b17e-53f3-8835-55489c18cc8f", DOI: "10.1109/ICSTW58534.2023.00015", Title: "We Tried and Failed: An Experience Report on a Collaborative Workflow for GUI-based Testing", TitleShort: "We Tried and Failed", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Andreas", NameLast: "Bauer"}, domainPaper.Author{NameFirst: "Emil", NameLast: "Alégroth"}}, PublicationYear: "2023", PublicationStatus: "published", PublicationStatusTimestamp: "2023-04-01T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/LCY46S2B/Bauer and Alégroth - 2023 - We Tried and Failed An Experience Report on a Col.pdf"}, Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2023 IEEE International Conference on Software Testing, Verification and Validation Workshops (ICSTW)", Pages: "1--9", ISBN: []string{"979-8-3503-3335-0"}, References: []string{"https://ieeexplore.ieee.org/document/10132215/"}, DataSource: "references.bib"}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:        "4",
			Owner:       "Andy",
			Name:        "GitHub & Empirical Research Stack",
			Description: "Research on mining software repositories (especially GitHub), empirical studies on developer behavior, and collaborative software testing practices.",
			Tags: []string{
				"mining software repositories",
				"github mining",
				"empirical software engineering",
				"software analytics",
				"collaborative testing",
				"software testing",
			},
			Visibility: domain.VisibilityPublic,
			Papers: []domainPaper.Paper{
				{UUID: "222762f7-c9ac-5b73-bc60-aeee47422bbd", DOI: "10.1145/2597073.2597074", Title: "The Promises and Perils of Mining GitHub", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Eirini", NameLast: "Kalliamvakou"}, domainPaper.Author{NameFirst: "Georgios", NameLast: "Gousios"}, domainPaper.Author{NameFirst: "Kelly", NameLast: "Blincoe"}, domainPaper.Author{NameFirst: "Leif", NameLast: "Singer"}, domainPaper.Author{NameFirst: "Daniel", NameMiddle: "M.", NameLast: "German"}, domainPaper.Author{NameFirst: "Daniela", NameLast: "Damian"}}, PublicationYear: "2014", PublicationStatus: "published", PublicationStatusTimestamp: "2014-05-31T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/K377WUYI/Kalliamvakou et al. - 2014 - The promises and perils of mining GitHub.pdf"}, Metadata: domainPaper.Metadata{Publisher: "ACM", PublishedIn: "Proceedings of the 11th Working Conference on Mining Software Repositories", Pages: "92--101", ISBN: []string{"978-1-4503-2863-0"}, References: []string{"https://dl.acm.org/doi/10.1145/2597073.2597074"}, DataSource: "references.bib"}},
				{UUID: "2641da64-4e21-51f3-8a4f-c11c82876150", DOI: "10.1007/s10664-015-9393-5", Title: "An In-Depth Study of the Promises and Perils of Mining GitHub", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Eirini", NameLast: "Kalliamvakou"}, domainPaper.Author{NameFirst: "Georgios", NameLast: "Gousios"}, domainPaper.Author{NameFirst: "Kelly", NameLast: "Blincoe"}, domainPaper.Author{NameFirst: "Leif", NameLast: "Singer"}, domainPaper.Author{NameFirst: "Daniel", NameMiddle: "M.", NameLast: "German"}, domainPaper.Author{NameFirst: "Daniela", NameLast: "Damian"}}, PublicationYear: "2016", PublicationStatus: "published", PublicationStatusTimestamp: "2016-10-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/3NIF9J7G/Kalliamvakou et al. - 2016 - An in-depth study of the promises and perils of mining GitHub.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "Empirical Software Engineering", Pages: "2035--2071", Volume: "21", Issue: "5", References: []string{"http://link.springer.com/10.1007/s10664-015-9393-5"}, DataSource: "references.bib"}},
				{UUID: "f4b69269-7ef1-5950-b2a6-1f744cec0d4d", DOI: "10.1109/CTS.2013.6567249", Title: "Towards Collaborative Testing of Workflows in WMVC-based Web Applications", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Marcel", NameLast: "Karam"}, domainPaper.Author{NameFirst: "Haidar", NameLast: "Safa"}}, PublicationYear: "2013", PublicationStatus: "published", PublicationStatusTimestamp: "2013-05-01T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/HUIAVWFV/Karam and Safa - 2013 - Towards collaborative testing of workflows in WMVC.pdf"}, Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2013 International Conference on Collaboration Technologies and Systems (CTS)", Pages: "324--331", ISBN: []string{"978-1-4673-6404-1", "978-1-4673-6403-4", "978-1-4673-6402-7"}, References: []string{"http://ieeexplore.ieee.org/document/6567249/"}, DataSource: "references.bib"}},
				{UUID: "acb6f6f9-72d7-55cd-9d42-f9042df36872", DOI: "10.1109/ICST.2009.16", Title: "Empirical Observations on Software Testing Automation", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Katja", NameLast: "Karhu"}, domainPaper.Author{NameFirst: "Tiina", NameLast: "Repo"}, domainPaper.Author{NameFirst: "Ossi", NameLast: "Taipale"}, domainPaper.Author{NameFirst: "Kari", NameLast: "Smolander"}}, PublicationYear: "2009", PublicationStatus: "published", PublicationStatusTimestamp: "2009-04-01T00:00:00Z", Type: "inproceedings", Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2009 International Conference on Software Testing Verification and Validation", Pages: "201--209", ISBN: []string{"978-1-4244-3775-7"}, References: []string{"http://ieeexplore.ieee.org/document/4815352/"}, DataSource: "references.bib"}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:        "5",
			Owner:       "Andy",
			Name:        "Delete Stack",
			Description: "Research on mining software repositories (especially GitHub), empirical studies on developer behavior, and collaborative software testing practices.",
			Tags: []string{
				"mining software repositories",
				"github mining",
				"empirical software engineering",
				"software analytics",
				"collaborative testing",
				"software testing",
			},
			Visibility: domain.VisibilityPublic,
			Papers: []domainPaper.Paper{
				{UUID: "222762f7-c9ac-5b73-bc60-aeee47422bbd", DOI: "10.1145/2597073.2597074", Title: "The Promises and Perils of Mining GitHub", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Eirini", NameLast: "Kalliamvakou"}, domainPaper.Author{NameFirst: "Georgios", NameLast: "Gousios"}, domainPaper.Author{NameFirst: "Kelly", NameLast: "Blincoe"}, domainPaper.Author{NameFirst: "Leif", NameLast: "Singer"}, domainPaper.Author{NameFirst: "Daniel", NameMiddle: "M.", NameLast: "German"}, domainPaper.Author{NameFirst: "Daniela", NameLast: "Damian"}}, PublicationYear: "2014", PublicationStatus: "published", PublicationStatusTimestamp: "2014-05-31T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/K377WUYI/Kalliamvakou et al. - 2014 - The promises and perils of mining GitHub.pdf"}, Metadata: domainPaper.Metadata{Publisher: "ACM", PublishedIn: "Proceedings of the 11th Working Conference on Mining Software Repositories", Pages: "92--101", ISBN: []string{"978-1-4503-2863-0"}, References: []string{"https://dl.acm.org/doi/10.1145/2597073.2597074"}, DataSource: "references.bib"}},
				{UUID: "2641da64-4e21-51f3-8a4f-c11c82876150", DOI: "10.1007/s10664-015-9393-5", Title: "An In-Depth Study of the Promises and Perils of Mining GitHub", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Eirini", NameLast: "Kalliamvakou"}, domainPaper.Author{NameFirst: "Georgios", NameLast: "Gousios"}, domainPaper.Author{NameFirst: "Kelly", NameLast: "Blincoe"}, domainPaper.Author{NameFirst: "Leif", NameLast: "Singer"}, domainPaper.Author{NameFirst: "Daniel", NameMiddle: "M.", NameLast: "German"}, domainPaper.Author{NameFirst: "Daniela", NameLast: "Damian"}}, PublicationYear: "2016", PublicationStatus: "published", PublicationStatusTimestamp: "2016-10-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/3NIF9J7G/Kalliamvakou et al. - 2016 - An in-depth study of the promises and perils of mining GitHub.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "Empirical Software Engineering", Pages: "2035--2071", Volume: "21", Issue: "5", References: []string{"http://link.springer.com/10.1007/s10664-015-9393-5"}, DataSource: "references.bib"}},
				{UUID: "f4b69269-7ef1-5950-b2a6-1f744cec0d4d", DOI: "10.1109/CTS.2013.6567249", Title: "Towards Collaborative Testing of Workflows in WMVC-based Web Applications", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Marcel", NameLast: "Karam"}, domainPaper.Author{NameFirst: "Haidar", NameLast: "Safa"}}, PublicationYear: "2013", PublicationStatus: "published", PublicationStatusTimestamp: "2013-05-01T00:00:00Z", Type: "inproceedings", PDFs: []string{"/Users/andi/Zotero/storage/HUIAVWFV/Karam and Safa - 2013 - Towards collaborative testing of workflows in WMVC.pdf"}, Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2013 International Conference on Collaboration Technologies and Systems (CTS)", Pages: "324--331", ISBN: []string{"978-1-4673-6404-1", "978-1-4673-6403-4", "978-1-4673-6402-7"}, References: []string{"http://ieeexplore.ieee.org/document/6567249/"}, DataSource: "references.bib"}},
				{UUID: "acb6f6f9-72d7-55cd-9d42-f9042df36872", DOI: "10.1109/ICST.2009.16", Title: "Empirical Observations on Software Testing Automation", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "Katja", NameLast: "Karhu"}, domainPaper.Author{NameFirst: "Tiina", NameLast: "Repo"}, domainPaper.Author{NameFirst: "Ossi", NameLast: "Taipale"}, domainPaper.Author{NameFirst: "Kari", NameLast: "Smolander"}}, PublicationYear: "2009", PublicationStatus: "published", PublicationStatusTimestamp: "2009-04-01T00:00:00Z", Type: "inproceedings", Metadata: domainPaper.Metadata{Publisher: "IEEE", PublishedIn: "2009 International Conference on Software Testing Verification and Validation", Pages: "201--209", ISBN: []string{"978-1-4244-3775-7"}, References: []string{"http://ieeexplore.ieee.org/document/4815352/"}, DataSource: "references.bib"}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:        "6",
			Owner:       "Will",
			Name:        "Test Visibility Stack",
			Description: "Collection of research on agile software development, software quality and verification, and academic research methodologies.",
			Tags: []string{
				"agile",
				"software quality",
				"software engineering",
				"research methods",
				"empirical software engineering",
			},
			Visibility: domain.VisibilityPublic,
			Papers: []domainPaper.Paper{
				{UUID: "0f324174-926b-585d-b121-3a1e3f7fee0b", DOI: "10.1109/52.28121", Title: "Software Inspections: An Effective Verification Process", TitleShort: "Software Inspections", Authors: []domainPaper.Author{domainPaper.Author{NameFirst: "A.F.", NameLast: "Ackerman"}, domainPaper.Author{NameFirst: "L.S.", NameLast: "Buchwald"}, domainPaper.Author{NameFirst: "F.H.", NameLast: "Lewski"}}, PublicationYear: "1989", PublicationStatus: "published", PublicationStatusTimestamp: "1989-05-01T00:00:00Z", Type: "article", PDFs: []string{"/Users/andi/Zotero/storage/8MIILJDX/Ackerman et al. - 1989 - Software inspections an effective verification process.pdf"}, Metadata: domainPaper.Metadata{PublishedIn: "IEEE Software", Pages: "31--36", Volume: "6", Issue: "3", References: []string{"http://ieeexplore.ieee.org/document/28121/"}, DataSource: "references.bib"}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}
