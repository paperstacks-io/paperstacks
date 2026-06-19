package memory

import (
	"time"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
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
		{UUID: "f4152eb2-b303-461c-a683-5bfe80258f8e", Name: "Research Methodology", IsPublic: true, Owner: demoUser(), Papers: researchMethodologyPapers(), CreatedAt: seedStackTimestamp5, UpdatedAt: seedStackTimestamp5},
	}
}

func demoUser() userDomain.User {
	return userDomain.User{ExternalID: demoUserExternalID}
}

func seedPaper(uuid, title string) paperDomain.Paper {
	return paperDomain.Paper{UUID: uuid, Title: title}
}

func codeReviewPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("0f324174-926b-585d-b121-3a1e3f7fee0b", "Software Inspections: An Effective Verification Process"),
		seedPaper("96ceaf05-06e6-599e-82b1-0e9e9e82993a", "Exploring the Advances in Identifying Useful Code Review Comments"),
		seedPaper("8dce5e51-4e48-5096-a17b-1bd4442b22b4", "Why Does Code Review Work for Open Source Software Communities?"),
		seedPaper("9e691f0e-a3df-56e8-bf5f-7c27f786573f", "State‐of‐the‐art: Software Inspections after 25 Years"),
		seedPaper("48da0a49-5dd5-5b12-96db-fb0c0f409fe6", "Expectations, Outcomes, and Challenges of Modern Code Review"),
		seedPaper("2b53961d-db83-5c68-9f82-9febe0827495", "Modern Code Reviews - Preliminary Results of a Systematic Mapping Study"),
		seedPaper("15427cd2-2d98-5e28-9a9f-62cbb2076496", "Modern Code Reviews—Survey of Literature and Practice"),
		seedPaper("bf1fe1c4-25ca-50f3-ab28-694416ea770d", "Code Reviews Enhance Software Quality"),
		seedPaper("543deeea-4ec1-55e7-822b-c48414f5f8ed", "Augmenting Code Review Experience Through Visualization"),
		seedPaper("b2d4e8dc-45f0-5bc8-ae6d-ce05765838c8", "Helping Developers Help Themselves: Automatic Decomposition of Code Review Changesets"),
		seedPaper("349462da-0860-52f5-b7ce-a489e26b6f57", "Code Review Guidelines for GUI-based Testing Artifacts"),
		seedPaper("661d999a-5f56-53a9-8c2e-e9d401c7fd07", "A Faceted Classification Scheme for Change-Based Industrial Code Review Processes"),
		seedPaper("641f851e-b6fe-5148-9680-dab2970112c3", "Four Eyes Are Better than Two: On the Impact of Code Reviews on Software Quality"),
		seedPaper("14815b6a-6e2d-5b73-8aa3-1a2d7b517106", "Characteristics of Useful Code Reviews: An Empirical Study at Microsoft"),
		seedPaper("93e5d6ed-cffc-526d-ba74-bfe2b9834863", "Process Aspects and Social Dynamics of Contemporary Code Review: Insights from Open Source Development and Industrial Practice at Microsoft"),
		seedPaper("404d1dd8-a939-558f-a888-86daa304758a", "Modern Code Review"),
		seedPaper("8fd47375-9722-55f3-8196-939353d0f592", "A Systematic Literature Review and Taxonomy of Modern Code Review"),
		seedPaper("eeeb65b5-b710-596f-8f6e-e537f6f92fed", "Tales from the Trenches: Expectations and Challenges from Practice for Code Review in the Generative AI Era"),
		seedPaper("8b446d32-87b2-5e0e-9975-02360d88b70c", "Survey on Pains and Best Practices of Code Review"),
		seedPaper("cfdf41e6-8f96-54c1-89e1-0b655d974d03", "Confusion Detection in Code Reviews"),
		seedPaper("b618d4ef-aee6-5261-8ed8-851cb0dab48b", "Confusion in Code Reviews: Reasons, Impacts, and Coping Strategies"),
		seedPaper("bc192fa5-b66d-5f29-b689-d0ce8c3cd930", "Individual, Social and Personnel Factors Influencing Modern Code Review Process"),
		seedPaper("825fbfa5-d9fb-5767-8d52-4f8703a11db7", "Knowledge Sharing, a Key Sustainable Practice Is on Risk: An Insight from Modern Code Review"),
		seedPaper("5874bb31-30c9-5621-8a71-4033e38d4e80", "Knowledge Sharing Factors for Modern Code Review to Minimize Software Engineering Waste"),
		seedPaper("5ed113ec-f48c-53c3-9c64-94232a481252", "Graph-Based Visualization of Merge Requests for Code Review"),
		seedPaper("b28986d3-7774-579d-831b-5688d6a5e003", "Code Review and Cooperative Pair Programming Best Practice"),
		seedPaper("516d41ef-147b-5c3a-9afa-b1e72ab0b145", "Do Explicit Review Strategies Improve Code Review Performance?"),
		seedPaper("94f75519-c062-5cf6-8f3b-89a579b7dfc1", "Do Explicit Review Strategies Improve Code Review Performance? Towards Understanding the Role of Cognitive Load"),
		seedPaper("87d3e0e4-887d-5081-b1be-5254f7c3d451", "Using a Balanced Scorecard to Identify Opportunities to Improve Code Review Effectiveness: An Industrial Experience Report"),
		seedPaper("7dacf3f6-2fa8-5687-9a16-146efcdd6984", "The Impact of Design and Code Reviews on Software Quality: An Empirical Study Based on PSP Data"),
		seedPaper("aa49b808-3364-54e7-9455-a9afa106ca8b", "Code Review Quality: How Developers See It"),
		seedPaper("b6a727af-afb6-51f3-9bcc-899acf86c62c", "Code Reviewing in the Trenches: Challenges and Best Practices"),
		seedPaper("2de544db-ba76-5a9b-bfa3-eb545ae4c833", "What Types of Defects Are Really Discovered in Code Reviews?"),
		seedPaper("fc84ac81-c484-52fa-b553-5f88ec74dca6", "An Empirical Study of the Impact of Modern Code Review Practices on Software Quality"),
		seedPaper("e462b32f-9eea-5d52-abe5-ec86da68c139", "Visualizing Code and Coverage Changes for Code Review"),
		seedPaper("01bf669e-b316-525a-9b93-1f14453fca44", "Information Needs in Contemporary Code Review"),
		seedPaper("8f1de5dd-0aee-5b62-8c63-054ce4ff9f5f", "Open Source Software Peer Review Practices: A Case Study of the Apache Server"),
		seedPaper("121de110-7bda-5a2b-9c21-461758b8fb64", "Understanding Broadcast Based Peer Review on Open Source Software Projects"),
		seedPaper("f21f1f94-bad4-5c46-ae3a-5b8d071219d1", "Contemporary Peer Review in Action: Lessons from Open Source Development"),
		seedPaper("30178b92-c31e-5e61-bc53-578bde4ec94d", "Convergent Contemporary Software Peer Review Practices"),
		seedPaper("590ae0d4-f6a3-52ce-8823-d695ee28f3f7", "The Effect of Checklist in Code Review for Inexperienced Students: An Empirical Study"),
		seedPaper("18929dd7-eca1-576a-bb05-ba6ebf541755", "Modern Code Review: A Case Study at Google"),
		seedPaper("ff678e7b-87cb-57af-b870-985ee31c431b", "What's Bothering Developers in Code Review?"),
		seedPaper("d228256d-dc0e-56e6-bb46-8c4081cee198", "When Testing Meets Code Review: Why and How Developers Review Tests"),
		seedPaper("ed13b1f8-7e03-5239-a49e-86a48db030ca", "Test-Driven Code Review: An Empirical Study"),
		seedPaper("65f3cd01-8c87-54fe-af35-23342f184d54", "Can Peer Code Reviews Be Exploited for Later Information Needs?"),
		seedPaper("6df179e1-c5b6-5ba7-8d13-9d6b41e75618", "What Makes a Code Review Useful to OpenDev Developers? An Empirical Investigation"),
		seedPaper("df051ddc-4722-5aad-895c-7529e3c33530", "All Eyes on the Reviewer: Understanding the Impact of GenAI on Mental Workload and Performance in Code Reviews"),
		seedPaper("f990680a-03b4-519b-96e6-1a6c2bed6419", "EvaCRC: Evaluating Code Review Comments"),
		seedPaper("ba79ef8e-4573-59cc-a246-2c6110ca262f", "Automatically Recommending Peer Reviewers in Modern Code Review"),
	}
}

func testingPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("8ca3b8ca-50f0-536a-8a75-9bef2033f39b", "Using Causal Inference and Bayesian Statistics to Explain the Capability of a Test Suite in Exposing Software Faults"),
		seedPaper("d00304a1-39d5-5cfe-9671-c5ac14a7567f", "Challenges in Automated Testing Through Graphical User Interface"),
		seedPaper("ce127349-93e0-57ab-b9ed-c4037a1dc6bd", "On the Industrial Applicability of Visual Gui Testing"),
		seedPaper("e1075d78-39fc-57a1-84aa-73f4645a0753", "JAutomate: A Tool for System- and Acceptance-test Automation"),
		seedPaper("2076a25d-e1c6-5d12-ac73-71458a2df853", "Conceptualization and Evaluation of Component-Based Testing Unified with Visual GUI Testing: An Empirical Study"),
		seedPaper("b23eb741-b0a2-5201-9814-22bf448806af", "Visual GUI Testing in Practice: Challenges, Problemsand Limitations"),
		seedPaper("19b76db9-3ecf-5ee9-94ff-2d15342f1c1f", "Maintenance of Automated Test Suites in Industry: An Empirical Study on Visual GUI Testing"),
		seedPaper("5611515a-515a-5d8d-8fee-8e554994416f", "On the Long-Term Use of Visual Gui Testing in Industrial Practice: A Case Study"),
		seedPaper("0bd4a755-d844-55f8-a200-4d6ef990c460", "Continuous Integration and Visual GUI Testing: Benefits and Drawbacks in Industrial Practice"),
		seedPaper("5f5ca994-661a-5fd9-83a5-e5908ba7f87b", "A Failed Attempt at Creating Guidelines for Visual GUI Testing: An Industrial Case Study"),
		seedPaper("a1067b9d-0595-5797-b9fd-41070bfc2d7a", "Practitioners’ Best Practices to Adopt, Use or Abandon Model-based Testing with Graphical Models for Software-intensive Systems"),
		seedPaper("41277a19-8396-5f6e-9971-4cf475a66038", "Automated Unit Test Improvement Using Large Language Models at Meta"),
		seedPaper("9b3aff9a-61de-5c35-870b-7eb7959c51d5", "Crowdsourced Software Testing: A Systematic Literature Review"),
		seedPaper("f280229c-8848-5b57-8f73-e44acce66830", "AI in GUI-Based Software Testing: Insights from a Survey with Industrial Practitioners"),
		seedPaper("c875491e-9ba6-5c20-99a1-2ebc22e9ddfc", "Espresso vs. EyeAutomate: An Experiment for the Comparison of Two Generations of Android GUI Testing"),
		seedPaper("5ce401cf-e4b8-56f1-99a9-95fb740a67b9", "Trends in Model-based GUI Testing"),
		seedPaper("bd83efd6-aae0-59f0-aef3-26a5374ab8be", "Usability Testing: A Software Engineering Perspective"),
		seedPaper("ea0fde12-7e32-5cec-869d-8f61ea8e5a74", "Graphical User Interface (GUI) Testing: Systematic Mapping and Repository"),
		seedPaper("ac4345f6-da8c-54c2-9365-73f9bdee8cad", "The Oracle Problem in Software Testing: A Survey"),
		seedPaper("3df8adca-b17e-53f3-8835-55489c18cc8f", "We Tried and Failed: An Experience Report on a Collaborative Workflow for GUI-based Testing"),
		seedPaper("55c2240f-2da6-5f22-9124-1e90a90f2add", "Augmented Testing to Support Manual GUI-based Regression Testing: An Empirical Study"),
		seedPaper("458a3f14-7242-5d63-bcd0-a37c79e4856a", "An Empirical Analysis of the Distribution of Unit Test Smells and Their Impact on Software Maintenance"),
		seedPaper("953aebd1-560b-5dad-b101-418241284140", "Observations and Lessons Learned from Automated Testing"),
		seedPaper("1796c5ea-d284-55be-b4fb-6bd8e176671d", "Automated System Testing Using Visual GUI Testing Tools: A Comparative Study in Industry"),
		seedPaper("93905617-7495-5180-80b5-733789696a2e", "Automated Gui Testing Guided by Usage Profiles"),
		seedPaper("cac901b1-1ef3-5ae4-8f62-3bd7bfc86220", "GUI Testing Using Computer Vision"),
		seedPaper("0a8a3d6a-54f0-5c16-a06b-4654bc5cc402", "Bad Smells and Refactoring Methods for GUI Test Scripts"),
		seedPaper("f1fdbe42-635c-5dc1-a26b-ca6feff1ba01", "Improving Crowd-Supported GUI Testing with Structural Guidance"),
		seedPaper("d50411d3-78a9-5b74-b2f4-23c41933eddb", "A Simple and Practical Approach to Unit Testing: The JML and JUnit Way"),
		seedPaper("74ec2737-429e-529a-a5fb-722fe810ed74", "Fragility of Layout-Based and Visual GUI Test Scripts: An Assessment Study on a Hybrid Mobile Application"),
		seedPaper("e7595e7a-97e5-5e23-a8fa-921bdd933a9e", "Scripted GUI Testing of Android Open-Source Apps: Evolution of Test Code and Fragility Causes"),
		seedPaper("558be10e-6953-5bd8-a71c-54baa141e977", "Mobile Testing: New Challenges and Perceived Difficulties From Developers of the Italian Industry"),
		seedPaper("8c0c603e-1c5f-5ccc-9cdf-12ddf2072a4b", "A Taxonomy of Metrics for GUI-based Testing Research: A Systematic Literature Review"),
		seedPaper("468f4ef3-a416-5be6-a245-cbaa033ac6ba", "Conceptualization of Multi-user Collaborative GUI-Testing for Web Applications"),
		seedPaper("7373e504-7b6d-5302-8854-5e64ef592024", "On Effectiveness and Efficiency of Gamified Exploratory GUI Testing"),
		seedPaper("75984cb2-73ab-5ea7-b4e9-64420ae04ffc", "Investigating the Robustness of Locators in Template-Based Web Application Testing Using a GUI Change Classification Model"),
		seedPaper("d215ce73-b4e8-51ad-be9c-189b70f77206", "Estimating Return on Investment for GUI Test Automation Frameworks"),
		seedPaper("4d0cccb3-040a-5322-a540-88d4827463cd", "Understanding Flaky Tests: The Developer’s Perspective"),
		seedPaper("1dd3a60a-5612-5075-8c83-55c62c136daa", "A Systematic Review on Regression Test Selection Techniques"),
		seedPaper("e7e9d846-029e-5c31-a18c-575dcb935ecc", "Testers’ Experiences of Tools and Automation"),
		seedPaper("6e1a2895-2458-52fa-b75d-e63fff04a738", "Test Tools: An Illusion of Usability?"),
		seedPaper("5966a651-6ed5-5bbc-bfe0-c76331d373a6", "Automating Acceptance Tests for GUI Applications in an Extreme Programming Environment"),
		seedPaper("5f14ebf6-ecd9-573e-a865-4a6ef307b542", "Guidelines for GUI Testing Maintenance: A Linter for Test Smell Detection"),
		seedPaper("c491a607-c3a5-5672-9fe5-1887337a9302", "Exploring Browser Automation: A Comparative Study of Selenium, Cypress, Puppeteer, and Playwright"),
		seedPaper("0b947dbd-02b1-5915-a97b-0934a78d0d8e", "Developing, Verifying, and Maintaining High-Quality Automated Test Scripts"),
		seedPaper("33ca1920-46c1-55c9-aa1f-99e7e13673e3", "Smells in Software Test Code: A Survey of Knowledge in Industry and Academia"),
		seedPaper("e5a22cb0-8bf1-53c1-8643-f2e70df95cd6", "A Survey on Software Testability"),
		seedPaper("bb6700d6-fd29-531d-b982-cba478879e8e", "Visual GUI Testing in Practice: An Extended Industrial Case Study"),
		seedPaper("b7e17ad6-53d4-5706-800e-24ee38e4f8e6", "Analysis and Design of Selenium WebDriver Automation Testing Framework"),
		seedPaper("6c7de263-dc9a-5486-b961-9944b4618a1c", "Creating GUI Testing Tools Using Accessibility Technologies"),
		seedPaper("a266da52-ef13-5c36-8df7-ad18092d200d", "Maintaining and Evolving GUI-directed Test Scripts"),
		seedPaper("d5449c2a-84c3-5dbb-b5bb-3f5b904c6ece", "How Can Manual Testing Processes Be Optimized? Developer Survey, Optimization Guidelines, and Case Studies"),
		seedPaper("24f4be49-54ea-5270-bbdc-2ba9dbb2b9ec", "Prioritizing Manual Test Cases in Traditional and Rapid Release Environments"),
		seedPaper("6b38853d-3db8-53e0-bfa5-1ddf97fae12a", "Exploratory Testing: A Multiple Case Study"),
		seedPaper("05d0f513-1c1f-57d2-ba23-08e090839413", "How Do Testers Do It? An Exploratory Study on Manual Testing Practices"),
		seedPaper("92734ac1-4eff-552e-baf3-fa3c9d0c7b20", "The Role of the Tester's Knowledge in Exploratory Software Testing"),
		seedPaper("acb6f6f9-72d7-55cd-9d42-f9042df36872", "Empirical Observations on Software Testing Automation"),
		seedPaper("1efc568d-7748-5911-a3cb-6e3e6a847355", "Inspecting Automated Test Code: A Preliminary Study"),
		seedPaper("80220a79-6cb6-5398-a384-72083db02abb", "Effects of Developer Experience on Learning and Applying Unit Test-Driven Development"),
		seedPaper("84b2b6b8-db75-5255-868e-14d4fa954059", "The Influence of Size and Coverage on Test Suite Effectiveness"),
		seedPaper("c955e641-dd73-5322-9662-48c3ffc4cb33", "Augmented Testing: Industry Feedback To Shape a New Testing Technology"),
		seedPaper("0b689e5d-6963-594a-a9e1-bd6436ccb0d2", "On the Industrial Applicability of Augmented Testing: An Empirical Study"),
		seedPaper("f64fad0e-534c-5f1c-b1eb-99ba3d57f35c", "Why Many Challenges with GUI Test Automation (Will) Remain"),
		seedPaper("a82c9c22-8203-584d-a606-8dda113b0eee", "Similarity-Based Web Element Localization for Robust Test Automation"),
		seedPaper("fb66d480-ac7f-56b8-9837-6b9b0087b458", "The Art of Software Testing"),
		seedPaper("61f1c112-3068-58ab-bccb-45d89bf42717", "Software Testing and Analysis: Process, Principles and Techniques"),
		seedPaper("6ba7cb92-7954-5cf1-87c1-c4ebcb0c92cf", "Impediments for Software Test Automation: A Systematic Literature Review: Impediments for Software Test Automation"),
		seedPaper("514bd880-abd6-5656-8c7e-69d05e29d1ec", "An Empirical Study of Bugs in Test Code"),
	}
}

func secondaryStudiesPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("fee26da6-4f03-5665-905e-27bfc1815803", "Shades of Grey: Guidelines for Working with the Grey Literature in Systematic Reviews for Management and Organizational Studies: Shades of Grey"),
		seedPaper("9b3aff9a-61de-5c35-870b-7eb7959c51d5", "Crowdsourced Software Testing: A Systematic Literature Review"),
		seedPaper("bce3eac8-0156-5518-83e0-753f47ec5e12", "Guidelines for Managing Threats to Validity of Secondary Studies in Software Engineering"),
		seedPaper("ea0fde12-7e32-5cec-869d-8f61ea8e5a74", "Graphical User Interface (GUI) Testing: Systematic Mapping and Repository"),
		seedPaper("8c0c603e-1c5f-5ccc-9cdf-12ddf2072a4b", "A Taxonomy of Metrics for GUI-based Testing Research: A Systematic Literature Review"),
		seedPaper("8fd47375-9722-55f3-8196-939353d0f592", "A Systematic Literature Review and Taxonomy of Modern Code Review"),
		seedPaper("1af38bde-d295-5515-9dc6-ebadb11c54dc", "The Need for Multivocal Literature Reviews in Software Engineering: Complementing Systematic Literature Reviews with Grey Literature"),
		seedPaper("35177f2b-816a-5611-a55a-41083843c3fb", "A Systematic Literature Review of Literature Reviews in Software Testing"),
		seedPaper("a900e357-a3b9-5baf-8202-f2b40f920855", "When and What to Automate in Software Testing? A Multi-Vocal Literature Review"),
		seedPaper("36ad41cb-88ba-562e-b656-5c9a4973ee54", "Guidelines for Including Grey Literature and Conducting Multivocal Literature Reviews in Software Engineering"),
		seedPaper("6df5173d-7e06-5c4d-adc3-abeba3daa3f5", "Guidelines for Performing Systematic Literature Reviews in Software Engineering"),
		seedPaper("9497413f-ad82-5e99-9225-18eb6f8f33c9", "On the Road to Interactive LLM-based Systematic Mapping Studies"),
		seedPaper("21eecc20-d375-5049-9006-eb85adac4aa1", "Reference-Based Search Strategies in Systematic Reviews"),
		seedPaper("c29cd7e8-984b-5fd8-9d93-da8fc0f0fcb9", "Taxonomies in Software Engineering: A Systematic Mapping Study and a Revised Taxonomy Development Method"),
		seedPaper("7a061311-b9e0-50ff-a39c-c4f9e5395d51", "A Systematic Process for Mining Software Repositories: Results from a Systematic Literature Review"),
		seedPaper("1d7be506-3050-52f6-92f1-9a01c0caee49", "On the Reliability of Mapping Studies in Software Engineering"),
		seedPaper("0837a72b-a72c-5c37-9152-462d91ced844", "Guidelines for Snowballing in Systematic Literature Studies and a Replication in Software Engineering"),
		seedPaper("d1af8ede-bad4-580d-b342-e1ac4e26f9de", "A Map of Threats to Validity of Systematic Literature Reviews in Software Engineering"),
	}
}

func bayesianPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("8ca3b8ca-50f0-536a-8a75-9bef2033f39b", "Using Causal Inference and Bayesian Statistics to Explain the Capability of a Test Suite in Exposing Software Faults"),
		seedPaper("9d6cb335-5a27-5b62-bbdc-c1ca5087393b", "Statisticians Issue Warning over Misuse of P Values"),
		seedPaper("88874208-6f1d-520c-929a-311a3fcc7338", "Brms: An R Package for Bayesian Multilevel Models Using Stan"),
		seedPaper("68bddc7e-fe18-5e60-9078-ed16a07f6e0a", "Hierarchical Bayesian Analysis of the Carryover Effect in Two-Period Crossover Designs"),
		seedPaper("00007003-e8c8-51f9-910d-d81ce21063a3", "Bayesian Data Analysis in Empirical Software Engineering Research"),
		seedPaper("49293981-c8e5-5bcc-b7c9-1474f3ef99b6", "Applying Bayesian Analysis Guidelines to Empirical Software Engineering Data: The Case of Programming Languages and Code Quality"),
		seedPaper("b63b8503-dea4-5ee7-9709-26a4e0e54999", "Bayesian Workflow"),
		seedPaper("2397b703-67d8-5340-b243-0adb21928961", "The Insignificance of Null Hypothesis Significance Testing"),
		seedPaper("8f1fb3cc-6df8-53ae-b322-e0624da68f26", "Practical Bayesian Model Evaluation Using Leave-One-out Cross-Validation and WAIC"),
	}
}

