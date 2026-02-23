package main

// evalQuery defines a test activation query and expected relevance criteria.
type evalQuery struct {
	Context      []string // query context strings
	Domain       string   // domain label for reporting
	ExpectedTags []string // terms that should appear in relevant results
	Description  string   // what this query tests
}

// evalQueries is the comprehensive set of test queries spanning all knowledge domains.
// Each query tests a different type of retrieval: precision, association, cross-domain.
var evalQueries = []evalQuery{
	// ── Physics ──────────────────────────────────────────────────────────────
	{
		Context:      []string{"quantum entanglement spooky action distance"},
		Domain:       "Physics",
		ExpectedTags: []string{"quantum", "entanglement", "particle", "wave", "superposition", "uncertainty", "photon", "spin"},
		Description:  "Quantum mechanics concepts",
	},
	{
		Context:      []string{"special relativity time dilation speed of light"},
		Domain:       "Physics",
		ExpectedTags: []string{"relativity", "Einstein", "light", "time", "mass", "energy", "spacetime", "Lorentz"},
		Description:  "Relativity concepts",
	},
	{
		Context:      []string{"thermodynamics entropy heat engine Carnot"},
		Domain:       "Physics",
		ExpectedTags: []string{"thermodynamics", "entropy", "heat", "temperature", "energy", "Carnot", "Kelvin"},
		Description:  "Thermodynamics",
	},
	{
		Context:      []string{"black hole event horizon singularity Hawking radiation"},
		Domain:       "Physics / Cosmology",
		ExpectedTags: []string{"black hole", "gravity", "relativity", "Hawking", "singularity", "horizon", "star"},
		Description:  "Black holes and gravity",
	},
	{
		Context:      []string{"electromagnetic wave Maxwell equations light electricity magnetism"},
		Domain:       "Physics",
		ExpectedTags: []string{"electromagnetism", "Maxwell", "wave", "light", "field", "charge", "electric", "magnetic"},
		Description:  "Electromagnetism",
	},
	{
		Context:      []string{"Higgs boson standard model particle physics LHC"},
		Domain:       "Particle Physics",
		ExpectedTags: []string{"Higgs", "boson", "particle", "standard model", "quark", "lepton", "LHC", "force"},
		Description:  "Particle physics",
	},

	// ── Biology & Neuroscience ────────────────────────────────────────────────
	{
		Context:      []string{"how neurons communicate synaptic transmission neurotransmitters"},
		Domain:       "Neuroscience",
		ExpectedTags: []string{"neuron", "synapse", "neurotransmitter", "brain", "signal", "action potential", "dendrite", "axon"},
		Description:  "Neural communication",
	},
	{
		Context:      []string{"DNA replication transcription protein synthesis genetic code"},
		Domain:       "Molecular Biology",
		ExpectedTags: []string{"DNA", "RNA", "protein", "gene", "replication", "transcription", "translation", "codon"},
		Description:  "Central dogma of molecular biology",
	},
	{
		Context:      []string{"CRISPR gene editing genetic engineering biotechnology"},
		Domain:       "Biotechnology",
		ExpectedTags: []string{"CRISPR", "gene", "DNA", "editing", "Cas9", "genetic", "therapy", "mutation"},
		Description:  "Gene editing technology",
	},
	{
		Context:      []string{"natural selection evolution Darwin adaptation species"},
		Domain:       "Evolution",
		ExpectedTags: []string{"evolution", "Darwin", "natural selection", "species", "adaptation", "fitness", "mutation", "genetics"},
		Description:  "Evolutionary biology",
	},
	{
		Context:      []string{"memory consolidation hippocampus long-term potentiation learning"},
		Domain:       "Neuroscience",
		ExpectedTags: []string{"memory", "hippocampus", "learning", "synapse", "neuron", "LTP", "brain", "consolidation"},
		Description:  "Memory and learning neuroscience",
	},
	{
		Context:      []string{"immune system T cells antibodies infection defense"},
		Domain:       "Immunology",
		ExpectedTags: []string{"immune", "antibody", "T cell", "B cell", "infection", "antigen", "lymphocyte", "vaccine"},
		Description:  "Immune system",
	},
	{
		Context:      []string{"cell division mitosis meiosis chromosome"},
		Domain:       "Cell Biology",
		ExpectedTags: []string{"cell", "division", "mitosis", "meiosis", "chromosome", "DNA", "nucleus", "reproduction"},
		Description:  "Cell division",
	},
	{
		Context:      []string{"photosynthesis chlorophyll light energy plant glucose"},
		Domain:       "Biology",
		ExpectedTags: []string{"photosynthesis", "chlorophyll", "plant", "light", "glucose", "carbon dioxide", "oxygen"},
		Description:  "Photosynthesis",
	},

	// ── Computer Science & AI ─────────────────────────────────────────────────
	{
		Context:      []string{"machine learning neural network gradient descent backpropagation"},
		Domain:       "AI/ML",
		ExpectedTags: []string{"neural network", "machine learning", "gradient", "backpropagation", "deep learning", "training", "weights"},
		Description:  "Deep learning training",
	},
	{
		Context:      []string{"transformer attention mechanism natural language processing BERT GPT"},
		Domain:       "AI/NLP",
		ExpectedTags: []string{"transformer", "attention", "language model", "NLP", "BERT", "GPT", "token", "embedding"},
		Description:  "Transformer models",
	},
	{
		Context:      []string{"graph algorithms shortest path Dijkstra BFS DFS"},
		Domain:       "Algorithms",
		ExpectedTags: []string{"graph", "algorithm", "Dijkstra", "BFS", "DFS", "path", "vertex", "edge", "tree"},
		Description:  "Graph traversal algorithms",
	},
	{
		Context:      []string{"sorting algorithm complexity quicksort merge sort big O notation"},
		Domain:       "Algorithms",
		ExpectedTags: []string{"sort", "algorithm", "complexity", "quicksort", "merge", "big O", "time complexity", "O(n log n)"},
		Description:  "Sorting algorithms",
	},
	{
		Context:      []string{"database indexing B-tree query optimization SQL"},
		Domain:       "Databases",
		ExpectedTags: []string{"database", "index", "B-tree", "SQL", "query", "optimization", "table", "key"},
		Description:  "Database internals",
	},
	{
		Context:      []string{"distributed systems consensus Raft Paxos fault tolerance"},
		Domain:       "Distributed Systems",
		ExpectedTags: []string{"distributed", "consensus", "Raft", "Paxos", "fault", "leader", "replication", "nodes"},
		Description:  "Distributed consensus",
	},
	{
		Context:      []string{"operating system process thread memory management virtual"},
		Domain:       "Systems",
		ExpectedTags: []string{"operating system", "process", "thread", "memory", "scheduler", "kernel", "virtual", "CPU"},
		Description:  "OS concepts",
	},
	{
		Context:      []string{"cryptography encryption RSA public key AES"},
		Domain:       "Security",
		ExpectedTags: []string{"cryptography", "encryption", "RSA", "AES", "public key", "cipher", "hash", "security"},
		Description:  "Cryptography",
	},
	{
		Context:      []string{"reinforcement learning reward policy agent environment"},
		Domain:       "AI/ML",
		ExpectedTags: []string{"reinforcement learning", "reward", "policy", "agent", "Q-learning", "Markov", "value function"},
		Description:  "Reinforcement learning",
	},

	// ── Mathematics ──────────────────────────────────────────────────────────
	{
		Context:      []string{"Riemann hypothesis prime numbers zeta function number theory"},
		Domain:       "Mathematics",
		ExpectedTags: []string{"Riemann", "prime", "zeta", "number theory", "complex", "hypothesis", "mathematics"},
		Description:  "Number theory",
	},
	{
		Context:      []string{"Fourier transform frequency domain signal processing"},
		Domain:       "Mathematics",
		ExpectedTags: []string{"Fourier", "transform", "frequency", "signal", "wave", "harmonic", "spectrum"},
		Description:  "Fourier analysis",
	},
	{
		Context:      []string{"topology manifold Euler characteristic homeomorphism"},
		Domain:       "Mathematics",
		ExpectedTags: []string{"topology", "manifold", "Euler", "homeomorphism", "surface", "space", "continuous"},
		Description:  "Topology",
	},
	{
		Context:      []string{"statistics Bayesian inference probability distribution"},
		Domain:       "Statistics",
		ExpectedTags: []string{"Bayesian", "probability", "statistics", "inference", "distribution", "prior", "posterior"},
		Description:  "Bayesian statistics",
	},

	// ── History ───────────────────────────────────────────────────────────────
	{
		Context:      []string{"Roman Empire fall decline Julius Caesar Augustus"},
		Domain:       "History",
		ExpectedTags: []string{"Roman", "Caesar", "Rome", "empire", "Augustus", "republic", "senate", "legion"},
		Description:  "Roman history",
	},
	{
		Context:      []string{"ancient Egypt pharaohs pyramids hieroglyphs Nile civilization"},
		Domain:       "History",
		ExpectedTags: []string{"Egypt", "pharaoh", "pyramid", "hieroglyph", "Nile", "ancient", "civilization", "tomb"},
		Description:  "Ancient Egypt",
	},
	{
		Context:      []string{"World War II Holocaust Nazi Germany Hitler Holocaust"},
		Domain:       "History",
		ExpectedTags: []string{"World War", "Nazi", "Hitler", "Holocaust", "Germany", "Europe", "Jews", "war"},
		Description:  "WWII history",
	},
	{
		Context:      []string{"French Revolution Enlightenment liberty equality Napoleon"},
		Domain:       "History",
		ExpectedTags: []string{"French Revolution", "Napoleon", "Enlightenment", "liberty", "equality", "republic", "monarchy"},
		Description:  "French Revolution",
	},
	{
		Context:      []string{"Industrial Revolution steam engine factory textile cotton"},
		Domain:       "History",
		ExpectedTags: []string{"industrial", "revolution", "steam", "factory", "engine", "coal", "Britain", "mechanization"},
		Description:  "Industrial Revolution",
	},
	{
		Context:      []string{"ancient Greece philosophy Athens democracy Socrates Plato Aristotle"},
		Domain:       "History",
		ExpectedTags: []string{"Greek", "Athens", "democracy", "Socrates", "Plato", "Aristotle", "philosophy", "polis"},
		Description:  "Ancient Greece",
	},
	{
		Context:      []string{"Byzantine Empire Constantinople medieval Justinian Orthodox"},
		Domain:       "History",
		ExpectedTags: []string{"Byzantine", "Constantinople", "Justinian", "Ottoman", "Orthodox", "medieval", "Roman"},
		Description:  "Byzantine history",
	},
	{
		Context:      []string{"Mongol Empire Genghis Khan conquest silk road trade"},
		Domain:       "History",
		ExpectedTags: []string{"Mongol", "Genghis Khan", "empire", "conquest", "silk road", "Asia", "trade"},
		Description:  "Mongol history",
	},

	// ── Philosophy ────────────────────────────────────────────────────────────
	{
		Context:      []string{"consciousness hard problem qualia subjective experience mind"},
		Domain:       "Philosophy",
		ExpectedTags: []string{"consciousness", "qualia", "mind", "Chalmers", "subjective", "experience", "philosophy", "brain"},
		Description:  "Philosophy of mind",
	},
	{
		Context:      []string{"Kant categorical imperative moral philosophy ethics duty"},
		Domain:       "Philosophy",
		ExpectedTags: []string{"Kant", "ethics", "categorical imperative", "moral", "duty", "philosophy", "reason"},
		Description:  "Kantian ethics",
	},
	{
		Context:      []string{"epistemology knowledge justified true belief Gettier"},
		Domain:       "Philosophy",
		ExpectedTags: []string{"epistemology", "knowledge", "belief", "truth", "justified", "Gettier", "philosophy"},
		Description:  "Epistemology",
	},
	{
		Context:      []string{"Plato forms cave allegory reality knowledge"},
		Domain:       "Philosophy",
		ExpectedTags: []string{"Plato", "forms", "cave", "allegory", "reality", "ideal", "knowledge", "Socrates"},
		Description:  "Platonic philosophy",
	},

	// ── Medicine & Psychology ─────────────────────────────────────────────────
	{
		Context:      []string{"Ebbinghaus forgetting curve memory retention learning spacing effect"},
		Domain:       "Psychology",
		ExpectedTags: []string{"Ebbinghaus", "memory", "forgetting", "learning", "retention", "spacing", "curve"},
		Description:  "Memory and forgetting (meta-test for a cognitive DB!)",
	},
	{
		Context:      []string{"cognitive behavioral therapy depression anxiety mental health"},
		Domain:       "Psychology",
		ExpectedTags: []string{"cognitive", "therapy", "depression", "anxiety", "mental health", "CBT", "behavior", "psychological"},
		Description:  "Mental health treatment",
	},
	{
		Context:      []string{"dopamine reward system addiction motivation neurotransmitter"},
		Domain:       "Neuroscience",
		ExpectedTags: []string{"dopamine", "reward", "addiction", "motivation", "neurotransmitter", "brain", "pleasure"},
		Description:  "Dopamine reward system",
	},
	{
		Context:      []string{"CRISPR cancer gene therapy personalized medicine tumor"},
		Domain:       "Medicine",
		ExpectedTags: []string{"cancer", "gene therapy", "tumor", "CRISPR", "medicine", "treatment", "genetic", "oncology"},
		Description:  "Cancer gene therapy",
	},
	{
		Context:      []string{"Pavlov classical conditioning behavioral psychology reflex"},
		Domain:       "Psychology",
		ExpectedTags: []string{"Pavlov", "conditioning", "reflex", "stimulus", "response", "psychology", "behavior", "learning"},
		Description:  "Classical conditioning",
	},

	// ── Arts & Literature ─────────────────────────────────────────────────────
	{
		Context:      []string{"Shakespeare hamlet tragedy tragedy soliloquy theater"},
		Domain:       "Literature",
		ExpectedTags: []string{"Shakespeare", "Hamlet", "tragedy", "theater", "play", "Elizabethan", "soliloquy"},
		Description:  "Shakespearean literature",
	},
	{
		Context:      []string{"Renaissance art Leonardo da Vinci Michelangelo perspective painting"},
		Domain:       "Arts",
		ExpectedTags: []string{"Renaissance", "Leonardo", "Michelangelo", "art", "perspective", "Florence", "painting"},
		Description:  "Renaissance art",
	},
	{
		Context:      []string{"Beethoven symphony classical music composition orchestra"},
		Domain:       "Music",
		ExpectedTags: []string{"Beethoven", "symphony", "classical", "music", "composer", "orchestra", "sonata"},
		Description:  "Classical music",
	},

	// ── Technology ────────────────────────────────────────────────────────────
	{
		Context:      []string{"TCP/IP networking internet protocol OSI model packets"},
		Domain:       "Networking",
		ExpectedTags: []string{"TCP", "IP", "network", "internet", "protocol", "packet", "OSI", "routing"},
		Description:  "Network protocols",
	},
	{
		Context:      []string{"Linux kernel Unix POSIX operating system open source"},
		Domain:       "Technology",
		ExpectedTags: []string{"Linux", "Unix", "kernel", "operating system", "open source", "POSIX", "shell"},
		Description:  "Linux/Unix",
	},
	{
		Context:      []string{"blockchain Bitcoin Ethereum cryptocurrency decentralized ledger"},
		Domain:       "Technology",
		ExpectedTags: []string{"blockchain", "Bitcoin", "Ethereum", "cryptocurrency", "decentralized", "ledger", "hash"},
		Description:  "Blockchain and crypto",
	},
	{
		Context:      []string{"cloud computing AWS microservices containers Docker Kubernetes"},
		Domain:       "Technology",
		ExpectedTags: []string{"cloud", "AWS", "microservices", "container", "Docker", "Kubernetes", "serverless"},
		Description:  "Cloud and containerization",
	},

	// ── Astronomy & Cosmology ─────────────────────────────────────────────────
	{
		Context:      []string{"Big Bang cosmic inflation early universe dark matter dark energy"},
		Domain:       "Cosmology",
		ExpectedTags: []string{"Big Bang", "universe", "inflation", "dark matter", "dark energy", "cosmic", "expansion"},
		Description:  "Cosmology and universe origin",
	},
	{
		Context:      []string{"stellar evolution star formation nebula supernova neutron star"},
		Domain:       "Astronomy",
		ExpectedTags: []string{"star", "stellar", "nebula", "supernova", "neutron star", "evolution", "nuclear fusion"},
		Description:  "Stellar evolution",
	},
	{
		Context:      []string{"exoplanet Kepler habitable zone search for life Mars"},
		Domain:       "Astronomy",
		ExpectedTags: []string{"exoplanet", "Kepler", "habitable zone", "Mars", "life", "planet", "orbit", "telescope"},
		Description:  "Search for extraterrestrial life",
	},

	// ── Cross-domain associative queries ──────────────────────────────────────
	{
		Context:      []string{"pattern recognition associative memory brain neural network"},
		Domain:       "Cross-domain",
		ExpectedTags: []string{"pattern", "recognition", "memory", "neural", "brain", "network", "association", "learning"},
		Description:  "Cross-domain: neuroscience + AI",
	},
	{
		Context:      []string{"information entropy Shannon bits complexity"},
		Domain:       "Cross-domain",
		ExpectedTags: []string{"entropy", "Shannon", "information", "bits", "complexity", "thermodynamics", "communication"},
		Description:  "Cross-domain: physics + information theory",
	},
	{
		Context:      []string{"game theory Nash equilibrium strategy decision cooperation"},
		Domain:       "Cross-domain",
		ExpectedTags: []string{"game theory", "Nash", "equilibrium", "strategy", "decision", "cooperation", "prisoner"},
		Description:  "Cross-domain: math + economics + philosophy",
	},
	{
		Context:      []string{"fractals self-similarity chaos theory Mandelbrot"},
		Domain:       "Cross-domain",
		ExpectedTags: []string{"fractal", "chaos", "Mandelbrot", "self-similarity", "iteration", "complex", "dimension"},
		Description:  "Cross-domain: math + physics + nature",
	},
	{
		Context:      []string{"epigenetics gene expression environment inheritance"},
		Domain:       "Cross-domain",
		ExpectedTags: []string{"epigenetic", "gene", "expression", "environment", "inheritance", "methylation", "histone"},
		Description:  "Cross-domain: genetics + environment",
	},
	{
		Context:      []string{"emergence complex systems ant colony swarm intelligence"},
		Domain:       "Cross-domain",
		ExpectedTags: []string{"emergence", "complex", "swarm", "ant", "self-organization", "collective", "intelligence"},
		Description:  "Cross-domain: complexity theory + biology + AI",
	},
}
