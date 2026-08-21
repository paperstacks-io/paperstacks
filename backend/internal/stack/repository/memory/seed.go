package memory

import (
	"time"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const demoUserExternalID = "dbe3febc-ab91-486c-b51f-38ab0f59a4d9"

var (
	seedStackTimestamp1 = time.Date(2026, time.April, 17, 14, 32, 0, 0, time.UTC)
	seedStackTimestamp2 = time.Date(2026, time.January, 20, 11, 32, 0, 0, time.UTC)
	seedStackTimestamp3 = time.Date(2026, time.March, 5, 14, 32, 0, 0, time.UTC)
	seedStackTimestamp4 = time.Date(2026, time.June, 11, 14, 32, 0, 0, time.UTC)
	seedStackTimestamp5 = time.Date(2026, time.February, 14, 14, 32, 0, 0, time.UTC)
)

func seedData() []domain.Stack {
	return []domain.Stack{
		{UUID: "9e1a819a-24ab-47b6-be29-92b49325e4c2", Name: "Code Review", IsPublic: true, Owner: demoUser(), Papers: codeReviewPapers(), CreatedAt: seedStackTimestamp1, UpdatedAt: seedStackTimestamp1},
		{UUID: "a8de9118-e7d9-4a0b-9e77-7872f08d8efa", Name: "Testing", IsPublic: true, Owner: demoUser(), Papers: testingPapers(), CreatedAt: seedStackTimestamp2, UpdatedAt: seedStackTimestamp2},
		{UUID: "c6ff032d-104f-4f5f-a9d7-87f874c75c0a", Name: "Secondary Studies", IsPublic: true, Owner: demoUser(), Papers: secondaryStudiesPapers(), CreatedAt: seedStackTimestamp3, UpdatedAt: seedStackTimestamp3},
		{UUID: "d67f909d-84a3-4c4e-823c-0c9a20e89790", Name: "Bayesian", IsPublic: true, Owner: demoUser(), Papers: bayesianPapers(), CreatedAt: seedStackTimestamp4, UpdatedAt: seedStackTimestamp4},
		{UUID: "f4152eb2-b303-461c-a683-5bfe80258f8e", Name: "Research Methodology", IsPublic: false, Owner: demoUser(), Papers: researchMethodologyPapers(), CreatedAt: seedStackTimestamp5, UpdatedAt: seedStackTimestamp5},
	}
}

func demoUser() userDomain.User {
	return userDomain.User{ExternalID: demoUserExternalID, Email: "demo@paperstacks.io"}
}

var seedPapersByUUID = indexSeedPapers(paperMemory.SeedData())

func indexSeedPapers(papers []paperDomain.Paper) map[string]paperDomain.Paper {
	papersByUUID := make(map[string]paperDomain.Paper, len(papers))
	for _, paper := range papers {
		papersByUUID[paper.UUID] = paper
	}

	return papersByUUID
}

func seedPaper(uuid string) paperDomain.Paper {
	paper, ok := seedPapersByUUID[uuid]
	if !ok {
		panic("missing paper seed data for " + uuid)
	}

	return paper
}

func codeReviewPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("0f324174-926b-585d-b121-3a1e3f7fee0b"),
		seedPaper("96ceaf05-06e6-599e-82b1-0e9e9e82993a"),
		seedPaper("8dce5e51-4e48-5096-a17b-1bd4442b22b4"),
		seedPaper("9e691f0e-a3df-56e8-bf5f-7c27f786573f"),
		seedPaper("48da0a49-5dd5-5b12-96db-fb0c0f409fe6"),
		seedPaper("2b53961d-db83-5c68-9f82-9febe0827495"),
		seedPaper("15427cd2-2d98-5e28-9a9f-62cbb2076496"),
		seedPaper("bf1fe1c4-25ca-50f3-ab28-694416ea770d"),
		seedPaper("543deeea-4ec1-55e7-822b-c48414f5f8ed"),
		seedPaper("b2d4e8dc-45f0-5bc8-ae6d-ce05765838c8"),
		seedPaper("349462da-0860-52f5-b7ce-a489e26b6f57"),
		seedPaper("661d999a-5f56-53a9-8c2e-e9d401c7fd07"),
		seedPaper("641f851e-b6fe-5148-9680-dab2970112c3"),
		seedPaper("14815b6a-6e2d-5b73-8aa3-1a2d7b517106"),
		seedPaper("93e5d6ed-cffc-526d-ba74-bfe2b9834863"),
		seedPaper("404d1dd8-a939-558f-a888-86daa304758a"),
		seedPaper("8fd47375-9722-55f3-8196-939353d0f592"),
		seedPaper("eeeb65b5-b710-596f-8f6e-e537f6f92fed"),
		seedPaper("8b446d32-87b2-5e0e-9975-02360d88b70c"),
		seedPaper("cfdf41e6-8f96-54c1-89e1-0b655d974d03"),
		seedPaper("b618d4ef-aee6-5261-8ed8-851cb0dab48b"),
		seedPaper("bc192fa5-b66d-5f29-b689-d0ce8c3cd930"),
		seedPaper("825fbfa5-d9fb-5767-8d52-4f8703a11db7"),
		seedPaper("5874bb31-30c9-5621-8a71-4033e38d4e80"),
		seedPaper("5ed113ec-f48c-53c3-9c64-94232a481252"),
		seedPaper("b28986d3-7774-579d-831b-5688d6a5e003"),
		seedPaper("516d41ef-147b-5c3a-9afa-b1e72ab0b145"),
		seedPaper("94f75519-c062-5cf6-8f3b-89a579b7dfc1"),
		seedPaper("87d3e0e4-887d-5081-b1be-5254f7c3d451"),
		seedPaper("7dacf3f6-2fa8-5687-9a16-146efcdd6984"),
		seedPaper("aa49b808-3364-54e7-9455-a9afa106ca8b"),
		seedPaper("b6a727af-afb6-51f3-9bcc-899acf86c62c"),
		seedPaper("2de544db-ba76-5a9b-bfa3-eb545ae4c833"),
		seedPaper("fc84ac81-c484-52fa-b553-5f88ec74dca6"),
		seedPaper("e462b32f-9eea-5d52-abe5-ec86da68c139"),
		seedPaper("01bf669e-b316-525a-9b93-1f14453fca44"),
		seedPaper("8f1de5dd-0aee-5b62-8c63-054ce4ff9f5f"),
		seedPaper("121de110-7bda-5a2b-9c21-461758b8fb64"),
		seedPaper("f21f1f94-bad4-5c46-ae3a-5b8d071219d1"),
		seedPaper("30178b92-c31e-5e61-bc53-578bde4ec94d"),
		seedPaper("590ae0d4-f6a3-52ce-8823-d695ee28f3f7"),
		seedPaper("18929dd7-eca1-576a-bb05-ba6ebf541755"),
		seedPaper("ff678e7b-87cb-57af-b870-985ee31c431b"),
		seedPaper("d228256d-dc0e-56e6-bb46-8c4081cee198"),
		seedPaper("ed13b1f8-7e03-5239-a49e-86a48db030ca"),
		seedPaper("65f3cd01-8c87-54fe-af35-23342f184d54"),
		seedPaper("6df179e1-c5b6-5ba7-8d13-9d6b41e75618"),
		seedPaper("df051ddc-4722-5aad-895c-7529e3c33530"),
		seedPaper("f990680a-03b4-519b-96e6-1a6c2bed6419"),
		seedPaper("ba79ef8e-4573-59cc-a246-2c6110ca262f"),
	}
}

func testingPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("8ca3b8ca-50f0-536a-8a75-9bef2033f39b"),
		seedPaper("d00304a1-39d5-5cfe-9671-c5ac14a7567f"),
		seedPaper("ce127349-93e0-57ab-b9ed-c4037a1dc6bd"),
		seedPaper("e1075d78-39fc-57a1-84aa-73f4645a0753"),
		seedPaper("2076a25d-e1c6-5d12-ac73-71458a2df853"),
		seedPaper("b23eb741-b0a2-5201-9814-22bf448806af"),
		seedPaper("19b76db9-3ecf-5ee9-94ff-2d15342f1c1f"),
		seedPaper("5611515a-515a-5d8d-8fee-8e554994416f"),
		seedPaper("0bd4a755-d844-55f8-a200-4d6ef990c460"),
		seedPaper("5f5ca994-661a-5fd9-83a5-e5908ba7f87b"),
		seedPaper("a1067b9d-0595-5797-b9fd-41070bfc2d7a"),
		seedPaper("41277a19-8396-5f6e-9971-4cf475a66038"),
		seedPaper("9b3aff9a-61de-5c35-870b-7eb7959c51d5"),
		seedPaper("f280229c-8848-5b57-8f73-e44acce66830"),
		seedPaper("c875491e-9ba6-5c20-99a1-2ebc22e9ddfc"),
		seedPaper("5ce401cf-e4b8-56f1-99a9-95fb740a67b9"),
		seedPaper("bd83efd6-aae0-59f0-aef3-26a5374ab8be"),
		seedPaper("ea0fde12-7e32-5cec-869d-8f61ea8e5a74"),
		seedPaper("ac4345f6-da8c-54c2-9365-73f9bdee8cad"),
		seedPaper("3df8adca-b17e-53f3-8835-55489c18cc8f"),
		seedPaper("55c2240f-2da6-5f22-9124-1e90a90f2add"),
		seedPaper("458a3f14-7242-5d63-bcd0-a37c79e4856a"),
		seedPaper("953aebd1-560b-5dad-b101-418241284140"),
		seedPaper("1796c5ea-d284-55be-b4fb-6bd8e176671d"),
		seedPaper("93905617-7495-5180-80b5-733789696a2e"),
		seedPaper("cac901b1-1ef3-5ae4-8f62-3bd7bfc86220"),
		seedPaper("0a8a3d6a-54f0-5c16-a06b-4654bc5cc402"),
		seedPaper("f1fdbe42-635c-5dc1-a26b-ca6feff1ba01"),
		seedPaper("d50411d3-78a9-5b74-b2f4-23c41933eddb"),
		seedPaper("74ec2737-429e-529a-a5fb-722fe810ed74"),
		seedPaper("e7595e7a-97e5-5e23-a8fa-921bdd933a9e"),
		seedPaper("558be10e-6953-5bd8-a71c-54baa141e977"),
		seedPaper("8c0c603e-1c5f-5ccc-9cdf-12ddf2072a4b"),
		seedPaper("468f4ef3-a416-5be6-a245-cbaa033ac6ba"),
		seedPaper("7373e504-7b6d-5302-8854-5e64ef592024"),
		seedPaper("75984cb2-73ab-5ea7-b4e9-64420ae04ffc"),
		seedPaper("d215ce73-b4e8-51ad-be9c-189b70f77206"),
		seedPaper("4d0cccb3-040a-5322-a540-88d4827463cd"),
		seedPaper("1dd3a60a-5612-5075-8c83-55c62c136daa"),
		seedPaper("e7e9d846-029e-5c31-a18c-575dcb935ecc"),
		seedPaper("6e1a2895-2458-52fa-b75d-e63fff04a738"),
		seedPaper("5966a651-6ed5-5bbc-bfe0-c76331d373a6"),
		seedPaper("5f14ebf6-ecd9-573e-a865-4a6ef307b542"),
		seedPaper("c491a607-c3a5-5672-9fe5-1887337a9302"),
		seedPaper("0b947dbd-02b1-5915-a97b-0934a78d0d8e"),
		seedPaper("33ca1920-46c1-55c9-aa1f-99e7e13673e3"),
		seedPaper("e5a22cb0-8bf1-53c1-8643-f2e70df95cd6"),
		seedPaper("bb6700d6-fd29-531d-b982-cba478879e8e"),
		seedPaper("b7e17ad6-53d4-5706-800e-24ee38e4f8e6"),
		seedPaper("6c7de263-dc9a-5486-b961-9944b4618a1c"),
		seedPaper("a266da52-ef13-5c36-8df7-ad18092d200d"),
		seedPaper("d5449c2a-84c3-5dbb-b5bb-3f5b904c6ece"),
		seedPaper("24f4be49-54ea-5270-bbdc-2ba9dbb2b9ec"),
		seedPaper("6b38853d-3db8-53e0-bfa5-1ddf97fae12a"),
		seedPaper("05d0f513-1c1f-57d2-ba23-08e090839413"),
		seedPaper("92734ac1-4eff-552e-baf3-fa3c9d0c7b20"),
		seedPaper("acb6f6f9-72d7-55cd-9d42-f9042df36872"),
		seedPaper("1efc568d-7748-5911-a3cb-6e3e6a847355"),
		seedPaper("80220a79-6cb6-5398-a384-72083db02abb"),
		seedPaper("84b2b6b8-db75-5255-868e-14d4fa954059"),
		seedPaper("c955e641-dd73-5322-9662-48c3ffc4cb33"),
		seedPaper("0b689e5d-6963-594a-a9e1-bd6436ccb0d2"),
		seedPaper("f64fad0e-534c-5f1c-b1eb-99ba3d57f35c"),
		seedPaper("a82c9c22-8203-584d-a606-8dda113b0eee"),
		seedPaper("fb66d480-ac7f-56b8-9837-6b9b0087b458"),
		seedPaper("61f1c112-3068-58ab-bccb-45d89bf42717"),
		seedPaper("6ba7cb92-7954-5cf1-87c1-c4ebcb0c92cf"),
		seedPaper("514bd880-abd6-5656-8c7e-69d05e29d1ec"),
	}
}

func secondaryStudiesPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("fee26da6-4f03-5665-905e-27bfc1815803"),
		seedPaper("6df179e1-c5b6-5ba7-8d13-9d6b41e75618"),
		seedPaper("9b3aff9a-61de-5c35-870b-7eb7959c51d5"),
		seedPaper("bce3eac8-0156-5518-83e0-753f47ec5e12"),
		seedPaper("ea0fde12-7e32-5cec-869d-8f61ea8e5a74"),
		seedPaper("8c0c603e-1c5f-5ccc-9cdf-12ddf2072a4b"),
		seedPaper("8fd47375-9722-55f3-8196-939353d0f592"),
		seedPaper("1af38bde-d295-5515-9dc6-ebadb11c54dc"),
		seedPaper("35177f2b-816a-5611-a55a-41083843c3fb"),
		seedPaper("a900e357-a3b9-5baf-8202-f2b40f920855"),
		seedPaper("36ad41cb-88ba-562e-b656-5c9a4973ee54"),
		seedPaper("6df5173d-7e06-5c4d-adc3-abeba3daa3f5"),
		seedPaper("9497413f-ad82-5e99-9225-18eb6f8f33c9"),
		seedPaper("21eecc20-d375-5049-9006-eb85adac4aa1"),
		seedPaper("c29cd7e8-984b-5fd8-9d93-da8fc0f0fcb9"),
		seedPaper("7a061311-b9e0-50ff-a39c-c4f9e5395d51"),
		seedPaper("1d7be506-3050-52f6-92f1-9a01c0caee49"),
		seedPaper("0837a72b-a72c-5c37-9152-462d91ced844"),
		seedPaper("d1af8ede-bad4-580d-b342-e1ac4e26f9de"),
	}
}

func bayesianPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("8ca3b8ca-50f0-536a-8a75-9bef2033f39b"),
		seedPaper("9d6cb335-5a27-5b62-bbdc-c1ca5087393b"),
		seedPaper("88874208-6f1d-520c-929a-311a3fcc7338"),
		seedPaper("68bddc7e-fe18-5e60-9078-ed16a07f6e0a"),
		seedPaper("00007003-e8c8-51f9-910d-d81ce21063a3"),
		seedPaper("49293981-c8e5-5bcc-b7c9-1474f3ef99b6"),
		seedPaper("b63b8503-dea4-5ee7-9709-26a4e0e54999"),
		seedPaper("2397b703-67d8-5340-b243-0adb21928961"),
		seedPaper("8f1fb3cc-6df8-53ae-b322-e0624da68f26"),
	}
}

func researchMethodologyPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("050e4957-1948-5e6b-bc93-e1fa5045879d"),
		seedPaper("cef605e1-cafe-5d29-bb9d-e242400c53c9"),
		seedPaper("1d84292b-a152-5f43-ba9c-4850e35227a8"),
		seedPaper("734c9f00-2910-54c6-a739-a33be104cd90"),
		seedPaper("1d01ab6a-d6b7-54d6-8541-820148131735"),
		seedPaper("c81901b5-8c0b-58e9-8b10-4ec084f63e20"),
		seedPaper("952e5fac-40ff-567f-887b-578e26149744"),
		seedPaper("4e3143ce-e7f2-5029-8d63-7caa957c0cee"),
		seedPaper("3523796c-63a9-5c5d-bb1a-581747689718"),
		seedPaper("1fa1a590-5f0d-529e-a6ec-5175a769fca4"),
		seedPaper("a4e72361-b5a2-5a8f-a1df-ebf123e3a8d7"),
		seedPaper("26d6d631-c853-5959-8bde-bb61b7481155"),
		seedPaper("c5847dee-8359-56cf-a2b7-efd5eb7c5226"),
		seedPaper("02368094-41ae-5b4b-8c86-d5adc81e411e"),
		seedPaper("653cf6dd-671d-5d2f-90c9-727998a68901"),
		seedPaper("0f88b274-a1f9-5b30-9b95-b9e290551301"),
		seedPaper("23d58066-49f8-5a28-9ad0-bc0df5d57038"),
		seedPaper("8307cec1-379c-5ca5-99a8-25037e7c4db6"),
		seedPaper("d58e17b1-bc50-57f7-a535-9ce39fb89231"),
		seedPaper("4aebf490-f7c0-55ee-8def-5c2e34cb21c8"),
		seedPaper("da710511-65d1-581d-b5da-58bd63939a29"),
		seedPaper("79a104e5-207e-5549-8447-ab59bc57cb23"),
		seedPaper("cb9a80de-2823-5795-9bbe-eece204683db"),
		seedPaper("65edf0ea-cf45-533d-a0e4-82a56b408cb2"),
		seedPaper("223fe559-0f75-51b9-b0e0-c72f3ad2aeca"),
		seedPaper("4aaa11c1-71cb-530c-b43a-8b97bd1cc284"),
		seedPaper("b47d7b79-d586-55fa-b4c1-d1ad856513d0"),
		seedPaper("452837c8-bed2-5282-99ca-c8bbf3888fd0"),
		seedPaper("11d26b29-8b5d-5da0-b683-7c1d89c58a89"),
		seedPaper("ec3b4b79-2516-5d28-a316-1538232efcb8"),
		seedPaper("d16f6c97-34c1-56c5-a618-abeb8dcb616e"),
		seedPaper("df2367ff-f0da-51a1-98ca-31de203cc092"),
	}
}