func researchMethodologyPapers() []paperDomain.Paper {
	return []paperDomain.Paper{
		seedPaper("050e4957-1948-5e6b-bc93-e1fa5045879d", "Distinguishing between Method and Methodology in Academic Research"),
		seedPaper("cef605e1-cafe-5d29-bb9d-e242400c53c9", "Member Checking: A Tool to Enhance Trustworthiness or Merely a Nod to Validation?"),
		seedPaper("1d84292b-a152-5f43-ba9c-4850e35227a8", "Using Thematic Analysis in Psychology"),
		seedPaper("734c9f00-2910-54c6-a739-a33be104cd90", "Successful Qualitative Research: A Practical Guide for Beginners"),
		seedPaper("1d01ab6a-d6b7-54d6-8541-820148131735", "Contemporary Empirical Methods in Software Engineering"),
		seedPaper("c81901b5-8c0b-58e9-8b10-4ec084f63e20", "Triangulation: Establishing the Validity of Qualitative Studies"),
		seedPaper("952e5fac-40ff-567f-887b-578e26149744", "Design Science Research in Information Systems"),
		seedPaper("4e3143ce-e7f2-5029-8d63-7caa957c0cee", "Experiences from Conducting Semi-structured Interviews in Empirical Software Engineering Research"),
		seedPaper("3523796c-63a9-5c5d-bb1a-581747689718", "A Method for Evaluating Rigor and Industrial Relevance of Technology Evaluations"),
		seedPaper("1fa1a590-5f0d-529e-a6ec-5175a769fca4", "The Logic of Qualitative Survey Research and Its Position in the Field of Social Research Methods"),
		seedPaper("a4e72361-b5a2-5a8f-a1df-ebf123e3a8d7", "Reporting Experiments in Software Engineering"),
		seedPaper("26d6d631-c853-5959-8bde-bb61b7481155", "A Practical Guide to Controlled Experiments of Software Engineering Tools with Human Participants"),
		seedPaper("c5847dee-8359-56cf-a2b7-efd5eb7c5226", "Qualitative Research Design: An Interactive Approach"),
		seedPaper("02368094-41ae-5b4b-8c86-d5adc81e411e", "Reliability and Inter-rater Reliability in Qualitative Research: Norms and Guidelines for CSCW and HCI Practice"),
		seedPaper("653cf6dd-671d-5d2f-90c9-727998a68901", "Empirical Software Engineering: From Discipline to Interdiscipline"),
		seedPaper("0f88b274-a1f9-5b30-9b95-b9e290551301", "Qualitative Research: A Guide to Design and Implementation"),
		seedPaper("23d58066-49f8-5a28-9ad0-bc0df5d57038", "The Qualitative Interview in IS Research: Examining the Craft"),
		seedPaper("8307cec1-379c-5ca5-99a8-25037e7c4db6", "Intercoder Reliability in Qualitative Research: Debates and Practical Guidelines"),
		seedPaper("d58e17b1-bc50-57f7-a535-9ce39fb89231", "Real World Research: A Resource for Users of Social Research Methods in Applied Settings"),
		seedPaper("4aebf490-f7c0-55ee-8def-5c2e34cb21c8", "Guidelines for Conducting and Reporting Case Study Research in Software Engineering"),
		seedPaper("da710511-65d1-581d-b5da-58bd63939a29", "Qualitative Methods in Empirical Studies of Software Engineering"),
		seedPaper("79a104e5-207e-5549-8447-ab59bc57cb23", "Building Theories in Software Engineering"),
		seedPaper("cb9a80de-2823-5795-9bbe-eece204683db", "Construct Validity in Software Engineering"),
		seedPaper("65edf0ea-cf45-533d-a0e4-82a56b408cb2", "Action Research in Software Engineering: Theory and Applications"),
		seedPaper("223fe559-0f75-51b9-b0e0-c72f3ad2aeca", "The ABC of Software Engineering Research"),
		seedPaper("4aaa11c1-71cb-530c-b43a-8b97bd1cc284", "Introduction to Qualitative Research Methods: A Guidebook and Resource"),
		seedPaper("b47d7b79-d586-55fa-b4c1-d1ad856513d0", "Straussian Grounded-Theory Method: An Illustration"),
		seedPaper("452837c8-bed2-5282-99ca-c8bbf3888fd0", "Qualitative Quality: Eight “Big-Tent” Criteria for Excellent Qualitative Research"),
		seedPaper("11d26b29-8b5d-5da0-b683-7c1d89c58a89", "Experimentation in Software Engineering"),
		seedPaper("ec3b4b79-2516-5d28-a316-1538232efcb8", "Case Study Research in Software Engineering—It Is a Case, and It Is a Study, but Is It a Case Study?"),
		seedPaper("d16f6c97-34c1-56c5-a618-abeb8dcb616e", "Guiding the Selection of Research Methodology in Industry–Academia Collaboration in Software Engineering"),
		seedPaper("df2367ff-f0da-51a1-98ca-31de203cc092", "Case Study Research and Applications: Design and Methods"),
	}
}
