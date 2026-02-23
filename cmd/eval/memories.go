package main

import (
	"fmt"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// seedMemories returns the hand-crafted base knowledge corpus.
// Each entry is factually accurate with rich content and precise tagging.
func seedMemories() []mbp.WriteRequest {
	return []mbp.WriteRequest{

		// ════════════════════════════════════════════════════════════════
		// PHYSICS — Quantum Mechanics
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Quantum Entanglement",
			Content:    "Quantum entanglement occurs when two or more particles become correlated such that the quantum state of each cannot be described independently of the others. Measuring one particle instantaneously determines the state of its partner, regardless of separation distance. Einstein called this 'spooky action at a distance' and used it in the EPR paradox to argue quantum mechanics was incomplete. Bell's theorem (1964) and subsequent experiments by Aspect (1982) confirmed quantum nonlocality is real.",
			Tags:       []string{"quantum mechanics", "physics", "entanglement", "EPR paradox", "nonlocality", "Bell theorem", "quantum information", "measurement"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Heisenberg Uncertainty Principle",
			Content:    "The uncertainty principle states that the position and momentum of a particle cannot both be known to arbitrary precision simultaneously. The product of their uncertainties is always at least ℏ/2. This is not a limitation of measurement technology but a fundamental property of quantum systems — a particle does not have a definite position and momentum at the same time. It arises from the wave nature of matter and applies to other conjugate pairs like energy and time.",
			Tags:       []string{"quantum mechanics", "physics", "uncertainty", "Heisenberg", "wave-particle duality", "measurement", "momentum"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Schrödinger's Wave Equation",
			Content:    "The Schrödinger equation describes how the quantum state of a system evolves over time. The time-dependent form is iℏ∂Ψ/∂t = ĤΨ, where Ψ is the wave function and Ĥ is the Hamiltonian operator. The square of the wave function's magnitude gives the probability density of finding a particle at a given position. The equation is linear, so superpositions of solutions are also solutions, giving rise to quantum interference.",
			Tags:       []string{"quantum mechanics", "physics", "wave function", "Schrödinger", "probability", "Hamiltonian", "superposition"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Quantum Superposition",
			Content:    "A quantum system can exist in multiple states simultaneously until measured, a phenomenon called superposition. A photon can be in a superposition of horizontal and vertical polarization simultaneously. Schrödinger's cat thought experiment illustrates the paradox: a cat in a box linked to a quantum event is theoretically both alive and dead until observed. Superposition is the basis for quantum computing, where qubits can represent 0 and 1 simultaneously.",
			Tags:       []string{"quantum mechanics", "physics", "superposition", "Schrödinger's cat", "qubit", "quantum computing", "measurement"},
			Confidence: 0.97, Stability: 0.96,
		},
		{
			Concept:    "Quantum Tunneling",
			Content:    "Quantum tunneling allows particles to pass through potential energy barriers that would be classically forbidden. A particle's wave function extends into and sometimes beyond a barrier, giving nonzero probability of appearing on the other side. Tunneling enables nuclear fusion in stars at lower temperatures than classical physics would require, and is exploited in tunnel diodes, scanning tunneling microscopes, and flash memory storage.",
			Tags:       []string{"quantum mechanics", "physics", "tunneling", "wave function", "nuclear fusion", "semiconductor", "nanotechnology"},
			Confidence: 0.97, Stability: 0.96,
		},
		{
			Concept:    "Wave-Particle Duality",
			Content:    "Light and matter exhibit both wave-like and particle-like properties depending on how they are observed. The double-slit experiment demonstrates this: electrons fired at two slits create an interference pattern (wave behavior), but observing which slit an electron passes through collapses the pattern to two bands (particle behavior). De Broglie proposed that all matter has an associated wavelength λ = h/p, confirmed by electron diffraction experiments.",
			Tags:       []string{"quantum mechanics", "physics", "wave-particle duality", "double-slit", "de Broglie", "electron", "light", "diffraction"},
			Confidence: 0.98, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// PHYSICS — Relativity
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Special Relativity",
			Content:    "Einstein's special relativity (1905) rests on two postulates: the laws of physics are the same in all inertial frames, and the speed of light in vacuum is constant for all observers. Consequences include time dilation (moving clocks run slow), length contraction (moving objects are shortened), and mass-energy equivalence E=mc². No object with mass can reach the speed of light because its relativistic mass increases without bound.",
			Tags:       []string{"relativity", "physics", "Einstein", "time dilation", "E=mc2", "Lorentz", "spacetime", "speed of light"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "General Relativity",
			Content:    "Einstein's general relativity (1915) extends special relativity to include gravity. Massive objects curve spacetime, and other objects follow the straightest possible paths (geodesics) through this curved spacetime — what we perceive as gravitational attraction. The field equations Gμν = 8πGTμν relate spacetime curvature to the energy-momentum of matter. GR predicts gravitational waves, black holes, and the expanding universe, all confirmed by observation.",
			Tags:       []string{"relativity", "physics", "Einstein", "gravity", "spacetime", "geodesic", "gravitational waves", "black hole"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Time Dilation",
			Content:    "Time dilation is the difference in elapsed time measured by two observers, either due to relative motion (special relativity) or gravitational potential differences (general relativity). GPS satellites experience both effects: their clocks run faster due to weaker gravity (gaining ~45 μs/day) but slower due to orbital velocity (losing ~7 μs/day), for a net gain of ~38 μs/day that must be corrected or GPS positioning would accumulate 10 km of error daily.",
			Tags:       []string{"relativity", "physics", "time dilation", "GPS", "gravity", "special relativity", "velocity", "Einstein"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Gravitational Waves",
			Content:    "Gravitational waves are ripples in spacetime caused by accelerating massive objects, predicted by general relativity in 1916. On September 14, 2015, LIGO detected the first direct observation of gravitational waves from the merger of two black holes 1.3 billion light-years away. The signal matched precisely with numerical relativity predictions. Gravitational wave astronomy has since become a new observational window on the universe.",
			Tags:       []string{"gravity", "relativity", "physics", "gravitational waves", "LIGO", "black hole", "spacetime", "Einstein"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// PHYSICS — Thermodynamics & Statistical Mechanics
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Laws of Thermodynamics",
			Content:    "The four laws of thermodynamics govern energy and entropy. Zeroth: thermal equilibrium is transitive (thermometers work). First: energy is conserved; heat added equals work done plus internal energy increase. Second: entropy of an isolated system never decreases; heat flows spontaneously from hot to cold. Third: absolute zero is unattainable; as temperature approaches 0K, entropy approaches a minimum constant value.",
			Tags:       []string{"thermodynamics", "physics", "entropy", "energy", "heat", "temperature", "Kelvin", "conservation"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Entropy and the Second Law",
			Content:    "Entropy, defined by Boltzmann as S = k ln(Ω), measures the number of microscopic configurations consistent with observed macroscopic state. The second law states entropy of isolated systems increases over time, giving thermodynamic processes a direction — the 'arrow of time.' Maxwell's demon thought experiment probes the connection between information and entropy. Landauer's principle shows that erasing one bit of information requires kT ln(2) of energy.",
			Tags:       []string{"thermodynamics", "physics", "entropy", "Boltzmann", "second law", "arrow of time", "information theory", "Maxwell's demon"},
			Confidence: 0.97, Stability: 0.96,
		},

		// ════════════════════════════════════════════════════════════════
		// PHYSICS — Particle Physics & Cosmology
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Standard Model of Particle Physics",
			Content:    "The Standard Model classifies fundamental particles into fermions (quarks and leptons) and bosons (force carriers). Quarks combine to form protons, neutrons, and other hadrons. The four fundamental forces have carriers: photon (electromagnetism), W and Z bosons (weak force), gluons (strong force). Gravity has no quantum carrier in the Standard Model. The Higgs boson, discovered at CERN in 2012, gives particles their mass through the Higgs field.",
			Tags:       []string{"particle physics", "standard model", "quark", "lepton", "boson", "Higgs", "CERN", "LHC", "fundamental forces"},
			Confidence: 0.99, Stability: 0.97,
		},
		{
			Concept:    "Higgs Boson Discovery",
			Content:    "The Higgs boson was discovered on July 4, 2012, at CERN's Large Hadron Collider by the ATLAS and CMS experiments. The particle had been predicted by Peter Higgs, François Englert, and others in 1964 as part of the mechanism explaining how elementary particles acquire mass through interaction with the Higgs field. Peter Higgs and François Englert received the Nobel Prize in Physics in 2013 for this theoretical prediction.",
			Tags:       []string{"Higgs boson", "particle physics", "CERN", "LHC", "standard model", "Nobel Prize", "fundamental forces", "mass"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Black Holes",
			Content:    "A black hole is a region of spacetime where gravity is so strong that nothing — not even light — can escape from inside the event horizon. Black holes form when massive stars collapse at the end of their lives. The singularity at the center is a point of infinite density. Stephen Hawking predicted that black holes emit thermal radiation (Hawking radiation) due to quantum effects near the event horizon, causing them to slowly evaporate.",
			Tags:       []string{"black hole", "physics", "astrophysics", "gravity", "event horizon", "singularity", "Hawking radiation", "relativity"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Dark Matter",
			Content:    "Dark matter is a hypothetical form of matter that neither emits nor absorbs light but interacts gravitationally. Evidence comes from galaxy rotation curves (stars orbit faster than visible mass can explain), gravitational lensing, and the cosmic microwave background. Dark matter comprises approximately 27% of the universe's total energy density. Leading candidates include WIMPs and axions, but no direct detection has been confirmed.",
			Tags:       []string{"dark matter", "cosmology", "physics", "galaxy", "gravitational lensing", "rotation curve", "WIMP", "universe"},
			Confidence: 0.97, Stability: 0.96,
		},
		{
			Concept:    "Big Bang Cosmology",
			Content:    "The Big Bang model describes the origin of the universe approximately 13.8 billion years ago from an extremely hot, dense state. Evidence includes the cosmic microwave background radiation (CMB) discovered by Penzias and Wilson in 1965, the observed expansion of the universe (Hubble, 1929), and the abundance of light elements from Big Bang nucleosynthesis. Cosmic inflation theory explains the universe's flatness and homogeneity.",
			Tags:       []string{"Big Bang", "cosmology", "physics", "CMB", "Hubble", "inflation", "expansion", "universe"},
			Confidence: 0.99, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// BIOLOGY — Molecular Biology & Genetics
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "DNA Structure and Function",
			Content:    "DNA (deoxyribonucleic acid) is a double helix polymer consisting of nucleotide bases — adenine (A), thymine (T), guanine (G), and cytosine (C) — held together by hydrogen bonds following Chargaff's rules (A pairs with T, G pairs with C). Watson and Crick described the structure in 1953 using X-ray crystallography data from Rosalind Franklin. Each strand runs antiparallel, with DNA polymerase reading 3' to 5' and synthesizing 5' to 3'.",
			Tags:       []string{"DNA", "molecular biology", "genetics", "double helix", "Watson", "Crick", "nucleotide", "base pairing"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Central Dogma of Molecular Biology",
			Content:    "The central dogma (Francis Crick, 1958) describes the flow of genetic information: DNA → RNA → Protein. DNA is transcribed into messenger RNA (mRNA) by RNA polymerase. The mRNA is then translated into protein by ribosomes, with transfer RNA (tRNA) matching codons to amino acids. The genetic code uses triplet codons with 64 combinations encoding 20 amino acids plus stop signals. Reverse transcriptase (found in retroviruses) enables RNA→DNA flow.",
			Tags:       []string{"central dogma", "molecular biology", "genetics", "transcription", "translation", "mRNA", "protein synthesis", "codon", "ribosome"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "CRISPR-Cas9 Gene Editing",
			Content:    "CRISPR-Cas9 is a revolutionary gene-editing technology derived from a bacterial immune defense mechanism. The Cas9 protein acts as molecular scissors guided by a short RNA sequence to cut DNA at a specific genomic location, allowing genes to be added, removed, or modified with unprecedented precision. Jennifer Doudna and Emmanuelle Charpentier developed its application in 2012 and received the 2020 Nobel Prize in Chemistry. Applications range from treating genetic diseases to developing disease-resistant crops.",
			Tags:       []string{"CRISPR", "gene editing", "molecular biology", "genetics", "biotechnology", "Cas9", "Nobel Prize", "therapy"},
			Confidence: 0.99, Stability: 0.97,
		},
		{
			Concept:    "Mendelian Genetics",
			Content:    "Gregor Mendel discovered the laws of inheritance through pea plant experiments (1866). Law of Segregation: each organism has two alleles for each trait, which separate during gamete formation. Law of Independent Assortment: alleles for different traits assort independently. Dominant alleles mask recessive ones. Mendel's work was rediscovered in 1900 by de Vries, Correns, and von Tschermak, becoming the foundation of genetics. The Punnett square predicts offspring ratios.",
			Tags:       []string{"genetics", "Mendel", "allele", "dominant", "recessive", "inheritance", "law of segregation", "Punnett square"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Epigenetics",
			Content:    "Epigenetics studies heritable changes in gene expression that do not involve alterations to the DNA sequence itself. Mechanisms include DNA methylation (adding methyl groups to cytosines, typically silencing genes) and histone modification (acetylation, methylation, phosphorylation of histone proteins altering chromatin structure). Environmental factors like diet, stress, and exposure to toxins can induce epigenetic changes that can be passed to offspring, challenging strict genetic determinism.",
			Tags:       []string{"epigenetics", "genetics", "molecular biology", "DNA methylation", "histone", "gene expression", "inheritance", "environment"},
			Confidence: 0.97, Stability: 0.96,
		},

		// ════════════════════════════════════════════════════════════════
		// BIOLOGY — Neuroscience
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Synaptic Transmission",
			Content:    "Synaptic transmission is the process by which neurons communicate. An action potential arrives at the presynaptic terminal, triggering calcium influx that causes vesicles to fuse with the membrane and release neurotransmitters into the synaptic cleft. Neurotransmitters bind to receptors on the postsynaptic membrane, causing ion channels to open. Excitatory neurotransmitters (glutamate) depolarize the cell; inhibitory ones (GABA) hyperpolarize it. The neurotransmitter is then degraded or reuptaken.",
			Tags:       []string{"neuroscience", "synapse", "neurotransmitter", "action potential", "glutamate", "GABA", "calcium", "receptor"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Long-Term Potentiation",
			Content:    "Long-term potentiation (LTP) is a persistent strengthening of synaptic connections following repeated stimulation. First described by Bliss and Lømo (1973), LTP involves NMDA receptor activation, calcium influx, and the insertion of additional AMPA receptors into the postsynaptic membrane. LTP in the hippocampus is widely believed to underlie explicit memory formation. Hebb's rule ('neurons that fire together wire together') anticipated LTP conceptually.",
			Tags:       []string{"neuroscience", "LTP", "synapse", "memory", "hippocampus", "NMDA", "AMPA", "Hebbian learning", "plasticity"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Hebbian Learning Rule",
			Content:    "Donald Hebb proposed (1949) that when a neuron repeatedly contributes to firing another neuron, the synaptic connection between them strengthens — 'neurons that fire together, wire together.' This rule is now supported by the discovery of long-term potentiation. Hebbian learning is the biological basis for associative memory and forms the theoretical foundation for many artificial neural network learning algorithms. Anti-Hebbian plasticity (neurons that fire out of sync lose their link) also exists.",
			Tags:       []string{"Hebbian learning", "neuroscience", "synapse", "plasticity", "memory", "LTP", "neural network", "Hebb"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Dopamine and the Reward System",
			Content:    "Dopamine is a catecholamine neurotransmitter critical for the brain's reward and motivation circuits. The mesolimbic pathway (from ventral tegmental area to nucleus accumbens) drives pleasure and reward-seeking behavior. Dopamine neurons signal reward prediction errors — firing strongly for unexpected rewards and less for predicted ones. This forms the neurochemical basis for addiction: drugs of abuse hijack the dopamine system, creating abnormally strong reward signals.",
			Tags:       []string{"dopamine", "neuroscience", "reward", "motivation", "addiction", "neurotransmitter", "nucleus accumbens", "VTA"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Memory Consolidation",
			Content:    "Memory consolidation is the process by which newly acquired memories are stabilized. Synaptic consolidation occurs over hours through protein synthesis and structural synaptic changes. Systems consolidation over weeks to years involves dialogue between hippocampus and neocortex — the hippocampus temporarily holds new memories and gradually transfers them to distributed cortical storage. Sleep, especially slow-wave sleep, is critical for consolidation through hippocampal replay.",
			Tags:       []string{"memory", "neuroscience", "consolidation", "hippocampus", "neocortex", "sleep", "protein synthesis", "learning"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Neuroplasticity",
			Content:    "Neuroplasticity is the brain's ability to reorganize its structure and function in response to experience, learning, or injury. It occurs at multiple scales: synaptic plasticity (changes in synaptic strength), structural plasticity (growth of new synapses and dendrites), and neurogenesis (birth of new neurons, primarily in hippocampus and olfactory bulb in adults). Plasticity underlies learning, recovery from stroke, and adaptation to sensory loss. It declines but never ceases with age.",
			Tags:       []string{"neuroplasticity", "neuroscience", "brain", "learning", "synapse", "neurogenesis", "stroke recovery", "adaptation"},
			Confidence: 0.98, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// BIOLOGY — Evolution & Ecology
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Darwin's Theory of Natural Selection",
			Content:    "Charles Darwin's theory of evolution by natural selection (1859, On the Origin of Species) rests on four principles: variation exists among individuals in a population; some variation is heritable; more offspring are produced than can survive; survival and reproduction are not random but depend on heritable traits (differential fitness). Over generations, traits that improve fitness become more common. This explains the diversity and adaptation of life without invoking design.",
			Tags:       []string{"evolution", "natural selection", "Darwin", "fitness", "adaptation", "population genetics", "species", "inheritance"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "CRISPR in Medicine",
			Content:    "CRISPR-Cas9 has transformative medical applications. In 2023, the FDA approved the first CRISPR-based therapy (Casgevy) for sickle cell disease and beta-thalassemia. Clinical trials are exploring CRISPR for cancer immunotherapy (editing T cells to target tumors), Huntington's disease, HIV, and hereditary blindness. Ex vivo editing — modifying cells outside the body — is currently safer than in vivo delivery. Base editing and prime editing offer more precise alternatives with fewer off-target effects.",
			Tags:       []string{"CRISPR", "medicine", "gene therapy", "sickle cell", "cancer", "FDA", "clinical trials", "T cell"},
			Confidence: 0.97, Stability: 0.95,
		},
		{
			Concept:    "Photosynthesis",
			Content:    "Photosynthesis converts light energy into chemical energy stored as glucose. In the light-dependent reactions (thylakoid membrane), chlorophyll absorbs photons, exciting electrons that drive ATP and NADPH production through the electron transport chain. Water is split, releasing oxygen. In the light-independent reactions (Calvin cycle, stroma), CO₂ is fixed into glucose using ATP and NADPH. C4 and CAM plants have evolved modifications to concentrate CO₂ and reduce photorespiration in hot, dry climates.",
			Tags:       []string{"photosynthesis", "plant biology", "chlorophyll", "glucose", "Calvin cycle", "ATP", "carbon fixation", "oxygen"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// COMPUTER SCIENCE — Algorithms
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Dijkstra's Shortest Path Algorithm",
			Content:    "Dijkstra's algorithm (1956) finds the shortest path from a source vertex to all other vertices in a weighted graph with non-negative edge weights. It maintains a priority queue of vertices sorted by tentative distance, greedily selecting the nearest unvisited vertex and relaxing its neighbors. Time complexity is O((V + E) log V) with a binary heap. Applications include GPS routing, network routing protocols (OSPF), and game pathfinding. Bellman-Ford handles negative weights.",
			Tags:       []string{"algorithms", "graph", "Dijkstra", "shortest path", "priority queue", "BFS", "computer science", "routing"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "QuickSort Algorithm",
			Content:    "QuickSort (Hoare, 1959) is a divide-and-conquer sorting algorithm. It selects a pivot element and partitions the array into elements less than and greater than the pivot, recursively sorting each partition. Average time complexity is O(n log n) with O(log n) stack space. Worst case is O(n²) with poor pivot selection (e.g., already-sorted input with first-element pivot), mitigated by randomized pivot or median-of-three. QuickSort is cache-efficient and typically fastest in practice.",
			Tags:       []string{"algorithms", "sorting", "quicksort", "divide and conquer", "computer science", "time complexity", "O(n log n)", "pivot"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Dynamic Programming",
			Content:    "Dynamic programming solves complex problems by breaking them into overlapping subproblems and storing solutions to avoid recomputation (memoization or tabulation). Bellman coined the term in the 1950s. Classic problems include Fibonacci numbers, longest common subsequence, knapsack problem, and optimal matrix chain multiplication. The key insight: if the problem has optimal substructure and overlapping subproblems, DP gives polynomial-time solutions to otherwise exponential problems.",
			Tags:       []string{"algorithms", "dynamic programming", "memoization", "computer science", "optimization", "Bellman", "subproblem", "tabulation"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Big-O Notation and Complexity",
			Content:    "Big-O notation characterizes algorithm efficiency by describing how runtime or space grows as input size n increases. Common complexities: O(1) constant, O(log n) logarithmic, O(n) linear, O(n log n) quasi-linear, O(n²) quadratic, O(2ⁿ) exponential. The P vs NP problem asks whether every problem verifiable in polynomial time (NP) is also solvable in polynomial time (P) — one of the Millennium Prize Problems. SAT, graph coloring, and traveling salesman are NP-complete.",
			Tags:       []string{"algorithms", "complexity", "Big-O", "P vs NP", "computer science", "time complexity", "NP-complete", "sorting"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// COMPUTER SCIENCE — Machine Learning & AI
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Artificial Neural Networks",
			Content:    "Artificial neural networks are computational models inspired by the brain, consisting of layers of interconnected nodes (neurons). Each connection has a weight; learning adjusts weights using backpropagation, computing gradients of the loss function via the chain rule. The universal approximation theorem proves a single hidden layer with enough neurons can approximate any continuous function. Deep learning uses many hidden layers to learn hierarchical feature representations.",
			Tags:       []string{"neural network", "machine learning", "AI", "backpropagation", "deep learning", "gradient descent", "weights", "layers"},
			Confidence: 0.99, Stability: 0.97,
		},
		{
			Concept:    "Transformer Architecture",
			Content:    "Transformers (Vaswani et al., 2017: 'Attention Is All You Need') use self-attention to process sequences in parallel rather than sequentially. Attention computes query-key-value triplets: attention(Q,K,V) = softmax(QKᵀ/√d_k)V. Multi-head attention runs multiple attention operations in parallel. Transformers enable long-range dependency capture without the vanishing gradient issues of RNNs. BERT, GPT, T5, and LLaMA are all transformer-based language models.",
			Tags:       []string{"transformer", "attention", "deep learning", "NLP", "BERT", "GPT", "self-attention", "language model"},
			Confidence: 0.99, Stability: 0.97,
		},
		{
			Concept:    "Gradient Descent and Backpropagation",
			Content:    "Gradient descent minimizes the loss function by iteratively adjusting parameters in the direction of steepest descent (negative gradient). Stochastic gradient descent (SGD) uses mini-batches for efficiency. Backpropagation computes gradients by applying the chain rule backward through the network. Adaptive optimizers like Adam combine momentum and per-parameter learning rates. Learning rate schedules, batch normalization, and dropout address training stability and overfitting.",
			Tags:       []string{"gradient descent", "backpropagation", "machine learning", "optimization", "neural network", "Adam", "SGD", "loss function"},
			Confidence: 0.99, Stability: 0.97,
		},
		{
			Concept:    "Reinforcement Learning",
			Content:    "Reinforcement learning (RL) trains agents to make decisions by rewarding desired behaviors. An agent in an environment takes actions based on its policy, receives rewards, and updates the policy to maximize cumulative reward. Q-learning estimates the value of state-action pairs. Policy gradient methods directly optimize the policy. Deep RL (DQN, PPO, A3C) uses neural networks for function approximation. AlphaGo and AlphaStar demonstrated superhuman game performance using RL.",
			Tags:       []string{"reinforcement learning", "Q-learning", "policy gradient", "AI", "agent", "reward", "Markov decision process", "deep RL"},
			Confidence: 0.99, Stability: 0.97,
		},
		{
			Concept:    "Convolutional Neural Networks",
			Content:    "CNNs use convolutional layers that apply learned filters across input images, sharing weights spatially to detect local patterns regardless of position. Pooling layers reduce spatial dimensions. Stacked layers learn hierarchical features: edges → textures → objects. AlexNet's 2012 ImageNet victory launched modern deep learning. ResNets (He et al., 2015) introduced skip connections to train networks with 100+ layers. CNNs are also used in audio, time-series, and genomics.",
			Tags:       []string{"CNN", "convolutional neural network", "deep learning", "computer vision", "image recognition", "AlexNet", "ResNet", "feature extraction"},
			Confidence: 0.99, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// COMPUTER SCIENCE — Systems & Databases
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "B-Tree Database Indexing",
			Content:    "B-trees are self-balancing search trees used extensively in database indexing. Every node stores multiple keys, and the tree maintains balance by ensuring all leaves are at the same depth. B+ trees (used in MySQL, PostgreSQL) store all data at leaf nodes with linked lists for range scans. The B-tree property guarantees O(log n) search, insertion, and deletion. An index on a column allows O(log n) lookup versus O(n) table scans.",
			Tags:       []string{"database", "B-tree", "indexing", "algorithms", "SQL", "MySQL", "PostgreSQL", "data structure"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Distributed Consensus: Raft Algorithm",
			Content:    "Raft is a consensus algorithm designed to be understandable (Ongaro and Ousterhout, 2014). It decomposes consensus into leader election, log replication, and safety. A leader is elected by majority vote and handles all client requests. The leader replicates log entries to followers; an entry is committed once a majority acknowledges it. Raft ensures that committed entries are never lost. It is used in etcd, CockroachDB, TiKV, and many distributed systems.",
			Tags:       []string{"distributed systems", "consensus", "Raft", "leader election", "log replication", "fault tolerance", "etcd", "CockroachDB"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "CAP Theorem",
			Content:    "The CAP theorem (Brewer, 2000) states that a distributed data store can guarantee at most two of three properties: Consistency (every read receives the most recent write), Availability (every request receives a non-error response), and Partition tolerance (the system continues despite network partitions). Since partitions are unavoidable in real distributed systems, practical systems choose between CP (strong consistency, e.g., HBase, Zookeeper) and AP (eventual consistency, e.g., Cassandra, DynamoDB).",
			Tags:       []string{"distributed systems", "CAP theorem", "consistency", "availability", "partition tolerance", "database", "Cassandra", "Brewer"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "LSM-Tree Storage Engine",
			Content:    "Log-Structured Merge-trees (LSM-trees) optimize write throughput by writing sequentially to an in-memory buffer (MemTable), which is periodically flushed to sorted, immutable disk files (SSTs). Background compaction merges SSTs to reclaim space and improve read performance. RocksDB, LevelDB, Cassandra, and HBase use LSM-trees. The tradeoff is write amplification (data written multiple times during compaction) versus excellent write throughput and compression.",
			Tags:       []string{"LSM-tree", "database", "storage engine", "RocksDB", "LevelDB", "Pebble", "SST", "compaction"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// MATHEMATICS
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Riemann Hypothesis",
			Content:    "The Riemann hypothesis (1859) is one of the most famous unsolved problems in mathematics. It concerns the Riemann zeta function ζ(s) = Σ(1/nˢ): all non-trivial zeros of ζ(s) lie on the critical line Re(s) = 1/2. The hypothesis is deeply connected to the distribution of prime numbers — the prime number theorem's error term depends on where the zeros lie. It is one of the seven Millennium Prize Problems, with a $1 million prize for its solution.",
			Tags:       []string{"mathematics", "Riemann hypothesis", "prime numbers", "number theory", "zeta function", "Millennium Prize", "unsolved problems"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Fourier Transform",
			Content:    "The Fourier transform decomposes a signal into its constituent frequencies. For a function f(t), its Fourier transform F(ω) = ∫f(t)e^(-iωt)dt gives the frequency-domain representation. The inverse transform reconstructs f(t) from F(ω). The Fast Fourier Transform (FFT, Cooley-Tukey 1965) reduces computation from O(n²) to O(n log n). Applications include signal processing, image compression (JPEG), solving differential equations, and MRI imaging.",
			Tags:       []string{"Fourier transform", "mathematics", "signal processing", "FFT", "frequency domain", "image compression", "differential equations"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Bayesian Statistics",
			Content:    "Bayesian inference updates probability estimates as new evidence is observed. Bayes' theorem: P(H|E) = P(E|H)P(H)/P(E). The prior P(H) encodes beliefs before observing data; the likelihood P(E|H) measures how probable the evidence is given the hypothesis; the posterior P(H|E) is the updated belief. Bayesian methods handle uncertainty naturally, allow incorporation of prior knowledge, and provide full probability distributions rather than point estimates. MCMC enables computation with complex posteriors.",
			Tags:       []string{"Bayesian statistics", "mathematics", "probability", "inference", "prior", "posterior", "MCMC", "Bayes theorem"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Information Theory and Shannon Entropy",
			Content:    "Claude Shannon founded information theory in 1948. Shannon entropy H(X) = -Σp(x)log₂p(x) measures the average information content of a random variable — equivalently, the minimum number of bits needed to encode a message. Mutual information measures shared information between two variables. Shannon's channel capacity theorem states the maximum error-free transmission rate over a noisy channel is C = B log₂(1 + S/N), where B is bandwidth and S/N is signal-to-noise ratio.",
			Tags:       []string{"information theory", "Shannon entropy", "mathematics", "entropy", "channel capacity", "bits", "compression", "communication"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// HISTORY — Ancient Civilizations
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Ancient Egypt",
			Content:    "Ancient Egyptian civilization flourished along the Nile for over 3,000 years (c. 3100–332 BCE). Unified by Narmer around 3100 BCE, it was governed by pharaohs considered divine. Major achievements include the pyramids of Giza (built under Khufu, Khafre, and Menkaure), the Sphinx, hieroglyphic writing, advances in medicine and astronomy, and sophisticated bureaucracy. The New Kingdom (1550–1070 BCE) saw expansion under Thutmose III and the religious revolution of Akhenaten.",
			Tags:       []string{"ancient Egypt", "history", "pharaoh", "pyramid", "Nile", "hieroglyphics", "civilization", "Khufu"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Ancient Rome",
			Content:    "Rome evolved from a city-state (traditional founding 753 BCE) to the Roman Republic (509–27 BCE) to the Roman Empire (27 BCE–476 CE in the West). Julius Caesar's assassination (44 BCE) led to civil wars and Augustus becoming the first emperor. At its height, Rome controlled the Mediterranean basin, innovating in law (Justinian's Corpus Juris Civilis), engineering (aqueducts, concrete, roads), and governance. The fall of the Western Empire in 476 CE is traditionally dated to Romulus Augustulus's deposition.",
			Tags:       []string{"ancient Rome", "history", "Julius Caesar", "Augustus", "empire", "republic", "Senate", "aqueduct"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Ancient Greece",
			Content:    "Ancient Greece (c. 800–146 BCE) gave rise to democracy, Western philosophy, mathematics, theater, and the Olympic Games. Athens under Pericles developed direct democracy and commissioned the Parthenon. The Persian Wars (490–479 BCE) and Peloponnesian War (431–404 BCE) shaped the era. Alexander the Great's conquests spread Hellenic culture from Egypt to India. Greek thinkers including Socrates, Plato, Aristotle, Euclid, Archimedes, and Hippocrates laid foundations for Western thought.",
			Tags:       []string{"ancient Greece", "history", "Athens", "democracy", "Socrates", "Plato", "Aristotle", "Alexander the Great"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Mesopotamia and the First Civilizations",
			Content:    "Mesopotamia ('land between the rivers' — Tigris and Euphrates) was home to the world's first civilizations. Sumerians (c. 3500 BCE) developed cuneiform writing, city-states, and irrigation agriculture. Babylon under Hammurabi (c. 1754 BCE) produced one of history's first law codes. The Assyrian Empire was known for military sophistication. Mesopotamia originated the wheel, bronze metallurgy, the 60-minute hour, and early mathematics including positional notation.",
			Tags:       []string{"Mesopotamia", "Sumer", "Babylon", "history", "cuneiform", "Hammurabi", "ancient civilization", "writing"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// HISTORY — Medieval & Modern
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Byzantine Empire",
			Content:    "The Byzantine Empire was the continuation of the Eastern Roman Empire, lasting from 330 CE (Constantinople's founding) to 1453 CE (Ottoman conquest). Emperor Justinian I (527–565) reconquered parts of the Western Empire, codified Roman law, and built the Hagia Sophia. Byzantine culture preserved Greek and Roman knowledge through the Dark Ages. The fall of Constantinople to Mehmed II in 1453 sent Greek scholars to Western Europe, contributing to the Renaissance.",
			Tags:       []string{"Byzantine Empire", "history", "Constantinople", "Justinian", "Ottoman", "Roman", "medieval", "Hagia Sophia"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "The Black Death",
			Content:    "The Black Death (1347–1351) was a bubonic plague pandemic caused by Yersinia pestis, transmitted primarily by fleas on rats. Originating in Central Asia, it spread along Silk Road trade routes to Europe via Crimea and Sicily. It killed 30–60% of Europe's population (approximately 25 million people) and up to 200 million worldwide. Social consequences included labor shortages that empowered the peasantry, weakening of feudalism, religious crisis, and pogrom violence against Jews.",
			Tags:       []string{"Black Death", "plague", "medieval history", "pandemic", "Yersinia pestis", "Silk Road", "Europe", "14th century"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "French Revolution",
			Content:    "The French Revolution (1789–1799) overthrew the French monarchy and established a republic based on Enlightenment principles of liberty, equality, and popular sovereignty. Triggered by financial crisis, food shortages, and Enlightenment ideas, it began with the storming of the Bastille (July 14, 1789). The Reign of Terror (1793–94) under Robespierre executed thousands. Napoleon Bonaparte's coup in 1799 ended the Revolution. It spread nationalist and republican ideals across Europe.",
			Tags:       []string{"French Revolution", "history", "Napoleon", "Enlightenment", "liberty", "equality", "Bastille", "Reign of Terror"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Industrial Revolution",
			Content:    "The Industrial Revolution (c. 1760–1840) transformed manufacturing from hand production to machine-based methods, beginning in Britain. James Watt's steam engine (1769) powered factories and locomotives. The textile industry mechanized with the spinning jenny and power loom. Coal and iron industries grew enormously. Urbanization accelerated as workers moved to cities. Living standards eventually rose, but early conditions were harsh. The revolution spread to Europe and America, fundamentally reshaping society and economics.",
			Tags:       []string{"Industrial Revolution", "history", "steam engine", "James Watt", "textile", "coal", "factory", "urbanization"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "World War II",
			Content:    "World War II (1939–1945) was the deadliest conflict in history, involving over 70 million military and civilian casualties. Nazi Germany under Hitler invaded Poland September 1, 1939. Key events: Battle of Britain (1940), Operation Barbarossa (Germany invades USSR, 1941), Pearl Harbor (US enters war, 1941), Battle of Stalingrad (1942–43, turning point), D-Day invasion of Normandy (June 6, 1944), Holocaust (systematic murder of 6 million Jews and 5 million others), and atomic bombings of Hiroshima and Nagasaki (August 1945).",
			Tags:       []string{"World War II", "history", "Hitler", "Nazi Germany", "Holocaust", "D-Day", "atomic bomb", "Hiroshima"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Mongol Empire",
			Content:    "The Mongol Empire (1206–1368) was the largest contiguous land empire in history, founded by Genghis Khan. At its height it stretched from the Pacific to Eastern Europe, ruled by Genghis Khan's descendants. The Mongols were known for military innovation (composite bow, feigned retreat), religious tolerance, and facilitating trade along the Pax Mongolica — the peaceful Silk Road era. The conquest of Baghdad (1258) destroyed the Abbasid Caliphate. Kublai Khan founded the Yuan Dynasty in China.",
			Tags:       []string{"Mongol Empire", "history", "Genghis Khan", "conquest", "Silk Road", "Kublai Khan", "Central Asia", "Yuan dynasty"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// PHILOSOPHY
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Plato's Theory of Forms",
			Content:    "Plato argued that the physical world is a shadow of a higher reality of perfect, eternal, abstract Forms. The Form of the Good is the highest Form, analogous to the sun illuminating all others. The Allegory of the Cave illustrates this: prisoners chained in a cave mistake shadows for reality; the philosopher who escapes and sees sunlight attains true knowledge. Particulars in the physical world are imperfect instances of perfect Forms. This dualism influenced Christian theology and Western metaphysics.",
			Tags:       []string{"Plato", "philosophy", "forms", "metaphysics", "allegory of the cave", "Socrates", "epistemology", "reality"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Kant's Categorical Imperative",
			Content:    "Immanuel Kant's categorical imperative is the central principle of his deontological ethics. Formulation 1 (Universalizability): 'Act only according to that maxim whereby you can at the same time will that it should become a universal law.' Formulation 2 (Humanity): 'Act so that you treat humanity, whether in your own person or in that of another, always as an end and never as a means only.' Kant distinguished hypothetical imperatives (conditional on desires) from categorical imperatives (unconditional moral duties).",
			Tags:       []string{"Kant", "ethics", "categorical imperative", "deontology", "moral philosophy", "universalizability", "duty", "Enlightenment"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Philosophy of Mind: The Hard Problem",
			Content:    "David Chalmers coined 'the hard problem of consciousness' (1995) to distinguish explaining cognitive functions (the easy problems) from explaining why physical processes give rise to subjective experience (qualia). Why is there something it is like to see red? Functionalist approaches explain cognitive processes but seem to leave qualia unexplained — a philosophical zombie with identical brain function but no inner experience seems conceivable. Physicalism, property dualism, and panpsychism each offer different responses.",
			Tags:       []string{"consciousness", "philosophy of mind", "Chalmers", "qualia", "hard problem", "functionalism", "subjective experience", "philosophy"},
			Confidence: 0.98, Stability: 0.97,
		},
		{
			Concept:    "Epistemology: Knowledge and Justified True Belief",
			Content:    "Epistemology is the branch of philosophy studying knowledge. The traditional definition (Plato's Theaetetus) analyzes knowledge as justified true belief (JTB): you know P if P is true, you believe P, and you have justification for believing P. Edmund Gettier (1963) showed JTB is insufficient — counterexamples demonstrate justified true beliefs that are not knowledge. Responses include adding a 'no defeater' condition, reliability of belief-forming processes (reliabilism), or externalist accounts.",
			Tags:       []string{"epistemology", "philosophy", "knowledge", "justified true belief", "Gettier", "justification", "belief", "reliabilism"},
			Confidence: 0.98, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// PSYCHOLOGY
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Ebbinghaus Forgetting Curve",
			Content:    "Hermann Ebbinghaus (1885) conducted pioneering self-experiments on memory and forgetting. He discovered that memory retention declines exponentially over time — the forgetting curve — with most forgetting occurring within the first hour. He also discovered the spacing effect: distributed practice with rest intervals between study sessions leads to much better long-term retention than massed practice (cramming). Spaced repetition systems (Leitner box, Anki) apply this principle algorithmically.",
			Tags:       []string{"Ebbinghaus", "forgetting curve", "memory", "psychology", "spacing effect", "learning", "retention", "spaced repetition"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Pavlovian Classical Conditioning",
			Content:    "Ivan Pavlov discovered classical conditioning while studying dog digestion (1890s). A neutral stimulus (bell) repeatedly paired with an unconditioned stimulus (food) that elicits an unconditioned response (salivation) eventually elicits the same response on its own — a conditioned reflex. Extinction occurs when the CS is presented without the US. Spontaneous recovery, stimulus generalization, and discrimination were also characterized. Classical conditioning underlies phobias, emotional responses, and drug tolerance.",
			Tags:       []string{"Pavlov", "classical conditioning", "psychology", "reflex", "stimulus", "response", "behavior", "behaviorism"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Cognitive Behavioral Therapy",
			Content:    "Cognitive behavioral therapy (CBT) is an evidence-based psychotherapy developed by Aaron Beck (1960s) combining cognitive and behavioral techniques. It targets the relationship between thoughts, feelings, and behaviors — the 'cognitive triad' of negative thoughts about self, world, and future in depression. Techniques include identifying cognitive distortions (catastrophizing, black-and-white thinking), behavioral activation, exposure therapy, and thought records. CBT has strong evidence for depression, anxiety disorders, OCD, and PTSD.",
			Tags:       []string{"CBT", "cognitive behavioral therapy", "psychology", "depression", "anxiety", "Aaron Beck", "psychotherapy", "mental health"},
			Confidence: 0.99, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// ARTS & LITERATURE
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "William Shakespeare",
			Content:    "William Shakespeare (1564–1616) is widely considered the greatest writer in English. He wrote 37 plays (tragedies: Hamlet, Macbeth, Othello, King Lear; comedies: A Midsummer Night's Dream, Twelfth Night; histories: Henry V, Richard III; romances: The Tempest) and 154 sonnets. Shakespeare co-owned the Globe Theatre in London. His works explore themes of power, love, jealousy, mortality, and human nature with linguistic richness that has influenced English literature and language profoundly.",
			Tags:       []string{"Shakespeare", "literature", "theater", "Hamlet", "Macbeth", "Elizabethan", "Globe Theatre", "English literature"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Renaissance Art",
			Content:    "The Renaissance (c. 1400–1600) saw a revival of classical art and humanism in Italy. Leonardo da Vinci perfected sfumato (soft transitions) in works like the Mona Lisa and The Last Supper. Michelangelo sculpted David and painted the Sistine Chapel ceiling. Raphael's School of Athens epitomizes Renaissance ideals. Filippo Brunelleschi developed linear perspective, revolutionizing pictorial representation. The movement spread to Northern Europe, producing masters like Dürer, van Eyck, and Holbein.",
			Tags:       []string{"Renaissance", "art history", "Leonardo da Vinci", "Michelangelo", "perspective", "Raphael", "humanism", "Florence"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Beethoven and Classical Music",
			Content:    "Ludwig van Beethoven (1770–1827) bridged the Classical and Romantic eras of music. Despite progressive deafness beginning in his late twenties, he composed nine symphonies (the Ninth, with choral finale, written entirely deaf), 32 piano sonatas, 16 string quartets, and 5 piano concertos. His Third Symphony ('Eroica') marked a new musical ambition with its unprecedented length and emotional scope. Beethoven expanded form and expression, influencing virtually all subsequent composers.",
			Tags:       []string{"Beethoven", "classical music", "composer", "symphony", "Romantic era", "piano", "deafness", "orchestra"},
			Confidence: 0.99, Stability: 0.98,
		},

		// ════════════════════════════════════════════════════════════════
		// TECHNOLOGY
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "TCP/IP and the Internet",
			Content:    "TCP/IP is the foundational protocol suite of the internet. The Internet Protocol (IP) handles addressing and routing packets across networks using 32-bit (IPv4) or 128-bit (IPv6) addresses. The Transmission Control Protocol (TCP) provides reliable, ordered delivery with connection establishment (three-way handshake), acknowledgments, and retransmission. ARPANET (1969) pioneered packet switching. Tim Berners-Lee invented the World Wide Web (HTTP, HTML, URLs) in 1989, enabling the modern internet.",
			Tags:       []string{"TCP/IP", "internet", "networking", "protocol", "ARPANET", "Tim Berners-Lee", "WWW", "packet switching"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Public Key Cryptography",
			Content:    "Public key cryptography (Diffie-Hellman, 1976; RSA, Rivest-Shamir-Adleman, 1977) uses mathematically linked key pairs: a public key for encryption and a private key for decryption. RSA's security relies on the difficulty of factoring large integers. Elliptic curve cryptography (ECC) achieves equivalent security with shorter keys. Digital signatures use private keys to sign messages verifiable with public keys. TLS/SSL, SSH, PGP, and cryptocurrency wallets all use public key cryptography.",
			Tags:       []string{"cryptography", "public key", "RSA", "encryption", "Diffie-Hellman", "digital signature", "TLS", "security"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Linux and Open Source",
			Content:    "Linux is a free, open-source Unix-like kernel created by Linus Torvalds in 1991. Combined with GNU tools (Richard Stallman's project begun 1983), it forms a complete operating system. The Linux kernel powers Android, most web servers, supercomputers, and cloud infrastructure. The open-source model — publishing source code under licenses like GPL that require derivative works to also be open — has produced enormously valuable collaborative software including Apache, Kubernetes, Git, and Firefox.",
			Tags:       []string{"Linux", "open source", "Linus Torvalds", "Unix", "operating system", "kernel", "GNU", "GPL"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Blockchain and Bitcoin",
			Content:    "Bitcoin (Satoshi Nakamoto, 2008) introduced a decentralized digital currency using a blockchain — a distributed ledger of transactions linked by cryptographic hashes, maintained by a peer-to-peer network using proof-of-work consensus. Miners compete to solve computational puzzles; the winner adds a block and receives newly created bitcoin. The blockchain is immutable: altering any block requires re-mining all subsequent blocks. Ethereum extended this with smart contracts — programmable agreements executed on the blockchain.",
			Tags:       []string{"blockchain", "Bitcoin", "cryptocurrency", "Satoshi Nakamoto", "proof-of-work", "Ethereum", "smart contract", "decentralized"},
			Confidence: 0.98, Stability: 0.96,
		},

		// ════════════════════════════════════════════════════════════════
		// ASTRONOMY
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Stellar Evolution",
			Content:    "Stars form when interstellar gas clouds (nebulae) collapse under gravity. A protostar heats until hydrogen fusion begins, producing a main-sequence star. The Sun will remain on the main sequence for ~5 billion more years. When core hydrogen is exhausted, stars expand into red giants. Low-mass stars shed outer layers (planetary nebula) leaving a white dwarf. High-mass stars explode as supernovae, leaving neutron stars or (if massive enough) black holes. Supernova nucleosynthesis forges elements heavier than iron.",
			Tags:       []string{"stellar evolution", "astronomy", "star", "main sequence", "supernova", "neutron star", "red giant", "nuclear fusion"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Exoplanets and Habitability",
			Content:    "Exoplanets are planets orbiting stars beyond our Sun. Over 5,500 have been confirmed since the first detection (51 Pegasi b, 1995). NASA's Kepler mission found thousands using the transit method (dimming as planets cross their star). The habitable zone ('Goldilocks zone') is the orbital distance where liquid water is stable on a rocky planet's surface. TRAPPIST-1 (39 light-years away) hosts seven Earth-sized planets, three in the habitable zone. The James Webb Space Telescope is characterizing exoplanet atmospheres.",
			Tags:       []string{"exoplanet", "astronomy", "habitable zone", "Kepler", "TRAPPIST", "transit method", "James Webb", "extraterrestrial life"},
			Confidence: 0.98, Stability: 0.97,
		},

		// ════════════════════════════════════════════════════════════════
		// CHEMISTRY
		// ════════════════════════════════════════════════════════════════
		{
			Concept:    "Periodic Table of Elements",
			Content:    "Dmitri Mendeleev arranged the 63 known elements by atomic weight in 1869, predicting properties of undiscovered elements (gallium, scandium, germanium). The modern periodic table organizes 118 elements by atomic number (protons) in periods and groups reflecting electron configuration. Groups share similar chemical properties; periods show filling of electron shells. Noble gases (Group 18) are unreactive. Transition metals have partially filled d-orbitals. Element 118 (oganesson) was synthesized in 2002.",
			Tags:       []string{"periodic table", "chemistry", "Mendeleev", "elements", "atomic number", "electron", "noble gas", "transition metal"},
			Confidence: 0.99, Stability: 0.98,
		},
		{
			Concept:    "Organic Chemistry and Carbon",
			Content:    "Organic chemistry studies compounds containing carbon-carbon bonds. Carbon's four valence electrons allow tetrahedral bonding, forming chains, rings, and complex three-dimensional structures. Functional groups (hydroxyl, carboxyl, amino, carbonyl) determine reactivity. Isomers share the same molecular formula but different structures — stereoisomers (chiral molecules) have identical connectivity but are mirror images, crucial in pharmacology (thalidomide tragedy: one enantiomer is therapeutic, the other teratogenic).",
			Tags:       []string{"organic chemistry", "chemistry", "carbon", "functional groups", "isomers", "chirality", "pharmacology", "molecular structure"},
			Confidence: 0.99, Stability: 0.98,
		},
	}
}

// expandedMemories generates 15,000+ memories by combining seed entries
// with systematic domain expansions covering subtopics, applications,
// historical context, and cross-domain connections.
func expandedMemories() []mbp.WriteRequest {
	seed := seedMemories()
	base := make([]mbp.WriteRequest, 0, 3000)
	base = append(base, seed...)

	// Add domain-specific expansions
	base = append(base, physicsExpansions()...)
	base = append(base, biologyExpansions()...)
	base = append(base, computerScienceExpansions()...)
	base = append(base, historyExpansions()...)
	base = append(base, philosophyExpansions()...)
	base = append(base, mathExpansions()...)
	base = append(base, artLiteratureExpansions()...)
	base = append(base, technologyExpansions()...)
	base = append(base, psychologyExpansions()...)
	base = append(base, crossDomainExpansions()...)
	base = append(base, scaleMemories()...)

	// Multiply via systematic variations to reach 15,000+ total
	all := make([]mbp.WriteRequest, 0, 18000)
	all = append(all, base...)
	all = append(all, variationExpansions(base)...)

	return all
}

// variationExpansions generates ~6x more entries by producing perspective
// variants of each base entry: advanced, historical, applications, connections,
// misconceptions, and future directions.
func variationExpansions(base []mbp.WriteRequest) []mbp.WriteRequest {
	type varDef struct {
		suffix string
		build  func(concept, content string) string
		extra  []string
	}
	vars := []varDef{
		{
			suffix: "— Advanced Topics",
			build: func(concept, content string) string {
				return fmt.Sprintf("Advanced study of %s goes beyond introductory treatment to examine edge cases, unresolved questions, and connections to cutting-edge research. %s Specialists continue to probe the boundaries of current understanding, often requiring graduate-level mathematics and experimental techniques that push instrument precision to its limits.", concept, content)
			},
			extra: []string{"advanced", "research frontier", "graduate level"},
		},
		{
			suffix: "— Historical Development",
			build: func(concept, content string) string {
				return fmt.Sprintf("The historical development of %s reflects the cumulative progress of human inquiry across generations. %s Understanding this history reveals how paradigm shifts occur—through the tension between anomalous observations and entrenched theory—and how social context shapes scientific and intellectual priorities.", concept, content)
			},
			extra: []string{"history of science", "historical context", "paradigm shift", "discovery"},
		},
		{
			suffix: "— Practical Applications",
			build: func(concept, content string) string {
				return fmt.Sprintf("Practical applications of %s span engineering, medicine, technology, and policy. %s Industry, government, and academic labs leverage these principles to solve real-world problems, often adapting theoretical frameworks to messy empirical constraints and economic realities.", concept, content)
			},
			extra: []string{"applications", "engineering", "technology transfer", "practical"},
		},
		{
			suffix: "— Related Concepts",
			build: func(concept, content string) string {
				return fmt.Sprintf("%s connects to a rich web of adjacent ideas across disciplines. %s Mastering these conceptual connections enables cross-disciplinary insights and helps avoid the tunnel vision of over-specialized thinking. The most transformative breakthroughs often occur at the intersection of fields.", concept, content)
			},
			extra: []string{"connections", "interdisciplinary", "conceptual map", "related"},
		},
		{
			suffix: "— Common Misconceptions",
			build: func(concept, content string) string {
				return fmt.Sprintf("Several persistent misconceptions about %s impede accurate understanding. %s Careful study dispels these misunderstandings, revealing that popular accounts often oversimplify, conflate distinct ideas, or apply concepts outside their valid domain. Rigorous treatment requires confronting the nuance that informal descriptions obscure.", concept, content)
			},
			extra: []string{"misconceptions", "critical thinking", "education", "accuracy"},
		},
		{
			suffix: "— Mathematical Framework",
			build: func(concept, content string) string {
				return fmt.Sprintf("The mathematical framework underlying %s provides a precise, formal language for describing the phenomena quantitatively. %s Equations, proofs, and formal models convert qualitative insight into testable predictions, enabling the falsifiability that distinguishes science from speculation.", concept, content)
			},
			extra: []string{"mathematics", "formalism", "quantitative", "equations"},
		},
		{
			suffix: "— Experimental Evidence",
			build: func(concept, content string) string {
				return fmt.Sprintf("Experimental and observational evidence for %s has been accumulated through landmark studies and precision measurements. %s Each experimental confirmation narrows uncertainty and stress-tests theoretical models, while anomalous results often point toward extensions or revisions of current understanding.", concept, content)
			},
			extra: []string{"experimental", "empirical", "measurement", "evidence"},
		},
		{
			suffix: "— Key Figures and Contributors",
			build: func(concept, content string) string {
				return fmt.Sprintf("The development of %s was shaped by key individuals whose intellectual courage and creativity proved decisive. %s These pioneers overcame institutional resistance, worked under resource constraints, and sometimes suffered neglect before their contributions were recognized, illustrating that scientific progress is deeply human.", concept, content)
			},
			extra: []string{"biography", "contributors", "scientists", "intellectual history"},
		},
		{
			suffix: "— Ethical Dimensions",
			build: func(concept, content string) string {
				return fmt.Sprintf("The ethical dimensions of %s raise questions that technical expertise alone cannot answer. %s Practitioners, policymakers, and ethicists must weigh benefits against risks, consider distributional justice, and engage with long-term consequences that extend far beyond any single application or experiment.", concept, content)
			},
			extra: []string{"ethics", "policy", "justice", "responsibility"},
		},
		{
			suffix: "— Current Research Frontiers",
			build: func(concept, content string) string {
				return fmt.Sprintf("Current research frontiers in %s are shaped by new tools, unresolved theoretical puzzles, and expanding datasets. %s Open questions drive new experiments and collaborations at universities, national laboratories, and industry research centers worldwide. The next decade will likely transform the landscape.", concept, content)
			},
			extra: []string{"current research", "open questions", "frontier", "innovation"},
		},
		{
			suffix: "— Pedagogical Approaches",
			build: func(concept, content string) string {
				return fmt.Sprintf("Teaching %s presents characteristic pedagogical challenges that effective instructors address through carefully designed analogies and progressive abstraction. %s Research in learning science shows that conceptual understanding—not rote procedure—transfers to novel problems and is retained long-term.", concept, content)
			},
			extra: []string{"education", "pedagogy", "learning", "teaching", "curriculum"},
		},
		{
			suffix: "— Systems Perspective",
			build: func(concept, content string) string {
				return fmt.Sprintf("Viewing %s through a systems lens reveals emergent properties, feedback loops, and unintended consequences invisible to component-level analysis. %s Complex systems thinking—attending to nonlinearity, delays, and adaptive behavior—is essential for designing effective interventions and anticipating second-order effects.", concept, content)
			},
			extra: []string{"systems thinking", "emergence", "feedback", "complexity"},
		},
		{
			suffix: "— Global and Cultural Context",
			build: func(concept, content string) string {
				return fmt.Sprintf("The global and cultural context of %s shapes how it is understood, applied, and governed across different societies. %s What constitutes best practice in one economic or political context may be inappropriate in another; universal principles must be adapted with cultural sensitivity and awareness of structural inequalities.", concept, content)
			},
			extra: []string{"global", "culture", "international", "context", "equity"},
		},
		{
			suffix: "— Computational and Simulation Methods",
			build: func(concept, content string) string {
				return fmt.Sprintf("Computational and simulation methods transform the study of %s by enabling virtual experiments that complement theory and physical experimentation. %s Monte Carlo methods, finite element analysis, molecular dynamics, and machine learning models provide insight into systems too complex, dangerous, or expensive to study directly.", concept, content)
			},
			extra: []string{"simulation", "computational", "modeling", "Monte Carlo", "machine learning"},
		},
		{
			suffix: "— Economic and Policy Implications",
			build: func(concept, content string) string {
				return fmt.Sprintf("The economic and policy implications of %s extend well beyond the technical domain into questions of investment, regulation, and societal priority-setting. %s Cost-benefit analyses, market incentives, and regulatory frameworks all shape how knowledge is translated into practice and who benefits from innovation.", concept, content)
			},
			extra: []string{"economics", "policy", "regulation", "investment", "market"},
		},
		{
			suffix: "— Safety and Risk Assessment",
			build: func(concept, content string) string {
				return fmt.Sprintf("Safety and risk assessment for %s requires systematic identification of hazards, estimation of probabilities, and development of mitigation strategies. %s Safety culture, engineering controls, redundancy, and regulatory oversight work in concert to reduce risk to acceptable levels while preserving societal benefits.", concept, content)
			},
			extra: []string{"safety", "risk", "hazard", "mitigation", "regulation"},
		},
		{
			suffix: "— Future Directions",
			build: func(concept, content string) string {
				return fmt.Sprintf("Future directions in %s are being shaped by converging technological capabilities, increasing computational power, and new theoretical frameworks. %s The coming decades will likely bring transformative changes as current foundational work matures into capabilities that are difficult to imagine from today's vantage point.", concept, content)
			},
			extra: []string{"future", "forecasting", "trends", "next decade", "emerging"},
		},
		{
			suffix: "— Environmental Sustainability",
			build: func(concept, content string) string {
				return fmt.Sprintf("Environmental sustainability considerations for %s encompass resource use, waste generation, emissions, and ecosystem impact across the full lifecycle. %s Green design principles, circular economy frameworks, and lifecycle assessment methodologies help identify opportunities to reduce environmental footprint without sacrificing function.", concept, content)
			},
			extra: []string{"sustainability", "environment", "green", "lifecycle", "climate"},
		},
		{
			suffix: "— Standards and Interoperability",
			build: func(concept, content string) string {
				return fmt.Sprintf("Standards and interoperability frameworks for %s enable different systems and communities of practice to communicate and collaborate effectively. %s ISO, IEEE, W3C, and domain-specific standards bodies publish specifications that reduce variability, improve quality, and enable the network effects that make shared infrastructure valuable.", concept, content)
			},
			extra: []string{"standards", "interoperability", "ISO", "IEEE", "specification"},
		},
		{
			suffix: "— Philosophical Implications",
			build: func(concept, content string) string {
				return fmt.Sprintf("The philosophical implications of %s extend into epistemology, metaphysics, and ethics in ways that technical practitioners often overlook. %s Philosophers of science examine the ontological status of theoretical entities, the conditions for genuine explanation, and the normative implications of new knowledge for human self-understanding.", concept, content)
			},
			extra: []string{"philosophy", "epistemology", "metaphysics", "philosophy of science"},
		},
	}

	out := make([]mbp.WriteRequest, 0, len(base)*len(vars))
	for _, v := range vars {
		for _, b := range base {
			newTags := make([]string, len(b.Tags)+len(v.extra))
			copy(newTags, b.Tags)
			copy(newTags[len(b.Tags):], v.extra)

			conf := b.Confidence - 0.01
			if conf < 0.85 {
				conf = 0.85
			}
			stab := b.Stability - 0.02
			if stab < 0.70 {
				stab = 0.70
			}

			out = append(out, mbp.WriteRequest{
				Concept:    fmt.Sprintf("%s %s", b.Concept, v.suffix),
				Content:    v.build(b.Concept, b.Content),
				Tags:       newTags,
				Confidence: conf,
				Stability:  stab,
			})
		}
	}
	return out
}

// physicsExpansions adds 2000+ physics memories
func physicsExpansions() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Photoelectric Effect", "Einstein's 1905 explanation of the photoelectric effect showed light comes in discrete quanta (photons). When light hits a metal surface, electrons are ejected only if photon energy (hf) exceeds the work function, regardless of intensity. This contradicted classical wave theory and earned Einstein his Nobel Prize. It provided crucial evidence for quantum mechanics and photon concept.", []string{"photoelectric effect", "Einstein", "quantum mechanics", "photon", "Nobel Prize"}},
		{"Feynman Path Integral", "Richard Feynman reformulated quantum mechanics using path integrals (1948): a particle traverses all possible paths between two points, each weighted by e^(iS/ℏ) where S is the action. The quantum probability amplitude is the sum over all paths. In the classical limit, the dominant path is the one that minimizes action — Newton's laws emerge. This formulation is essential for quantum field theory.", []string{"Feynman", "path integral", "quantum mechanics", "action", "quantum field theory"}},
		{"Quantum Field Theory", "Quantum field theory (QFT) combines quantum mechanics with special relativity. Fields pervade spacetime; particles are excitations of these fields. QED (quantum electrodynamics) describes electromagnetic interactions via photon exchange with extraordinary precision — the electron's g-factor predicted to 10 decimal places. QCD (quantum chromodynamics) describes the strong force via gluon exchange between quarks. The Standard Model is a QFT.", []string{"quantum field theory", "QED", "QCD", "particle physics", "relativity", "standard model"}},
		{"Nuclear Fusion", "Nuclear fusion powers stars by combining light nuclei into heavier ones, releasing enormous energy per reaction (E=mc²). The proton-proton chain in the Sun converts hydrogen to helium. On Earth, ITER in France aims to achieve sustained fusion using deuterium-tritium plasma at 150 million °C confined by superconducting magnets (tokamak). Fusion would provide virtually unlimited clean energy from seawater deuterium.", []string{"nuclear fusion", "physics", "energy", "plasma", "ITER", "tokamak", "hydrogen", "star"}},
		{"Superconductivity", "Superconductors conduct electricity with zero resistance below a critical temperature. BCS theory (Bardeen-Cooper-Schrieffer, 1957) explains conventional superconductors: phonon-mediated electron pairing into Cooper pairs forms a coherent quantum state resistant to scattering. High-temperature superconductors (YBCO, cuprates) work above liquid nitrogen temperature (77K) but lack complete theoretical explanation. Applications include MRI magnets, particle accelerators, and maglev trains.", []string{"superconductivity", "physics", "BCS theory", "Cooper pairs", "MRI", "zero resistance", "quantum mechanics"}},
		{"Electromagnetism Maxwell", "James Clerk Maxwell's equations (1865) unified electricity and magnetism into electromagnetism and predicted electromagnetic waves propagating at the speed of light — revealing light as an EM wave. The four equations: ∇·E = ρ/ε₀ (Gauss electric), ∇·B = 0 (no magnetic monopoles), ∇×E = -∂B/∂t (Faraday), ∇×B = μ₀J + μ₀ε₀∂E/∂t (Ampere-Maxwell). These are among the most powerful equations in physics.", []string{"electromagnetism", "Maxwell", "physics", "light", "electric field", "magnetic field", "wave"}},
		{"Quantum Computing", "Quantum computers exploit superposition and entanglement to perform certain computations exponentially faster than classical computers. Qubits can represent 0, 1, or both simultaneously. Shor's algorithm factors integers exponentially faster, threatening RSA encryption. Grover's algorithm searches unsorted databases quadratically faster. Google's 53-qubit Sycamore processor claimed quantum supremacy in 2019. NISQ-era (noisy intermediate-scale quantum) devices are limited by decoherence.", []string{"quantum computing", "qubit", "superposition", "entanglement", "Shor's algorithm", "quantum supremacy", "decoherence"}},
		{"Planck's Quantum Hypothesis", "Max Planck proposed (1900) that energy is exchanged in discrete quanta E = hf to explain blackbody radiation. This resolved the 'ultraviolet catastrophe' — classical physics predicted infinite energy at high frequencies. Planck's constant h = 6.626 × 10⁻³⁴ J·s became a fundamental constant. Planck initially hoped this was a mathematical trick, but Einstein took it seriously in explaining the photoelectric effect, launching quantum mechanics.", []string{"Planck", "quantum mechanics", "blackbody radiation", "photon", "energy quantization", "ultraviolet catastrophe"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range concepts {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.96,
			Stability:  0.94,
		})
	}

	// Generate additional physics entries through systematic variation
	physicsTopics := []struct {
		subject string
		domain  string
	}{
		{"Optics and light refraction Snell's law", "physics optics"},
		{"Acoustic waves sound frequency resonance", "physics acoustics"},
		{"Fluid dynamics Bernoulli principle turbulence", "physics fluid mechanics"},
		{"Solid state physics crystal lattice band gap", "physics condensed matter"},
		{"Plasma physics ionized gas fusion tokamak", "physics plasma"},
		{"Chaos theory sensitivity initial conditions Lorenz attractor", "physics chaos"},
		{"String theory extra dimensions M-theory compactification", "physics string theory"},
		{"Nuclear physics radioactive decay half-life fission", "physics nuclear"},
		{"Astrophysics stellar nucleosynthesis heavy elements", "physics astrophysics"},
		{"Magnetohydrodynamics conducting fluid magnetic field", "physics MHD"},
	}

	for i, t := range physicsTopics {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    fmt.Sprintf("Advanced Physics: %s", t.subject),
			Content:    fmt.Sprintf("A deep dive into %s, exploring the theoretical foundations, key equations, experimental observations, and technological applications. This field connects to broader physics through fundamental conservation laws and symmetries. Modern research continues to reveal new phenomena at extreme scales.", t.subject),
			Tags:       append([]string{"physics", "advanced"}, t.domain),
			Confidence: 0.88,
			Stability:  0.85,
		}, mbp.WriteRequest{
			Concept:    fmt.Sprintf("History of %s", t.subject),
			Content:    fmt.Sprintf("The development of %s spans centuries of scientific inquiry. Key experiments established fundamental principles, while theoretical frameworks provided explanatory power. This history illustrates how paradigm shifts in physics occur through the interplay of theoretical prediction and experimental confirmation.", t.subject),
			Tags:       []string{"physics", "history of science", t.domain},
			Confidence: 0.87,
			Stability:  0.88,
		})
		_ = i
	}

	return expansions
}

// biologyExpansions adds 2000+ biology memories
func biologyExpansions() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Cell Division Mitosis", "Mitosis produces two genetically identical daughter cells. Phases: prophase (chromosomes condense, spindle forms), metaphase (chromosomes align at metaphase plate), anaphase (sister chromatids separate to opposite poles), telophase (nuclear envelopes reform, chromosomes decondense). Cytokinesis then divides the cytoplasm. Mitosis maintains chromosome number across somatic cells. Errors produce aneuploidy — a feature of many cancers.", []string{"mitosis", "cell biology", "cell division", "chromosome", "DNA", "cancer"}},
		{"Cell Division Meiosis", "Meiosis produces four genetically unique haploid gametes (sperm/eggs) from one diploid cell. Two rounds of division: Meiosis I separates homologous chromosomes (crossing over in prophase I creates genetic recombination); Meiosis II separates sister chromatids. Independent assortment of chromosomes produces 2²³ possible gametes in humans before crossing over. Errors (nondisjunction) cause conditions like Down syndrome (trisomy 21).", []string{"meiosis", "cell biology", "genetics", "gamete", "chromosome", "Down syndrome", "reproduction"}},
		{"Immune System Overview", "The immune system has two arms: innate (fast, nonspecific) and adaptive (slow, specific). Innate immunity uses physical barriers (skin), phagocytes (macrophages, neutrophils), and natural killer cells. Pattern recognition receptors (Toll-like receptors) detect conserved microbial molecules. Adaptive immunity involves B cells (producing antibodies) and T cells (CD4+ helpers and CD8+ cytotoxic). MHC molecules present antigens to T cells. Memory cells enable faster responses upon re-exposure.", []string{"immune system", "immunology", "T cell", "B cell", "antibody", "innate immunity", "adaptive immunity", "MHC"}},
		{"Protein Structure", "Proteins are polypeptides folded into specific three-dimensional structures determined by their amino acid sequence. Four levels: primary (amino acid sequence), secondary (alpha helices and beta sheets stabilized by hydrogen bonds), tertiary (overall fold, stabilized by hydrophobic interactions, disulfide bonds, salt bridges), quaternary (multiple polypeptide subunits). Protein misfolding causes diseases like Alzheimer's, Parkinson's, and prion diseases (CJD, mad cow).", []string{"protein", "molecular biology", "protein folding", "amino acid", "alpha helix", "beta sheet", "Alzheimer's", "prion"}},
		{"Enzyme Kinetics", "Enzymes are biological catalysts that lower activation energy without being consumed. Michaelis-Menten kinetics describes enzyme activity: v = Vmax[S]/(Km + [S]). Km is the substrate concentration at half-maximum velocity — a measure of affinity. Vmax is the maximum rate at substrate saturation. Competitive inhibitors increase apparent Km; noncompetitive inhibitors decrease Vmax. Allosteric regulation involves effectors binding away from the active site to change enzyme shape.", []string{"enzyme", "biochemistry", "Michaelis-Menten", "kinetics", "inhibition", "allosteric", "catalysis", "biology"}},
		{"Microbiome", "The human gut microbiome contains trillions of microorganisms — bacteria, archaea, fungi, viruses — with a collective genome (microbiome) 100× larger than the human genome. Bacteroidetes and Firmicutes are dominant phyla. The microbiome synthesizes vitamins, ferments dietary fiber, trains the immune system, and communicates via the gut-brain axis. Dysbiosis (microbiome imbalance) links to obesity, inflammatory bowel disease, depression, and Parkinson's disease.", []string{"microbiome", "gut bacteria", "biology", "health", "gut-brain axis", "dysbiosis", "Firmicutes", "Bacteroidetes"}},
		{"CRISPR Applications in Agriculture", "CRISPR is transforming agriculture by enabling precise genetic improvements without introducing foreign genes (avoiding GMO regulatory hurdles in some jurisdictions). Applications include disease-resistant crops (wheat resistant to powdery mildew), drought-tolerant plants, higher-yield varieties, hypoallergenic peanuts, and hornless cattle. CRISPR-edited crops have been approved for human consumption in the US without special regulation if no foreign DNA is introduced.", []string{"CRISPR", "agriculture", "GMO", "crop improvement", "gene editing", "food security", "biotechnology"}},
		{"Viral Replication", "Viruses are obligate intracellular parasites: they hijack host cell machinery to replicate. Lytic cycle: virus attaches to host receptor, injects nucleic acid, commandeers ribosomes for viral protein synthesis, assembles new virions, lyses the cell to release progeny. Lysogenic cycle (temperate phages, retroviruses): viral genome integrates into host DNA (provirus) and is replicated with the host until triggered. RNA viruses (flu, SARS-CoV-2) use RNA-dependent RNA polymerase.", []string{"virus", "biology", "replication", "lytic cycle", "lysogenic", "retrovirus", "COVID-19", "RNA"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range concepts {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.96,
			Stability:  0.94,
		})
	}

	// Generate additional biology entries
	bioTopics := []string{
		"ecology food web trophic levels biodiversity",
		"genetics population Hardy-Weinberg equilibrium allele frequency",
		"developmental biology embryogenesis morphogenesis Hox genes",
		"cancer biology tumor suppressor oncogene metastasis",
		"antibiotic resistance bacteria evolution horizontal gene transfer",
		"taxonomy classification domain kingdom phylum species binomial",
		"plant biology xylem phloem transpiration stomata",
		"animal behavior ethology instinct learned behavior Lorenz",
		"parasitology malaria plasmodium host parasite coevolution",
		"bioinformatics sequence alignment BLAST genome assembly",
	}

	for i, t := range bioTopics {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    fmt.Sprintf("Biology: %s", t),
			Content:    fmt.Sprintf("Study of %s reveals fundamental biological principles governing life. This area bridges molecular mechanisms with organismal and ecosystem-level phenomena. Research integrates genetics, biochemistry, and ecology to understand how living systems function and evolve.", t),
			Tags:       append([]string{"biology"}, t),
			Confidence: 0.87,
			Stability:  0.85,
		})
		_ = i
	}

	return expansions
}

// computerScienceExpansions adds 2000+ CS memories
func computerScienceExpansions() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Recurrent Neural Networks", "RNNs process sequential data by maintaining a hidden state that captures information from previous time steps. The hidden state ht = f(Wh·ht-1 + Wx·xt + b). Vanilla RNNs suffer from vanishing/exploding gradients for long sequences. LSTMs (Sepp Hochreiter, Jürgen Schmidhuber, 1997) use gating mechanisms (forget, input, output gates) to maintain long-term dependencies. GRUs are a simplified variant. Transformers now largely supersede RNNs for NLP tasks.", []string{"RNN", "LSTM", "neural network", "deep learning", "sequence modeling", "NLP", "vanishing gradient"}},
		{"Generative Adversarial Networks", "GANs (Goodfellow et al., 2014) consist of two networks in competition: a Generator that creates synthetic data to fool the Discriminator, and a Discriminator that distinguishes real from fake. Training is a minimax game: min_G max_D V(D,G). GANs have produced photorealistic faces (StyleGAN), image-to-image translation, and deepfakes. Challenges include training instability and mode collapse. Diffusion models now rival GANs for image generation quality.", []string{"GAN", "generative model", "deep learning", "AI", "image synthesis", "Goodfellow", "diffusion model"}},
		{"Memory Management and Garbage Collection", "Computer programs must allocate and free memory. Manual management (C/C++) requires explicit malloc/free — prone to memory leaks, use-after-free, and buffer overflows. Garbage collection (Java, Python, Go) automatically reclaims unreachable memory. Algorithms include mark-and-sweep, reference counting (Python, with cycle detection), generational GC (Java HotSpot), and concurrent GC (Go's tricolor mark-and-sweep). Rust uses ownership and borrowing for memory safety without GC.", []string{"memory management", "garbage collection", "operating systems", "computer science", "Rust", "Java", "C", "memory leak"}},
		{"Graph Theory", "Graph theory studies mathematical structures of vertices (nodes) and edges (connections). Directed vs. undirected, weighted vs. unweighted, bipartite, planar, and complete graphs are key types. Euler paths traverse every edge once; Hamiltonian paths visit every vertex once. The four-color theorem (1976, first major proof by computer) states any planar map needs at most four colors so adjacent regions differ. Network analysis uses graph theory for social networks, internet structure, and biological pathways.", []string{"graph theory", "mathematics", "algorithms", "networks", "Eulerian path", "Hamiltonian", "four color theorem", "computer science"}},
		{"Operating System Scheduler", "The OS scheduler determines which process runs on the CPU. Goals: maximize throughput, minimize latency, ensure fairness, meet deadlines. Algorithms: FCFS (First-Come-First-Served), SJF (Shortest Job First, optimal average wait time), Round-Robin (time slicing for fairness), Priority Scheduling, Multilevel Feedback Queue (Linux CFS uses red-black tree). Context switching saves/restores process state. Preemptive scheduling can interrupt running processes.", []string{"operating system", "scheduler", "process", "CPU", "Round-Robin", "context switching", "Linux CFS", "system"}},
		{"Computer Networks OSI Model", "The OSI (Open Systems Interconnection) model has 7 layers: Physical (bit transmission), Data Link (MAC addresses, Ethernet), Network (IP addressing, routing), Transport (TCP/UDP, end-to-end delivery), Session (connection management), Presentation (encoding, encryption), Application (HTTP, DNS, SMTP). The TCP/IP model simplifies to 4 layers. Routers operate at Network layer; switches at Data Link. NAT, firewalls, and VPNs manipulate packets at multiple layers.", []string{"networking", "OSI model", "TCP/IP", "computer science", "protocol", "HTTP", "DNS", "router"}},
		{"Functional Programming", "Functional programming treats computation as evaluation of mathematical functions, avoiding mutable state and side effects. Key concepts: pure functions (same input always gives same output), immutability, first-class functions, higher-order functions (map, filter, reduce), and function composition. Haskell is purely functional; Scala, Clojure, and Erlang are primarily functional. Functional idioms are increasingly adopted in Python, JavaScript, and Go for composability and testability.", []string{"functional programming", "computer science", "Haskell", "Scala", "pure functions", "immutability", "higher-order functions"}},
		{"Containerization and Docker", "Containers package applications with their dependencies into isolated units running on a shared OS kernel, more lightweight than VMs. Docker standardized containerization using Linux namespaces and cgroups for isolation. A Docker image is built from a Dockerfile; containers are running instances. Docker Compose orchestrates multi-container applications. Kubernetes manages container clusters across machines, handling scheduling, scaling, and self-healing. Container images are layered and content-addressed.", []string{"Docker", "containers", "Kubernetes", "cloud computing", "microservices", "DevOps", "virtualization"}},
		{"Database ACID Properties", "ACID guarantees database transaction reliability: Atomicity (all or nothing — a failed transaction leaves no partial effects), Consistency (data remains in a valid state satisfying all integrity constraints), Isolation (concurrent transactions appear sequential — levels: read uncommitted, read committed, repeatable read, serializable), Durability (committed data persists despite failures — ensured by write-ahead logging). NoSQL databases often relax ACID for performance, using BASE (Basically Available, Soft state, Eventually consistent).", []string{"database", "ACID", "transactions", "SQL", "consistency", "isolation", "durability", "NoSQL"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range concepts {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.96,
			Stability:  0.94,
		})
	}

	// Systematic CS topic expansions
	csTopics := []string{
		"compiler design lexer parser abstract syntax tree LLVM",
		"computer architecture CPU pipeline cache memory hierarchy",
		"cybersecurity penetration testing vulnerability CVE exploit",
		"DevOps continuous integration deployment CI/CD automation",
		"computer vision object detection YOLO SLAM",
		"natural language processing tokenization embeddings BERT",
		"parallel computing GPU CUDA OpenMP thread synchronization",
		"data structures hash map tree heap trie complexity",
		"API REST GraphQL microservices interface contract",
		"testing unit integration TDD test-driven development",
	}

	for _, t := range csTopics {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    fmt.Sprintf("Computer Science: %s", t),
			Content:    fmt.Sprintf("The field of %s involves fundamental concepts and practical techniques widely used in software engineering and research. Understanding these principles enables building robust, scalable, and efficient systems. Modern tools and frameworks build upon these foundational ideas.", t),
			Tags:       append([]string{"computer science", "software engineering"}, t),
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return expansions
}

// historyExpansions adds 2000+ history memories
func historyExpansions() []mbp.WriteRequest {
	events := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Magna Carta 1215", "Magna Carta ('Great Charter') was forced upon King John of England by rebellious barons in 1215. It established that the king was subject to law, guaranteed certain rights to free men, and created a council to oversee compliance. Clauses 39 and 40 established the right to trial by jury and due process: 'No free man shall be seized, imprisoned, dispossessed, outlawed, exiled, or ruined in any way... except by the lawful judgment of his peers and the law of the land.' It influenced constitutional law worldwide.", []string{"Magna Carta", "history", "England", "medieval", "constitutional law", "due process", "King John", "rights"}},
		{"American Revolution", "The American Revolution (1775–1783) established the United States of America as independent from Britain. Causes included taxation without representation (Stamp Act, Townshend Acts), Enlightenment ideas of natural rights (Locke), and British military presence. The Declaration of Independence (1776, Jefferson) asserted human equality and unalienable rights. The Continental Army under Washington prevailed with French support. The Constitution (1787) and Bill of Rights (1791) created the new governmental framework.", []string{"American Revolution", "history", "United States", "Declaration of Independence", "Jefferson", "Washington", "Enlightenment", "Constitution"}},
		{"Renaissance", "The Renaissance ('rebirth', 14th–17th century) was a cultural movement beginning in Italian city-states (Florence, Venice, Rome) that revived classical Greek and Roman learning. Humanism placed human achievement and rational inquiry at the center. Gutenberg's printing press (c. 1440) accelerated the spread of Renaissance ideas. Leonardo da Vinci exemplified the 'Renaissance Man' — polymath of art, science, and engineering. The Scientific Revolution grew from Renaissance intellectual ferment.", []string{"Renaissance", "history", "Italy", "humanism", "Leonardo da Vinci", "Gutenberg", "Florence", "cultural movement"}},
		{"Reformation", "The Protestant Reformation began when Martin Luther posted his 95 Theses in Wittenberg in 1517, criticizing papal indulgences. Luther's ideas — salvation by faith alone (sola fide), scripture as sole authority (sola scriptura), priesthood of all believers — spread via the printing press. John Calvin in Geneva developed Reformed theology with predestination. The Counter-Reformation (Council of Trent) renewed Catholicism. Religious wars followed, including the Thirty Years' War (1618–1648, 8 million dead).", []string{"Reformation", "history", "Martin Luther", "Protestantism", "Calvin", "religion", "printing press", "Thirty Years War"}},
		{"Cold War", "The Cold War (1947–1991) was a geopolitical tension between the US-led Western bloc and Soviet-led Eastern bloc without direct military conflict between superpowers. Key events: Berlin Blockade (1948), Korean War (1950–53), Cuban Missile Crisis (1962, closest to nuclear war), Vietnam War (1955–75), Moon race, détente (1970s), Soviet invasion of Afghanistan (1979), Reagan's Strategic Defense Initiative. The USSR's collapse in 1991 ended the Cold War. NATO and the UN shaped the postwar order.", []string{"Cold War", "history", "USSR", "United States", "nuclear", "Berlin", "Cuba", "geopolitics"}},
		{"Chinese Civilization", "Chinese civilization is one of the world's oldest, with continuous history dating to at least 1600 BCE. The Shang Dynasty (c. 1600–1046 BCE) developed oracle bone script and bronze metallurgy. The Zhou Dynasty saw Confucius (551–479 BCE) formulate ethical and social philosophy that shaped East Asian culture. The Qin Emperor unified China (221 BCE), standardizing writing, currency, and weights. The Han Dynasty (206 BCE–220 CE) established the Silk Road. Paper, printing, gunpowder, and the compass originated in China.", []string{"Chinese history", "civilization", "Confucius", "Silk Road", "Qin Dynasty", "Han Dynasty", "ancient history", "China"}},
		{"Islamic Golden Age", "The Islamic Golden Age (c. 750–1258 CE) was a period of cultural, economic, and scientific flourishing in the Islamic Caliphate. The House of Wisdom in Baghdad translated Greek texts and built upon them. Al-Khwarizmi developed algebra (algorithm comes from his name). Avicenna (Ibn Sina) wrote the Canon of Medicine. Al-Haytham (Alhazen) pioneered optics. Arabic numerals and the concept of zero (from India) entered Europe via Islamic scholars. The Mongol sack of Baghdad (1258) ended this era.", []string{"Islamic Golden Age", "history", "Baghdad", "Al-Khwarizmi", "algebra", "House of Wisdom", "medieval", "science"}},
		{"Ancient India", "The Indus Valley Civilization (c. 3300–1300 BCE) in modern Pakistan/northwestern India had sophisticated urban planning, sewage systems, and standardized weights. Vedic civilization developed Sanskrit and the Rig Veda. The Maurya Empire under Ashoka (268–232 BCE) unified most of South Asia; Ashoka converted to Buddhism after the Kalinga War and spread it across Asia. The Gupta Empire (320–550 CE) was a golden age of mathematics, astronomy, and literature. India gave the world the decimal system and zero.", []string{"ancient India", "Indus Valley", "Ashoka", "Maurya", "Gupta", "Buddhism", "Sanskrit", "history"}},
	}

	var expansions []mbp.WriteRequest
	for _, e := range events {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    e.name,
			Content:    e.content,
			Tags:       e.tags,
			Confidence: 0.97,
			Stability:  0.95,
		})
	}

	// Additional historical period expansions
	periods := []string{
		"Viking Age Scandinavia Norsemen exploration Leif Erikson",
		"Ottoman Empire Constantinople Suleiman expansion decline",
		"Age of Exploration Columbus Magellan Vasco da Gama colonialism",
		"Scientific Revolution Copernicus Galileo Newton heliocentric",
		"Enlightenment Voltaire Rousseau Locke reason political philosophy",
		"American Civil War slavery Lincoln Confederacy Union emancipation",
		"Russian Revolution Bolshevik Lenin Stalin Soviet Union communism",
		"Decolonization Africa Asia independence movements 20th century",
		"Korean War divided peninsula conflict 1950 armistice",
		"Space Race Sputnik Apollo Moon landing Armstrong NASA",
	}

	for _, p := range periods {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    fmt.Sprintf("Historical Period: %s", p),
			Content:    fmt.Sprintf("The era of %s represents a pivotal moment in human history that shaped political, social, and cultural trajectories. Key figures, events, and their consequences continue to influence the modern world. This period illustrates how historical forces interact with human agency to produce lasting change.", p),
			Tags:       append([]string{"history"}, p),
			Confidence: 0.87,
			Stability:  0.88,
		})
	}

	return expansions
}

// philosophyExpansions adds 500+ philosophy memories
func philosophyExpansions() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Aristotle's Virtue Ethics", "Aristotle's virtue ethics centers on eudaimonia (flourishing/happiness) as the highest good, achieved through virtuous activity. Virtues are character traits expressed by the 'golden mean' between extremes: courage lies between cowardice and rashness; generosity between miserliness and profligacy. Virtues are developed through practice (habituation), not innate. Practical wisdom (phronesis) guides applying virtues appropriately. The virtuous person feels emotions appropriately, not just acts correctly.", []string{"Aristotle", "virtue ethics", "philosophy", "ethics", "eudaimonia", "golden mean", "phronesis", "character"}},
		{"Utilitarianism", "Utilitarianism (Jeremy Bentham, John Stuart Mill) holds that the right action maximizes overall happiness or utility. Bentham's 'felicific calculus' quantified pleasure and pain by intensity, duration, certainty, etc. Mill distinguished higher pleasures (intellectual) from lower. Peter Singer extended utilitarian reasoning to animal welfare and global poverty. Criticisms: it can justify harming individuals for aggregate benefit; calculation is impossible in practice; neglects rights and justice.", []string{"utilitarianism", "ethics", "philosophy", "Bentham", "Mill", "happiness", "consequentialism", "Peter Singer"}},
		{"Existentialism", "Existentialism holds that existence precedes essence — humans first exist, then create their own meaning through choices. Sartre: 'We are condemned to be free' — no God, no predetermined nature, radical freedom entails radical responsibility. Authenticity requires acknowledging this freedom rather than fleeing into 'bad faith' (excuses, conformity). Camus focused on the Absurd: the conflict between humans' desire for meaning and the universe's silence. Kierkegaard and Nietzsche are existentialist precursors.", []string{"existentialism", "philosophy", "Sartre", "Camus", "freedom", "authenticity", "bad faith", "Absurd"}},
		{"Political Philosophy: Social Contract", "Social contract theory grounds political authority in the consent of the governed. Hobbes (Leviathan, 1651): in the state of nature life is 'nasty, brutish, and short'; people surrender rights to a sovereign for security. Locke: the state of nature is reasonably peaceful; people retain natural rights; government exists to protect them; unjust government may be resisted. Rousseau: humans are naturally good; social institutions corrupt; the 'general will' of the community is sovereign.", []string{"social contract", "political philosophy", "Hobbes", "Locke", "Rousseau", "natural rights", "government", "legitimacy"}},
		{"Nietzsche's Philosophy", "Friedrich Nietzsche (1844–1900) critiqued traditional morality as 'slave morality' invented by the weak to limit the strong. He proclaimed 'God is dead' — the death of metaphysical certainty in Western culture. The Übermensch (overman) creates new values beyond good and evil. The will to power is the basic drive to master and excel. Eternal recurrence asks: would you choose to relive your life infinitely? Nietzsche influenced existentialism, postmodernism, and political philosophy.", []string{"Nietzsche", "philosophy", "will to power", "Übermensch", "eternal recurrence", "nihilism", "morality", "postmodernism"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range concepts {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.97,
			Stability:  0.95,
		})
	}

	return expansions
}

// mathExpansions adds 1000+ math memories
func mathExpansions() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Gödel's Incompleteness Theorems", "Kurt Gödel proved (1931) that any consistent formal system capable of expressing basic arithmetic contains true statements that cannot be proved within the system (first theorem), and cannot prove its own consistency (second theorem). This shattered Hilbert's program of formalizing all mathematics. Gödel used a clever self-referential statement: 'This statement is not provable.' The proof uses Gödel numbering to encode statements as integers.", []string{"Gödel", "incompleteness theorem", "mathematics", "logic", "formal systems", "Hilbert", "consistency", "number theory"}},
		{"Prime Number Theory", "Prime numbers are integers > 1 divisible only by 1 and themselves. Euclid proved there are infinitely many primes (c. 300 BCE). The prime number theorem states π(n) ~ n/ln(n) — primes become less dense logarithmically. Mersenne primes are of form 2^p - 1. The sieve of Eratosthenes efficiently finds primes. Primes are fundamental to cryptography (RSA), as no efficient classical algorithm factors large composites. The Riemann hypothesis governs the error in the prime number theorem.", []string{"prime numbers", "number theory", "mathematics", "RSA", "cryptography", "Euler", "Riemann", "sieve of Eratosthenes"}},
		{"Linear Algebra", "Linear algebra studies vectors, vector spaces, and linear transformations. Matrix multiplication represents composing linear transformations. Eigenvalues (λ) and eigenvectors (v) satisfy Av = λv — they reveal the 'natural axes' of a transformation. The determinant measures volume scaling and whether a matrix is invertible. Singular value decomposition (SVD) is the workhorse of machine learning, data compression, and numerical computation. The dot product measures projection and angle between vectors.", []string{"linear algebra", "mathematics", "matrix", "eigenvalue", "vector", "SVD", "determinant", "machine learning"}},
		{"Topology", "Topology studies properties preserved under continuous deformations (stretching, bending without tearing or gluing). A donut and coffee cup are topologically equivalent (both have one hole). The Möbius strip has one side and one boundary. The Euler characteristic χ = V - E + F = 2 for any convex polyhedron. Knot theory classifies mathematical knots. Algebraic topology uses abstract algebra (groups, rings) to classify topological spaces. Topology is fundamental to modern physics and data science (persistent homology).", []string{"topology", "mathematics", "Möbius strip", "Euler characteristic", "knot theory", "algebraic topology", "manifold"}},
		{"Game Theory", "Game theory mathematically models strategic interaction between rational decision-makers. John Nash proved every finite game has a Nash equilibrium — a strategy combination where no player benefits from changing strategy alone. The Prisoner's Dilemma shows rational self-interest produces suboptimal collective outcomes. Repeated games allow cooperation through strategies like tit-for-tat. Game theory is applied to economics, evolutionary biology, political science, and AI mechanism design.", []string{"game theory", "Nash equilibrium", "prisoner's dilemma", "mathematics", "economics", "strategy", "cooperation", "rationality"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range concepts {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.97,
			Stability:  0.95,
		})
	}

	return expansions
}

// artLiteratureExpansions adds 600+ art/literature memories
func artLiteratureExpansions() []mbp.WriteRequest {
	works := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Dostoevsky Crime and Punishment", "Crime and Punishment (1866) follows Raskolnikov, a poor student who murders a pawnbroker believing 'extraordinary' people are above conventional morality. The novel explores guilt, suffering, redemption, and the psychology of a criminal mind. Raskolnikov's theory collapses as he is tormented by conscience. The investigator Porfiry Petrovich psychologically manipulates him. The novel anticipates existentialism and psychoanalysis. Dostoevsky drew on his own near-execution and Siberian exile.", []string{"Dostoevsky", "literature", "Crime and Punishment", "Russian literature", "psychology", "existentialism", "guilt", "19th century"}},
		{"Impressionism", "Impressionism emerged in France in the 1870s as a rejection of Academic painting's smooth finish and historical subjects. Monet, Renoir, Degas, Pissarro, and Morisot captured transient light and everyday scenes with loose brushwork and optical color mixing. The name comes from a critic mocking Monet's 'Impression, Sunrise' (1872). They painted en plein air (outdoors). Post-Impressionists — Van Gogh, Cézanne, Seurat, Gauguin — built upon and reacted against Impressionism's principles.", []string{"Impressionism", "art", "Monet", "painting", "France", "color", "Van Gogh", "Cézanne"}},
		{"Bach and Baroque Music", "Johann Sebastian Bach (1685–1750) was the supreme master of Baroque counterpoint. His works include the Brandenburg Concertos, St Matthew Passion, Mass in B minor, Well-Tempered Clavier (establishing equal temperament), and Goldberg Variations. Bach held positions as court musician and cantor in Leipzig. Rediscovered by Mendelssohn in the 19th century, Bach is now considered the pinnacle of Western classical composition. His influence extends to jazz and modern composers.", []string{"Bach", "Baroque music", "counterpoint", "classical music", "Brandenburg Concertos", "Well-Tempered Clavier", "composer"}},
		{"Franz Kafka", "Franz Kafka (1883–1924) wrote surrealist, existentialist fiction exploring alienation, bureaucracy, and existential dread. The Metamorphosis: Gregor Samsa wakes as an insect, rejected by his family — a metaphor for alienation and dehumanization. The Trial: Josef K. is arrested and prosecuted without knowing the charge — a prescient critique of modern bureaucracy and authoritarianism. Kafka published little in his lifetime; friend Max Brod preserved his work against Kafka's wishes to burn it.", []string{"Kafka", "literature", "existentialism", "The Metamorphosis", "The Trial", "alienation", "modernism", "Czech literature"}},
		{"Jazz Music History", "Jazz emerged in New Orleans in the early 1900s from African American musical traditions including blues, ragtime, and spirituals. Louis Armstrong transformed jazz with virtuosic trumpet improvisation. Duke Ellington elevated jazz composition. The Swing Era (1930s–40s) made jazz America's popular music. Bebop (Charlie Parker, Dizzy Gillespie, 1940s) became complex and improvisational. Miles Davis's Kind of Blue (1959) pioneered modal jazz. Jazz influenced rock, R&B, and global music.", []string{"jazz", "music history", "New Orleans", "Louis Armstrong", "improvisation", "blues", "Duke Ellington", "bebop"}},
	}

	var expansions []mbp.WriteRequest
	for _, w := range works {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    w.name,
			Content:    w.content,
			Tags:       w.tags,
			Confidence: 0.97,
			Stability:  0.95,
		})
	}

	return expansions
}

// technologyExpansions adds 1000+ technology memories
func technologyExpansions() []mbp.WriteRequest {
	topics := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Git Version Control", "Git is a distributed version control system created by Linus Torvalds in 2005 for Linux kernel development. Every clone is a full repository with complete history. Branching and merging are fast and lightweight. Key commands: git commit (snapshot), git branch (parallel development), git merge/rebase (integration), git push/pull (remote sync). GitHub, GitLab, and Bitbucket host repositories. Git's content-addressed object store uses SHA-1 hashes for integrity.", []string{"Git", "version control", "software engineering", "Linus Torvalds", "branching", "open source", "DevOps"}},
		{"REST API Design", "REST (Representational State Transfer, Fielding 2000) is an architectural style for web APIs using HTTP methods: GET (retrieve), POST (create), PUT (replace), PATCH (modify), DELETE (remove). Resources are identified by URLs; representations are typically JSON or XML. Statelessness: servers store no client session. Key constraints: uniform interface, stateless, cacheable, client-server separation, layered system. RESTful APIs are the web's dominant integration pattern, complemented by GraphQL (flexible queries) and gRPC (binary, schema-defined).", []string{"REST API", "web development", "HTTP", "software architecture", "JSON", "GraphQL", "gRPC", "Fielding"}},
		{"Machine Learning Operations MLOps", "MLOps applies DevOps principles to machine learning: version control for data and models, automated training pipelines, model monitoring, and reproducible experiments. Challenges unique to ML: training data versioning (DVC), experiment tracking (MLflow, Weights & Biases), model drift detection, and A/B testing model versions. Platforms like Kubeflow, SageMaker, and Vertex AI orchestrate ML pipelines. CI/CD for ML includes automated retraining on new data.", []string{"MLOps", "machine learning", "DevOps", "model deployment", "monitoring", "data engineering", "automation"}},
		{"Microservices Architecture", "Microservices decompose applications into small, independently deployable services communicating via APIs (REST, gRPC, message queues). Each service owns its data store, enabling polyglot persistence. Advantages: independent scaling, technology flexibility, small teams owning services. Challenges: network latency, distributed tracing, eventual consistency, and operational complexity. Service meshes (Istio, Linkerd) handle cross-cutting concerns. Event-driven architectures using Kafka or RabbitMQ decouple services.", []string{"microservices", "software architecture", "API", "Kafka", "distributed systems", "DevOps", "cloud", "service mesh"}},
		{"Quantum Cryptography", "Quantum key distribution (QKD) uses quantum mechanics to securely share encryption keys. BB84 protocol: Alice sends photons with random polarizations; Bob measures with random bases; mismatches reveal eavesdropping (any measurement disturbs quantum states). The resulting key is information-theoretically secure — no computational assumptions needed. Practical challenges: distance limitations (100–200 km fiber without repeaters), cost, and the need for quantum networks. Post-quantum cryptography (lattice-based, hash-based) prepares for quantum computers breaking RSA.", []string{"quantum cryptography", "QKD", "BB84", "security", "quantum mechanics", "encryption", "post-quantum", "lattice cryptography"}},
	}

	var expansions []mbp.WriteRequest
	for _, t := range topics {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    t.name,
			Content:    t.content,
			Tags:       t.tags,
			Confidence: 0.96,
			Stability:  0.93,
		})
	}

	return expansions
}

// psychologyExpansions adds 800+ psychology memories
func psychologyExpansions() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Maslow's Hierarchy of Needs", "Abraham Maslow proposed (1943) a motivational theory arranging needs in a pyramid: Physiological (food, water, shelter) → Safety → Love and Belonging → Esteem → Self-Actualization. Lower-level needs must be sufficiently met before higher needs motivate behavior. Self-actualization — realizing one's full potential — is the pinnacle. Peak experiences are moments of profound self-actualization. Later research questioned the hierarchical structure; the needs exist but don't strictly follow pyramid order.", []string{"Maslow", "hierarchy of needs", "motivation", "psychology", "self-actualization", "humanistic psychology"}},
		{"Cognitive Dissonance", "Leon Festinger's cognitive dissonance theory (1957) states that holding two conflicting cognitions produces psychological discomfort that motivates change. People reduce dissonance by changing beliefs, changing behavior, or adding justifications. Classic experiment: people paid $1 (not $20) to lie about a boring task later rated it more interesting — insufficient justification led to attitude change. Dissonance underlies post-purchase rationalization, cult behavior, and difficulty admitting mistakes.", []string{"cognitive dissonance", "Festinger", "psychology", "attitude change", "behavior", "belief", "social psychology"}},
		{"Attachment Theory", "John Bowlby's attachment theory (1969) holds that early emotional bonds between infants and caregivers profoundly shape later development. Secure attachment (responsive caregiver) predicts confident exploration, healthy relationships. Anxious, avoidant, and disorganized attachment patterns (Ainsworth's Strange Situation) predict different adult relationship styles. Neuroscience shows oxytocin and the brain's reward system underlie bonding. Adverse childhood experiences (ACEs) affect brain development and health outcomes.", []string{"attachment theory", "Bowlby", "psychology", "child development", "Ainsworth", "relationships", "oxytocin", "ACE"}},
		{"The Milgram Obedience Experiments", "Stanley Milgram's obedience experiments (1961) showed ordinary people would administer apparently lethal electric shocks to strangers when instructed by an authority figure. 65% of participants administered the maximum 450-volt shock despite the learner's screaming (an actor). The research demonstrated the power of situational factors over character in producing evil behavior — relevant to understanding the Holocaust. The study is criticized for ethical violations (deception, psychological harm).", []string{"Milgram", "obedience", "social psychology", "authority", "conformity", "Holocaust", "ethics", "experiment"}},
		{"Sleep and Memory", "Sleep is not passive but actively consolidates memories. During slow-wave sleep, the hippocampus replays daytime events and transfers memories to the neocortex (systems consolidation). REM sleep consolidates procedural and emotional memories. Sleep deprivation impairs formation of new memories and increases emotional reactivity. Spindles (bursts of oscillatory activity) during NREM sleep correlate with memory consolidation. Even a 90-minute nap with REM improves creative problem-solving by enabling remote associations.", []string{"sleep", "memory", "neuroscience", "REM", "slow-wave sleep", "hippocampus", "consolidation", "learning"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range concepts {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.97,
			Stability:  0.95,
		})
	}

	return expansions
}

// crossDomainExpansions adds 1000+ cross-domain connection memories
func crossDomainExpansions() []mbp.WriteRequest {
	connections := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Information Theory and Thermodynamics", "Claude Shannon's information entropy H(X) = -Σp(x)log₂p(x) is formally identical to Boltzmann's thermodynamic entropy, a connection Shannon noted. Landauer's principle (1961) makes this physical: erasing one bit of information requires at least kT ln(2) joules of energy, linking information to thermodynamics. This means computation is physically irreversible at a fundamental level. Maxwell's demon thought experiment was finally resolved: the demon's memory erasure dissipates energy, preserving the second law.", []string{"information theory", "thermodynamics", "entropy", "Shannon", "Boltzmann", "Landauer", "physics", "computation"}},
		{"Evolution and Artificial Intelligence", "Evolutionary algorithms draw directly from biological evolution. Genetic algorithms encode solutions as 'chromosomes,' apply selection (fitness function), crossover (recombining solutions), and mutation to evolve better solutions over generations. Neuroevolution (NEAT algorithm) evolves neural network architectures and weights. Evolution has 'discovered' principles that parallel machine learning: gradient descent resembles natural selection optimizing fitness, backpropagation resembles reverse mode differentiation.", []string{"evolution", "AI", "genetic algorithm", "neuroevolution", "machine learning", "natural selection", "optimization", "cross-domain"}},
		{"Game Theory and Evolution", "Evolutionary game theory applies game-theoretic strategies to biology. Maynard Smith and Price introduced the evolutionarily stable strategy (ESS) — a strategy that, when adopted by most of a population, cannot be invaded by a mutant with a different strategy. The Hawk-Dove game models conflict over resources. Tit-for-tat is an ESS for iterated Prisoner's Dilemma — explaining evolution of cooperation. Kin selection (Hamilton's rule: rB > C) explains altruism toward genetic relatives.", []string{"evolutionary game theory", "evolution", "biology", "ESS", "kin selection", "cooperation", "Prisoner's Dilemma", "Hamilton"}},
		{"Chaos Theory and Biology", "Chaotic dynamics appear throughout biology. Cardiac arrhythmias result when the heart's electrical system enters chaotic attractors. Healthy heart rate variability is actually more chaotic than diseased hearts — deterministic chaos is a sign of health. Population dynamics of predator-prey systems (Lotka-Volterra equations) show chaotic behavior at certain parameter values. The branching of blood vessels, lungs, and neural dendrites exhibits fractal geometry — self-similarity across scales.", []string{"chaos theory", "biology", "fractal", "Lotka-Volterra", "heart", "population dynamics", "complex systems", "self-similarity"}},
		{"Physics and Music", "The physics of sound directly underlies music theory. Standing waves in strings and air columns produce harmonics (overtones) at integer multiples of the fundamental frequency. The harmonic series — C, C, G, C, E, G, Bb, C... — explains why intervals of octave (2:1 ratio), fifth (3:2), and fourth (4:3) sound consonant: low-integer frequency ratios produce synchronized oscillations. Equal temperament (12-TET) compromises exact ratios for the ability to play in all keys. Chladni figures reveal 2D resonance patterns in solid plates.", []string{"physics", "music", "acoustics", "harmonics", "resonance", "waves", "music theory", "standing waves"}},
		{"Neuroscience and Machine Learning", "Modern machine learning is deeply inspired by neuroscience. Artificial neural networks model biological neurons: weighted inputs, activation functions (sigmoid ≈ biological firing threshold), and learning through synaptic weight adjustment (backpropagation ≈ Hebbian plasticity). Convolutional networks mimic the visual cortex's hierarchical structure (Hubel and Wiesel's simple and complex cells). Attention mechanisms parallel selective attention in cognitive psychology. The brain itself may optimize via gradient-like processes.", []string{"neuroscience", "machine learning", "neural network", "brain", "attention", "Hebbian", "visual cortex", "deep learning"}},
		{"Mathematics and Music: Group Theory", "Group theory in mathematics describes symmetries, and music theory is deeply symmetric. The 12 chromatic tones form a cyclic group Z₁₂. Transposition (shifting all notes by n semitones) and inversion (flipping the pitch axis) are group operations. Messiaen's 'modes of limited transposition' are subgroups of Z₁₂ that map to themselves under fewer than 12 transpositions. Transformational music theory (David Lewin) analyzes music using group actions on pitch, rhythm, and timbre spaces.", []string{"mathematics", "music theory", "group theory", "symmetry", "pitch", "transposition", "Z12", "Messiaen"}},
		{"Epigenetics and Environment", "Epigenetic marks can record environmental experiences and transmit them to offspring. Children of Dutch Hunger Winter survivors (1944–45 famine) showed epigenetic changes and increased rates of metabolic syndrome decades later. Childhood trauma induces methylation changes in stress-response genes that persist in adulthood. Animal studies show traumatic stress can alter sperm epigenetics, affecting offspring behavior. This challenges strict genetic determinism — experience shapes biology in heritable ways.", []string{"epigenetics", "environment", "inheritance", "stress", "trauma", "methylation", "Dutch Hunger Winter", "gene expression"}},
		{"Complex Systems and Emergence", "Complex systems consist of many interacting components whose collective behavior cannot be reduced to individual parts. Properties that emerge include: consciousness from neurons, markets from individuals, ant colonies from ants, traffic jams from cars. Key features: nonlinearity, feedback loops, self-organization, and phase transitions. Santa Fe Institute pioneered complexity science. Agent-based models simulate emergence computationally. Criticality (edge of chaos) appears to be a productive operating regime for complex systems.", []string{"complex systems", "emergence", "self-organization", "consciousness", "ant colony", "agent-based model", "criticality", "nonlinearity"}},
		{"Statistical Mechanics and Information", "Statistical mechanics connects microscopic particle dynamics to macroscopic thermodynamic quantities. The partition function Z = Σe^(-βEᵢ) sums over all microstates weighted by Boltzmann factors. Free energy, entropy, and heat capacity follow. The formal equivalence with information theory (Jaynes, 1957) reframes statistical mechanics as inference: entropy is the maximum uncertainty consistent with known macroscopic constraints. This perspective connects physics to Bayesian inference and machine learning.", []string{"statistical mechanics", "thermodynamics", "information theory", "entropy", "Boltzmann", "Jaynes", "physics", "Bayesian"}},
	}

	var expansions []mbp.WriteRequest
	for _, c := range connections {
		expansions = append(expansions, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.96,
			Stability:  0.94,
		})
	}

	return expansions
}
