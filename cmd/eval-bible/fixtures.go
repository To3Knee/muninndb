package main

import "time"

// fixture is a curated verse with hand-set cognitive state for Phase 2.
type fixture struct {
	Concept     string
	Content     string
	Tags        []string
	DaysAgo     int     // how long ago this was "last read"
	AccessCount uint32  // how many times it was "accessed"
	Stability   float32 // decay resistance
}

// fixtureGroup groups fixtures by reading period.
type fixtureGroup struct {
	Name     string
	Fixtures []fixture
}

// phase2Fixtures is the curated corpus for Phase 2. It simulates a reader with
// a decade of Bible study: older books heavily decayed, recent books fresh.
var phase2Fixtures = []fixtureGroup{
	{
		Name: "torah",
		Fixtures: []fixture{
			{Concept: "Genesis 1:1", Content: "In the beginning God created the heaven and the earth.", Tags: []string{"Old Testament", "Genesis", "law"}, DaysAgo: 3650, AccessCount: 2, Stability: 10},
			{Concept: "Genesis 3:15", Content: "And I will put enmity between thee and the woman, and between thy seed and her seed; it shall bruise thy head, and thou shalt bruise his heel.", Tags: []string{"Old Testament", "Genesis", "law"}, DaysAgo: 3650, AccessCount: 1, Stability: 10},
			{Concept: "Exodus 20:3", Content: "Thou shalt have no other gods before me.", Tags: []string{"Old Testament", "Exodus", "law"}, DaysAgo: 3650, AccessCount: 1, Stability: 10},
			{Concept: "Deuteronomy 6:4", Content: "Hear, O Israel: The LORD our God is one LORD.", Tags: []string{"Old Testament", "Deuteronomy", "law"}, DaysAgo: 3650, AccessCount: 2, Stability: 10},
		},
	},
	{
		Name: "psalms_proverbs",
		Fixtures: []fixture{
			{Concept: "Psalm 23:1", Content: "The LORD is my shepherd; I shall not want.", Tags: []string{"Old Testament", "Psalms", "poetry"}, DaysAgo: 90, AccessCount: 8, Stability: 30},
			{Concept: "Psalm 23:4", Content: "Yea, though I walk through the valley of the shadow of death, I will fear no evil: for thou art with me; thy rod and thy staff they comfort me.", Tags: []string{"Old Testament", "Psalms", "poetry"}, DaysAgo: 90, AccessCount: 6, Stability: 30},
			{Concept: "Psalm 51:10", Content: "Create in me a clean heart, O God; and renew a right spirit within me.", Tags: []string{"Old Testament", "Psalms", "poetry"}, DaysAgo: 2555, AccessCount: 3, Stability: 10},
			{Concept: "Proverbs 3:5", Content: "Trust in the LORD with all thine heart; and lean not unto thine own understanding.", Tags: []string{"Old Testament", "Proverbs", "wisdom"}, DaysAgo: 2555, AccessCount: 2, Stability: 10},
		},
	},
	{
		Name: "prophets",
		Fixtures: []fixture{
			{Concept: "Isaiah 53:5", Content: "But he was wounded for our transgressions, he was bruised for our iniquities: the chastisement of our peace was upon him; and with his stripes we are healed.", Tags: []string{"Old Testament", "Isaiah", "prophecy"}, DaysAgo: 2190, AccessCount: 2, Stability: 10},
			{Concept: "Isaiah 7:14", Content: "Therefore the Lord himself shall give you a sign; Behold, a virgin shall conceive, and bear a son, and shall call his name Immanuel.", Tags: []string{"Old Testament", "Isaiah", "prophecy"}, DaysAgo: 2190, AccessCount: 1, Stability: 10},
			{Concept: "Jeremiah 29:11", Content: "For I know the thoughts that I think toward you, saith the LORD, thoughts of peace, and not of evil, to give you an expected end.", Tags: []string{"Old Testament", "Jeremiah", "prophecy"}, DaysAgo: 2190, AccessCount: 2, Stability: 10},
		},
	},
	{
		Name: "gospels",
		Fixtures: []fixture{
			{Concept: "John 1:1", Content: "In the beginning was the Word, and the Word was with God, and the Word was God.", Tags: []string{"New Testament", "John", "gospel"}, DaysAgo: 730, AccessCount: 7, Stability: 30},
			{Concept: "John 3:16", Content: "For God so loved the world, that he gave his only begotten Son, that whosoever believeth in him should not perish, but have everlasting life.", Tags: []string{"New Testament", "John", "gospel"}, DaysAgo: 730, AccessCount: 8, Stability: 30},
			{Concept: "John 10:11", Content: "I am the good shepherd: the good shepherd giveth his life for the sheep.", Tags: []string{"New Testament", "John", "gospel"}, DaysAgo: 730, AccessCount: 5, Stability: 30},
			{Concept: "John 11:35", Content: "Jesus wept.", Tags: []string{"New Testament", "John", "gospel"}, DaysAgo: 730, AccessCount: 5, Stability: 30},
			{Concept: "Matthew 1:23", Content: "Behold, a virgin shall be with child, and shall bring forth a son, and they shall call his name Emmanuel, which being interpreted is, God with us.", Tags: []string{"New Testament", "Matthew", "gospel"}, DaysAgo: 730, AccessCount: 4, Stability: 30},
			{Concept: "Luke 23:34", Content: "Then said Jesus, Father, forgive them; for they know not what they do. And they parted his raiment, and cast lots.", Tags: []string{"New Testament", "Luke", "gospel"}, DaysAgo: 730, AccessCount: 5, Stability: 30},
		},
	},
	{
		Name: "epistles",
		Fixtures: []fixture{
			{Concept: "Romans 8:28", Content: "And we know that all things work together for good to them that love God, to them who are the called according to his purpose.", Tags: []string{"New Testament", "Romans", "epistle"}, DaysAgo: 365, AccessCount: 6, Stability: 30},
			{Concept: "Romans 6:23", Content: "For the wages of sin is death; but the gift of God is eternal life through Jesus Christ our Lord.", Tags: []string{"New Testament", "Romans", "epistle"}, DaysAgo: 365, AccessCount: 5, Stability: 30},
			{Concept: "1 Corinthians 13:4", Content: "Charity suffereth long, and is kind; charity envieth not; charity vaunteth not itself, is not puffed up.", Tags: []string{"New Testament", "1 Corinthians", "epistle"}, DaysAgo: 365, AccessCount: 5, Stability: 30},
			{Concept: "Ephesians 2:8", Content: "For by grace are ye saved through faith; and that not of yourselves: it is the gift of God.", Tags: []string{"New Testament", "Ephesians", "epistle"}, DaysAgo: 365, AccessCount: 4, Stability: 30},
			{Concept: "1 Peter 2:24", Content: "Who his own self bare our sins in his own body on the tree, that we, being dead to sins, should live unto righteousness: by whose stripes ye were healed.", Tags: []string{"New Testament", "1 Peter", "epistle"}, DaysAgo: 365, AccessCount: 4, Stability: 30},
			{Concept: "Philippians 4:13", Content: "I can do all things through Christ which strengtheneth me.", Tags: []string{"New Testament", "Philippians", "epistle"}, DaysAgo: 365, AccessCount: 6, Stability: 30},
		},
	},
	{
		Name: "revelation",
		Fixtures: []fixture{
			{Concept: "Revelation 12:9", Content: "And the great dragon was cast out, that old serpent, called the Devil, and Satan, which deceiveth the whole world: he was cast out into the earth, and his angels were cast out with him.", Tags: []string{"New Testament", "Revelation", "prophecy"}, DaysAgo: 90, AccessCount: 3, Stability: 15},
			{Concept: "Revelation 21:4", Content: "And God shall wipe away all tears from their eyes; and there shall be no more death, neither sorrow, nor crying, neither shall there be any more pain: for the former things are passed away.", Tags: []string{"New Testament", "Revelation", "prophecy"}, DaysAgo: 90, AccessCount: 3, Stability: 15},
		},
	},
}

