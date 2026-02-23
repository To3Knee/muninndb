package main

import (
	"fmt"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// largeScaleTopics generates a large volume of topically-differentiated entries
// by applying multiple analytical perspectives to broad topic lists.
// Each topic generates 4 entries: overview, historical, application, research.
func largeScaleTopics() []mbp.WriteRequest {
	var all []mbp.WriteRequest
	all = append(all, topicMatrix("physics", physicsTopicList())...)
	all = append(all, topicMatrix("biology", biologyTopicList())...)
	all = append(all, topicMatrix("computer science", csTopicList())...)
	all = append(all, topicMatrix("mathematics", mathTopicList())...)
	all = append(all, topicMatrix("history", historyTopicList())...)
	all = append(all, topicMatrix("philosophy", philosophyTopicList())...)
	all = append(all, topicMatrix("neuroscience", neuroscienceTopicList())...)
	all = append(all, topicMatrix("chemistry", chemTopicList())...)
	all = append(all, topicMatrix("astronomy", astronomyTopicList())...)
	all = append(all, topicMatrix("medicine", medicineTopicList())...)
	all = append(all, topicMatrix("technology", techTopicList2())...)
	all = append(all, topicMatrix("economics", economicsTopicList())...)
	return all
}

// topicMatrix generates 4 entries per topic (overview, history, applications, research)
func topicMatrix(domain string, topics []string) []mbp.WriteRequest {
	var out []mbp.WriteRequest
	for _, t := range topics {
		out = append(out,
			mbp.WriteRequest{
				Concept:    fmt.Sprintf("%s overview: %s", domain, t),
				Content:    fmt.Sprintf("Overview of %s in %s: This concept covers fundamental principles, key definitions, and the theoretical framework underpinning %s. Understanding %s requires familiarity with related concepts in the field and the mathematical or empirical tools used to study it. The scope ranges from foundational theory to practical manifestations.", t, domain, t, t),
				Tags:       []string{domain, t, "overview"},
				Confidence: 0.82,
				Stability:  0.80,
			},
			mbp.WriteRequest{
				Concept:    fmt.Sprintf("%s history: %s", domain, t),
				Content:    fmt.Sprintf("Historical development of %s: The concept of %s emerged through a series of discoveries, experiments, and theoretical insights. Key figures contributed foundational work that shaped modern understanding. Debates and paradigm shifts occurred as evidence accumulated. The history of %s reflects broader trends in %s research.", t, t, t, domain),
				Tags:       []string{domain, t, "history", "historical"},
				Confidence: 0.80,
				Stability:  0.82,
			},
			mbp.WriteRequest{
				Concept:    fmt.Sprintf("%s applications: %s", domain, t),
				Content:    fmt.Sprintf("Practical applications of %s in %s: Knowledge of %s enables numerous technological, medical, engineering, or scientific applications. Industry and research institutions have developed methods to harness these principles. Economic impact and societal benefits arise from applied %s research. Current engineering challenges involve scaling and optimizing these applications.", t, domain, t, domain),
				Tags:       []string{domain, t, "applications", "engineering"},
				Confidence: 0.80,
				Stability:  0.80,
			},
			mbp.WriteRequest{
				Concept:    fmt.Sprintf("%s research: %s", domain, t),
				Content:    fmt.Sprintf("Current research frontiers in %s: Open questions in %s drive active investigation. Recent experimental findings have challenged or confirmed theoretical predictions. New techniques and instruments have enabled observations not previously possible. Interdisciplinary connections to other %s subfields and adjacent disciplines create new research opportunities.", t, t, domain),
				Tags:       []string{domain, t, "research", "frontier"},
				Confidence: 0.78,
				Stability:  0.78,
			},
		)
	}
	return out
}

func physicsTopicList() []string {
	return []string{
		"quantum mechanics wave function", "special relativity Lorentz transformation",
		"general relativity curved spacetime", "electromagnetism Maxwell equations",
		"thermodynamics entropy heat", "statistical mechanics partition function",
		"quantum field theory Feynman diagram", "standard model particle physics",
		"nuclear physics fission fusion", "condensed matter superconductor",
		"optics diffraction interference", "acoustics sound wave",
		"fluid dynamics turbulence", "plasma physics tokamak",
		"chaos theory Lorenz attractor", "string theory M-theory",
		"dark matter detection", "dark energy Lambda CDM",
		"gravitational waves LIGO", "black hole singularity",
		"quantum computing qubit", "quantum cryptography BB84",
		"laser coherent light photon", "semiconductor transistor band gap",
		"magnetic resonance NMR spin", "X-ray crystallography diffraction",
		"radioactivity decay half-life", "cosmic ray muon detector",
		"Higgs boson electroweak symmetry", "neutrino oscillation mass",
		"Bose-Einstein condensate superfluidity", "spin glass frustrated magnet",
		"topological insulator quantum Hall", "metamaterial negative refractive",
		"photovoltaic solar cell efficiency", "thermoelectric Seebeck coefficient",
		"piezoelectric crystal pressure voltage", "nonlinear optics harmonic generation",
		"particle accelerator synchrotron cyclotron", "muon g-2 anomalous magnetic moment",
		"CP violation matter antimatter asymmetry", "Casimir effect vacuum fluctuation",
		"Josephson junction flux quantum", "Hall effect van der Pauw",
		"Mössbauer effect nuclear gamma", "ultrafast laser attosecond pulse",
		"quantum dot semiconductor nanocrystal", "spintronics giant magnetoresistance",
		"photonic crystal optical band gap", "graphene Dirac fermion",
		"quantum entanglement Bell inequality", "decoherence quantum-to-classical",
	}
}

func biologyTopicList() []string {
	return []string{
		"DNA replication polymerase fidelity", "RNA transcription promoter",
		"protein synthesis ribosome translation", "cell cycle checkpoint regulation",
		"apoptosis caspase programmed death", "stem cell differentiation iPSC",
		"epigenetics methylation histone modification", "gene regulation transcription factor",
		"immune system T cell B cell antibody", "vaccine adjuvant immunity",
		"neural circuit learning memory engram", "neurotransmitter receptor synapse",
		"heart muscle contraction cardiac", "kidney nephron filtration",
		"liver metabolism detoxification", "lung alveoli gas exchange",
		"hormone endocrine feedback cortisol", "reproductive biology gamete fertilization",
		"development embryo gastrulation", "aging senescence telomere",
		"cancer oncogene tumor suppressor", "metastasis invasion angiogenesis",
		"photosynthesis chloroplast light", "cellular respiration mitochondria ATP",
		"enzyme kinetics Michaelis-Menten", "membrane transport pump channel",
		"cell signaling receptor kinase", "CRISPR gene editing Cas9",
		"sequencing genome assembly alignment", "metagenomics microbiome diversity",
		"evolution natural selection fitness", "population genetics Hardy-Weinberg",
		"ecology food web trophic level", "symbiosis mutualism parasitism",
		"antibiotic resistance mechanism efflux", "viral replication infection",
		"prion misfolding aggregation", "autophagy lysosome degradation",
		"CRISPR agriculture crop improvement", "synthetic biology BioBrick pathway",
		"optogenetics channelrhodopsin circuit", "single cell sequencing atlas",
		"organoid 3D culture model", "CRISPR screen fitness essential gene",
		"DNA repair mismatch nucleotide excision", "chromosome segregation kinetochore",
		"protein folding chaperone aggregation", "proteomics mass spectrometry identification",
	}
}

func csTopicList() []string {
	return []string{
		"algorithm complexity Big O notation", "data structure tree heap graph",
		"sorting quicksort mergesort heapsort", "dynamic programming memoization",
		"graph algorithm Dijkstra BFS DFS", "string algorithm suffix tree",
		"machine learning supervised gradient", "neural network backpropagation",
		"transformer attention BERT GPT NLP", "convolutional network image classification",
		"reinforcement learning policy reward", "generative model GAN diffusion",
		"operating system process scheduler", "memory management virtual paging",
		"file system journaling block device", "database SQL ACID transaction",
		"distributed system consensus Raft", "CAP theorem consistency availability",
		"networking TCP congestion control", "cryptography public key RSA",
		"compiler parser AST code generation", "programming language type system",
		"functional programming monad category", "concurrent programming lock-free",
		"cloud computing serverless container", "Kubernetes orchestration service mesh",
		"microservices API gateway circuit breaker", "event-driven CQRS event sourcing",
		"security vulnerability SQL injection XSS", "authentication OAuth JWT session",
		"testing TDD property-based mutation", "version control git branching merge",
		"CI/CD deployment pipeline rollout", "observability metrics tracing logging",
		"data engineering ETL pipeline warehouse", "streaming Kafka Flink real-time",
		"recommendation collaborative filtering", "search ranking BM25 vector",
		"natural language processing tokenizer", "computer vision object detection",
		"robotics SLAM path planning kinematics", "embedded system RTOS interrupt",
		"formal verification model checking proof", "program synthesis automated",
		"quantum algorithm Shor Grover simulation", "federated learning privacy",
		"interpretability explainable AI SHAP LIME", "MLOps model deployment monitoring",
	}
}

func mathTopicList() []string {
	return []string{
		"prime number distribution Riemann hypothesis", "algebraic topology homology",
		"differential geometry Riemannian curvature", "complex analysis Cauchy Riemann",
		"abstract algebra group ring field", "linear algebra matrix eigenvalue",
		"probability measure theory sigma algebra", "statistics Bayesian frequentist",
		"number theory Diophantine equation", "combinatorics generating function",
		"graph theory chromatic number planar", "topology manifold homeomorphism",
		"set theory ZFC axiom choice", "mathematical logic Gödel completeness",
		"functional analysis Hilbert Banach space", "partial differential equations",
		"ordinary differential equation stability", "numerical analysis finite element",
		"optimization convex duality KKT", "game theory Nash equilibrium",
		"information theory Shannon entropy", "coding theory error correction",
		"category theory functor adjunction", "algebraic geometry variety scheme",
		"representation theory character table", "harmonic analysis Fourier transform",
		"stochastic processes Brownian motion", "Markov chain Monte Carlo",
		"knot theory polynomial invariant", "Ramsey theory monochromatic",
		"analytic number theory L-function", "arithmetic geometry Langlands",
		"tropical geometry Newton polytope", "Floer homology symplectic",
		"modular form elliptic curve", "p-adic number ultrametric",
		"geometric group theory word metric", "model theory saturation ultraproduct",
		"proof theory ordinal analysis", "reverse mathematics subsystem",
		"descriptive set theory Borel hierarchy", "ergodic theory measure-preserving",
		"dynamical system bifurcation chaos", "random matrix theory eigenvalue",
		"percolation theory phase transition", "optimal transport Wasserstein",
		"spectral theory operator self-adjoint", "noncommutative geometry Connes",
	}
}

func historyTopicList() []string {
	return []string{
		"Mesopotamia Sumer cuneiform city-state", "ancient Egypt pharaoh pyramid Nile",
		"ancient Greece democracy philosophy", "Roman Republic Senate consul",
		"Roman Empire Augustus expansion", "Byzantine Empire Constantinople orthodox",
		"Islamic caliphate expansion Arabia", "Mongol Empire Genghis Khan steppe",
		"Medieval Europe feudalism crusade", "Black Death plague 14th century",
		"Renaissance humanism Florence art", "Reformation Luther Protestant",
		"Age of Exploration Columbus Portugal", "Scientific Revolution Copernicus Newton",
		"Enlightenment Voltaire reason rights", "French Revolution Bastille republic",
		"American Revolution independence", "Napoleonic Wars Europe congress",
		"Industrial Revolution factory steam", "British Empire colony trade",
		"American Civil War slavery reconstruction", "Meiji Restoration Japan modernization",
		"World War I trench gas alliance", "Russian Revolution Bolshevik Lenin",
		"Great Depression stock market 1929", "World War II Holocaust Normandy",
		"Cold War nuclear containment Berlin", "decolonization Africa independence",
		"Chinese Revolution Mao Communist", "Korean War armistice DMZ",
		"Vietnam War guerrilla protest", "Cuban Missile Crisis Kennedy",
		"civil rights movement Rosa Parks", "space race Sputnik Apollo moon",
		"fall of Berlin Wall USSR collapse", "apartheid Mandela South Africa",
		"Yugoslav Wars Balkans ethnic", "Gulf War Iraq Kuwait 1991",
		"September 11 terrorism Afghanistan", "Arab Spring revolution Egypt Tunisia",
		"ancient China dynasty Confucius", "Han Dynasty Silk Road trade",
		"Tang Dynasty golden age poetry", "Qing Dynasty Manchu decline",
		"Mughal India Akbar Taj Mahal", "Ottoman Empire Suleiman decline",
		"Aztec Empire Tenochtitlan Cortés", "Inca Empire Andean Pizarro",
		"Viking Age Scandinavia raid", "medieval Japan samurai shogunate",
	}
}

func philosophyTopicList() []string {
	return []string{
		"epistemology justified true belief", "metaphysics substance causation",
		"philosophy of mind consciousness qualia", "free will determinism compatibilism",
		"ethics consequentialism deontology", "virtue ethics Aristotle eudaimonia",
		"political philosophy justice Rawls", "philosophy of language meaning",
		"logic validity soundness inference", "philosophy of science falsification",
		"phenomenology Husserl intentionality", "existentialism Sartre authenticity",
		"pragmatism Dewey James Peirce", "analytic philosophy Russell logic",
		"continental philosophy Heidegger", "postmodernism Derrida deconstruction",
		"philosophy of religion ontological argument", "Buddhist philosophy impermanence",
		"Confucian ethics ren li filial", "Stoic philosophy virtue indifferent",
		"philosophy of mathematics Platonism", "aesthetics beauty sublime Kant",
		"social epistemology testimony evidence", "feminist philosophy care ethics",
		"environmental ethics intrinsic value", "bioethics autonomy informed consent",
		"philosophy of time A-theory B-theory", "personal identity Parfit",
		"philosophy of action intention reason", "meta-ethics realism expressivism",
	}
}

func neuroscienceTopicList() []string {
	return []string{
		"action potential neuron sodium potassium", "synaptic transmission vesicle release",
		"long-term potentiation NMDA calcium", "hippocampus memory consolidation",
		"prefrontal cortex executive function", "amygdala fear conditioning emotion",
		"basal ganglia reward dopamine striatum", "cerebellum coordination motor learning",
		"visual cortex V1 V2 receptive field", "auditory cortex tonotopy",
		"neurogenesis adult brain SVZ SGZ", "glial cell astrocyte myelination",
		"blood brain barrier tight junction", "neuroinflammation microglia cytokine",
		"Alzheimer's disease amyloid tau plaque", "Parkinson's dopamine substantia nigra",
		"epilepsy seizure EEG anticonvulsant", "schizophrenia dopamine hallucination",
		"depression serotonin neuroplasticity SSRI", "addiction dopamine reward circuit",
		"sleep REM slow wave consolidation", "circadian rhythm melatonin SCN",
		"pain nociception substance P gate control", "proprioception muscle spindle",
		"olfaction olfactory bulb receptor", "taste gustatory receptor",
		"cortical oscillation gamma beta theta", "default mode network resting state",
		"brain connectivity fMRI diffusion tensor", "optogenetics channelrhodopsin circuit",
		"CRISPR neural gene therapy delivery", "deep brain stimulation Parkinson",
		"brain computer interface electrode decode", "consciousness global workspace theory",
	}
}

func chemTopicList() []string {
	return []string{
		"periodic table electronegativity trend", "chemical bonding covalent ionic",
		"acid base buffer Henderson pH", "oxidation reduction electrochemistry",
		"organic reaction mechanism SN1 SN2", "aromatic benzene resonance",
		"polymer addition condensation chain", "thermochemistry enthalpy Hess",
		"kinetics rate law Arrhenius activation", "equilibrium Le Chatelier Keq",
		"spectroscopy NMR IR mass structure", "chromatography HPLC separation",
		"pharmaceutical drug design synthesis", "catalysis heterogeneous zeolite",
		"green chemistry solvent sustainability", "materials metal alloy crystal",
		"electrochemistry battery fuel cell", "surface chemistry adsorption",
		"atmospheric chemistry ozone photolysis", "analytical chemistry titration",
		"biochemistry enzyme coenzyme pathway", "lipid membrane phospholipid",
		"carbohydrate glucose glycolysis", "amino acid peptide protein bond",
		"nucleotide nucleic acid RNA DNA", "photochemistry excited state reaction",
		"coordination complex ligand field", "organometallic catalysis palladium",
		"supramolecular host guest assembly", "computational DFT molecular orbital",
	}
}

func astronomyTopicList() []string {
	return []string{
		"stellar formation molecular cloud collapse", "main sequence HR diagram",
		"red giant helium fusion shell", "white dwarf degenerate cooling",
		"neutron star pulsar millisecond", "black hole accretion disk jet",
		"supernova core collapse Type Ia", "planetary nebula asymptotic giant",
		"binary star mass transfer evolution", "variable star Cepheid RR Lyrae",
		"exoplanet transit radial velocity", "habitable zone liquid water",
		"brown dwarf T dwarf substellar", "protoplanetary disk planet formation",
		"solar system Jupiter Trojan asteroid", "Kuiper belt Oort cloud comet",
		"Galaxy Milky Way spiral arm halo", "galactic center Sgr A* orbit",
		"galaxy merger elliptical formation", "starburst galaxy intense star formation",
		"active galactic nucleus quasar jet", "gravitational lensing arc Einstein",
		"cosmic microwave background anisotropy", "Big Bang nucleosynthesis helium",
		"dark matter rotation curve halo", "dark energy accelerating expansion",
		"cosmic inflation slow roll spectrum", "large scale structure filament void",
		"gravitational wave merger inspiral", "fast radio burst localization magnetar",
		"space telescope JWST infrared imaging", "radio telescope VLBI resolution",
		"multi-messenger astronomy EM neutron", "Hubble constant tension distance",
	}
}

func medicineTopicList() []string {
	return []string{
		"cardiovascular atherosclerosis plaque", "hypertension blood pressure ACE",
		"diabetes insulin resistance HbA1c", "obesity BMI metabolic syndrome",
		"cancer immunotherapy checkpoint PD-1", "chemotherapy cytotoxic resistance",
		"radiation therapy fractionation damage", "surgery laparoscopic robotic",
		"anesthesia regional spinal epidural", "pain management opioid NSAIDs",
		"infectious disease antibiotic stewardship", "vaccine mRNA spike immune",
		"HIV antiretroviral HAART CD4", "tuberculosis multidrug resistance",
		"malaria artemisinin Plasmodium", "hepatitis liver cirrhosis antiviral",
		"respiratory COPD asthma inhaler bronchodilator", "pneumonia antibiotic oxygen",
		"neurology stroke ischemia thrombolysis", "epilepsy seizure EEG anticonvulsant",
		"psychiatry depression anxiety SSRI CBT", "schizophrenia antipsychotic",
		"Alzheimer's amyloid tau dementia", "Parkinson's levodopa dopamine",
		"kidney chronic renal failure dialysis", "liver transplant rejection",
		"autoimmune rheumatoid lupus biologic", "inflammatory bowel Crohn's colitis",
		"dermatology psoriasis eczema topical", "ophthalmology retina glaucoma laser",
		"orthopedics fracture fixation joint", "pediatrics growth vaccination milestone",
		"geriatrics frailty polypharmacy fall", "palliative care hospice symptom",
		"radiology CT MRI PET scan contrast", "pathology biopsy histology diagnosis",
		"pharmacology receptor agonist antagonist", "clinical trial randomized blinding",
	}
}

func techTopicList2() []string {
	return []string{
		"internet protocol TCP IP routing BGP", "HTTP HTTPS TLS certificate authority",
		"DNS resolver authoritative zone", "wireless WiFi 5G spectrum OFDM",
		"processor CPU pipeline cache RISC", "GPU compute shader CUDA parallelism",
		"memory DRAM HBM bandwidth latency", "storage SSD NVMe flash wear leveling",
		"cloud AWS Azure GCP serverless", "container Docker image layer registry",
		"operating system Linux kernel syscall", "file system ext4 btrfs ZFS",
		"database PostgreSQL index ACID", "NoSQL MongoDB Cassandra document",
		"message queue Kafka topic partition", "search Elasticsearch inverted index",
		"machine learning TensorFlow PyTorch", "inference serving ONNX quantization",
		"cryptography AES RSA elliptic curve", "blockchain Bitcoin mining hash",
		"smart contract Ethereum Solidity EVM", "IPFS distributed content address",
		"IoT sensor MQTT edge gateway", "autonomous vehicle LIDAR fusion",
		"natural language speech recognition", "computer vision segmentation detection",
		"augmented reality spatial computing", "virtual reality haptic rendering",
		"renewable solar wind grid storage", "electric vehicle battery BMS",
		"quantum computing superconducting ion", "neuromorphic chip spiking",
		"3D printing additive FDM SLA sintering", "semiconductor fab photolithography EUV",
	}
}

func economicsTopicList() []string {
	return []string{
		"supply demand equilibrium price elasticity", "GDP growth inflation unemployment",
		"monetary policy interest rate central bank", "fiscal policy government spending tax",
		"comparative advantage trade Ricardo", "market failure externality public good",
		"game theory Nash prisoner dilemma", "behavioral economics prospect theory",
		"microeconomics consumer surplus utility", "macroeconomics Keynesian multiplier",
		"international trade WTO globalization", "currency exchange rate purchasing power",
		"financial markets equity bond derivative", "option pricing Black-Scholes",
		"portfolio theory Markowitz diversification", "efficient market hypothesis",
		"labor economics wage inequality union", "income inequality Gini coefficient",
		"development economics poverty trap", "health economics cost-effectiveness",
		"environmental economics carbon tax Pigou", "auction theory mechanism design",
		"information economics adverse selection", "principal agent moral hazard",
		"industrial organization monopoly Cournot", "network effects platform economy",
		"innovation economics patents R&D growth", "economic history industrial revolution",
		"financial crisis contagion systemic risk", "cryptocurrency Bitcoin economics",
	}
}