// hebbianPairs are cross-testament pairs to co-activate at corpus load time,
// simulating a reader who explicitly connected these passages.
// OT verse is stale; NT verse is fresh. The Hebbian link should pull the stale
// OT verse up despite its age.
var hebbianPairs = [][2]string{
	{"Genesis 1:1", "John 1:1"},        // Creation / Logos
	{"Isaiah 53:5", "1 Peter 2:24"},    // Suffering servant / fulfillment
	{"Psalm 23:1", "John 10:11"},       // Shepherd
	{"Genesis 3:15", "Revelation 12:9"}, // Serpent defeated
	{"Isaiah 7:14", "Matthew 1:23"},    // Virgin birth / fulfillment
}

// anchorQueries are 6 queries used in Sub-experiment A.
// For each query, we know which loaded fixtures are semantically relevant.
// freshRefs = NT fixtures (DaysAgo ≤ 730); staleRefs = OT fixtures (DaysAgo ≥ 2190).
type anchorQuery struct {
	Context   string
	Label     string
	FreshRefs []string // NT verse concepts expected in results
	StaleRefs []string // OT verse concepts expected in results
}

var anchorQueries = []anchorQuery{
	{
		Context:   "forgiveness sins healing redemption suffering",
		Label:     "forgiveness",
		FreshRefs: []string{"1 Peter 2:24", "Luke 23:34", "Romans 6:23"},
		StaleRefs: []string{"Isaiah 53:5", "Psalm 51:10"},
	},
	{
		Context:   "shepherd flock sheep pasture protection",
		Label:     "shepherd",
		FreshRefs: []string{"John 10:11"},
		StaleRefs: []string{"Psalm 23:1", "Psalm 23:4"},
	},
	{
		Context:   "creation beginning word God earth heaven",
		Label:     "creation",
		FreshRefs: []string{"John 1:1"},
		StaleRefs: []string{"Genesis 1:1"},
	},
	{
		Context:   "eternal life salvation faith grace gift God",
		Label:     "salvation",
		FreshRefs: []string{"John 3:16", "Romans 6:23", "Ephesians 2:8"},
		StaleRefs: []string{"Isaiah 53:5"},
	},
	{
		Context:   "serpent enemy defeat victory defeated cast out",
		Label:     "serpent defeated",
		FreshRefs: []string{"Revelation 12:9"},
		StaleRefs: []string{"Genesis 3:15"},
	},
	{
		Context:   "virgin birth child son Emmanuel prophecy fulfilled",
		Label:     "virgin birth",
		FreshRefs: []string{"Matthew 1:23"},
		StaleRefs: []string{"Isaiah 7:14"},
	},
}

// daysAgoToTime converts a DaysAgo value to an absolute time.Time.
func daysAgoToTime(daysAgo int) time.Time {
	return time.Now().UTC().Add(-time.Duration(daysAgo) * 24 * time.Hour)
}
