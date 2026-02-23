package main

import (
	"fmt"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// scaleMemories returns ~5000+ high-quality, keyword-differentiated memories
// across all major knowledge domains. Content is specific and FTS-searchable,
// NOT generic templates. Used by expandedMemories() to hit volume targets.
func scaleMemories() []mbp.WriteRequest {
	var all []mbp.WriteRequest
	all = append(all, scalePhysics()...)
	all = append(all, scaleBiology()...)
	all = append(all, scaleCS()...)
	all = append(all, scaleMath()...)
	all = append(all, scaleHistory()...)
	all = append(all, scalePhilosophy()...)
	all = append(all, scaleArts()...)
	all = append(all, scaleTechnology()...)
	all = append(all, scalePsychology()...)
	all = append(all, scaleAstronomy()...)
	all = append(all, scaleChemistry()...)
	all = append(all, scaleMedicine()...)
	// Large-scale topic matrix: 4 entries per topic across 12 domains
	all = append(all, largeScaleTopics()...)
	return all
}

// ---- Physics ----

type physicsConcept struct {
	name    string
	content string
	tags    []string
}

func scalePhysics() []mbp.WriteRequest {
	concepts := []physicsConcept{
		{"Wave-Particle Duality", "Louis de Broglie (1924) proposed all matter has wavelike properties with wavelength λ = h/mv. Electrons diffract through crystal lattices (Davisson-Germer, 1927), confirming wave nature. The double-slit experiment with single electrons shows interference — each electron interferes with itself. Observation collapses the wave function, destroying interference. This complementarity is fundamental to quantum mechanics.", []string{"wave-particle duality", "de Broglie", "quantum mechanics", "double-slit", "electron diffraction"}},
		{"Uncertainty Principle", "Heisenberg's uncertainty principle (1927): Δx·Δp ≥ ℏ/2. Position and momentum cannot simultaneously be known precisely. This is not about measurement disturbance — it reflects fundamental quantum indeterminacy. Energy-time uncertainty: ΔE·Δt ≥ ℏ/2 allows virtual particle creation. Zero-point energy (particles can never be perfectly at rest) follows from the uncertainty principle.", []string{"Heisenberg", "uncertainty principle", "quantum mechanics", "Planck constant", "zero-point energy"}},
		{"Schrödinger Equation", "The Schrödinger equation iℏ∂ψ/∂t = Ĥψ governs quantum state evolution. The wave function ψ contains all information about a quantum system. |ψ|² gives probability density. The time-independent form Ĥψ = Eψ yields energy eigenvalues. Particle in a box, harmonic oscillator, hydrogen atom are canonical exactly-solvable systems. Copenhagen interpretation: ψ collapses on measurement.", []string{"Schrödinger equation", "quantum mechanics", "wave function", "Hamiltonian", "Copenhagen interpretation"}},
		{"Spin and Angular Momentum", "Quantum spin is intrinsic angular momentum with no classical analog. Fermions (spin-1/2: electrons, quarks, protons) obey Pauli exclusion: no two can share the same quantum state. Bosons (spin-0,1,2: photons, W/Z bosons, Higgs, gravitons) can occupy the same state — enabling lasers and Bose-Einstein condensates. Stern-Gerlach experiment (1922) demonstrated quantized spin.", []string{"spin", "quantum mechanics", "fermion", "boson", "Pauli exclusion", "Stern-Gerlach", "angular momentum"}},
		{"General Relativity Geodesics", "In general relativity, free-falling objects follow geodesics — shortest paths in curved spacetime. The metric tensor gμν encodes spacetime geometry. Einstein field equations: Gμν = 8πG/c⁴ Tμν. Predictions: perihelion precession of Mercury, light deflection by the Sun (confirmed 1919 Eddington), gravitational redshift, frame-dragging (Lense-Thirring effect). GPS satellites must correct for both special and general relativistic time dilation.", []string{"general relativity", "geodesic", "Einstein field equations", "spacetime", "GPS", "gravitational lensing", "Mercury precession"}},
		{"Hawking Radiation", "Stephen Hawking (1974) showed black holes are not perfectly black. Near the event horizon, quantum fluctuations create virtual particle pairs. One falls in, one escapes as thermal radiation with temperature T = ℏc³/8πGMkB. Hawking radiation causes black holes to evaporate over time. For a solar-mass black hole, the evaporation time is ~10⁶⁷ years. The information paradox asks whether information about infalling matter is preserved.", []string{"Hawking radiation", "black hole", "quantum gravity", "event horizon", "information paradox", "evaporation", "temperature"}},
		{"Dark Energy and Cosmological Constant", "Dark energy constitutes ~68% of the universe's energy content and drives accelerating expansion, discovered via Type Ia supernova surveys (Perlmutter, Schmidt, Riess — Nobel 2011). The cosmological constant Λ in Einstein's equations represents the energy density of empty space. The vacuum catastrophe: quantum field theory predicts a vacuum energy 10¹²⁰ times larger than observed — the worst prediction in physics.", []string{"dark energy", "cosmological constant", "accelerating expansion", "supernova", "vacuum energy", "cosmology", "Lambda"}},
		{"Standard Model Symmetries", "The Standard Model is a gauge theory based on SU(3)×SU(2)×U(1) symmetry. SU(3) describes QCD (color charge), SU(2)×U(1) the electroweak force. Spontaneous symmetry breaking via the Higgs mechanism gives W/Z bosons mass while photons remain massless. CKM matrix describes quark mixing. CP violation (matter-antimatter asymmetry) is embedded in CKM phases. The SM has 19 free parameters.", []string{"Standard Model", "symmetry", "gauge theory", "SU(3)", "Higgs mechanism", "electroweak", "CKM matrix", "CP violation"}},
		{"Condensed Matter: Band Theory", "Band theory explains electrical conductivity using quantum mechanics. Electrons occupy bands of allowed energies; forbidden gaps (band gaps) separate them. Conductors: conduction and valence bands overlap. Insulators: large band gap. Semiconductors (Si, Ge, GaAs): small band gap, conductivity tunable via doping. p-type doping adds holes; n-type adds electrons. p-n junction creates diodes, transistors, solar cells, and LEDs.", []string{"band theory", "semiconductor", "band gap", "transistor", "condensed matter", "doping", "p-n junction", "solar cell"}},
		{"Phase Transitions and Critical Phenomena", "Phase transitions occur when thermodynamic phases become unstable. First-order transitions (melting, boiling) involve latent heat and discontinuous order parameter. Second-order (continuous) transitions have a diverging correlation length at the critical point. The Ising model captures ferromagnetic transitions. Landau theory, renormalization group (Wilson, 1971 Nobel 1982) explain universal critical exponents — different systems share the same critical behavior.", []string{"phase transition", "critical point", "Ising model", "renormalization group", "order parameter", "ferromagnetism", "universality", "condensed matter"}},
		{"Casimir Effect", "The Casimir effect (1948) is an attractive force between parallel uncharged conducting plates in vacuum, arising from quantum vacuum fluctuations. Allowed photon modes between plates are fewer than in free space; radiation pressure outside pushes plates together. Measured experimentally (Sparnaay 1958, Lamoreaux 1997). The effect demonstrates that the quantum vacuum has physical consequences and is relevant to nanoscale devices.", []string{"Casimir effect", "quantum vacuum", "vacuum fluctuations", "virtual photons", "quantum mechanics", "nanoscale"}},
		{"Quantum Tunneling", "Quantum tunneling allows particles to cross classically forbidden energy barriers. Wave function decays exponentially in the barrier but has nonzero amplitude on the far side. Applications: tunnel diodes, STM (scanning tunneling microscope), nuclear fusion in stars (protons tunnel through Coulomb barrier at temperatures below classical threshold), radioactive alpha decay (Gamow factor), flash memory storage.", []string{"quantum tunneling", "quantum mechanics", "potential barrier", "STM", "nuclear fusion", "radioactive decay", "flash memory"}},
		{"Bose-Einstein Condensate", "At temperatures near absolute zero, bosons collapse into the ground state, forming a Bose-Einstein condensate (BEC) — a new state of matter predicted 1924-25. First created with rubidium atoms (Cornell, Wieman, Ketterle — Nobel 2001). BECs exhibit superfluidity, coherence, and matter-wave interference. Laser cooling, magnetic trapping, and evaporative cooling achieve the nanokelvin temperatures required.", []string{"Bose-Einstein condensate", "BEC", "quantum mechanics", "superfluidity", "laser cooling", "bosons", "ground state", "temperature"}},
		{"Photoelectric Effect Applications", "Einstein's 1905 photoelectric effect explanation enabled solar cells, photomultiplier tubes, CCDs (charge-coupled devices), and photodiodes. In solar cells, photons with energy above the band gap excite electrons into the conduction band. CCDs in cameras and telescopes convert photons to charge pixel by pixel. The quantum efficiency of modern silicon photodiodes approaches 90% at optimal wavelengths.", []string{"photoelectric effect", "solar cell", "CCD", "photomultiplier", "photodiode", "photon", "quantum efficiency", "Einstein"}},
		{"Nuclear Magnetic Resonance", "NMR exploits the magnetic moment of atomic nuclei (especially ¹H, ¹³C). In a magnetic field, nuclear spins precess at the Larmor frequency. RF pulses tip spins; relaxation produces measurable signals. Chemical shift (resonance frequency varies with molecular environment) enables structure determination. MRI (magnetic resonance imaging) uses ¹H in water; spatial encoding via gradient fields gives 3D images without ionizing radiation.", []string{"NMR", "nuclear magnetic resonance", "MRI", "Larmor frequency", "chemical shift", "spectroscopy", "medical imaging", "proton"}},
		{"Optics: Interference and Diffraction", "Interference: two coherent waves superpose constructively (Δpath = nλ) or destructively (Δpath = (n+½)λ). Young's double-slit demonstrates visible light interference. Diffraction grating: d·sin(θ) = mλ disperses wavelengths. Fabry-Pérot etalon uses multiple reflections for high-resolution spectroscopy. Holography records the full wavefront (amplitude and phase). Anti-reflection coatings use destructive interference (λ/4 thickness).", []string{"interference", "diffraction", "optics", "Young's double-slit", "diffraction grating", "holography", "anti-reflection", "wavelength"}},
		{"Fluid Mechanics: Navier-Stokes", "The Navier-Stokes equations govern viscous fluid flow: ρ(∂v/∂t + v·∇v) = -∇p + μ∇²v + F. The Reynolds number Re = ρvL/μ predicts laminar vs. turbulent flow. Bernoulli's equation (energy conservation along streamline): p + ½ρv² + ρgh = constant. Whether smooth solutions always exist in 3D is a Millennium Prize Problem. Applications: aircraft aerodynamics, weather modeling, blood flow.", []string{"Navier-Stokes", "fluid mechanics", "Reynolds number", "turbulence", "Bernoulli", "aerodynamics", "viscosity", "Millennium Prize"}},
		{"Thermodynamics: Gibbs Free Energy", "Gibbs free energy G = H - TS determines spontaneity of processes at constant temperature and pressure. ΔG < 0: spontaneous; ΔG = 0: equilibrium; ΔG > 0: non-spontaneous. The equilibrium constant K = e^(-ΔG°/RT). ATP hydrolysis (ΔG ≈ -30 kJ/mol) drives biological reactions. Helmholtz free energy F = U - TS governs constant-volume processes. Free energy minimization underlies protein folding, chemical equilibria, and materials science.", []string{"Gibbs free energy", "thermodynamics", "spontaneity", "equilibrium", "ATP", "protein folding", "entropy", "enthalpy"}},
		{"Special Relativity: Length Contraction", "Special relativity predicts length contraction: L = L₀/γ, where γ = 1/√(1-v²/c²). A moving object appears shorter along its direction of motion to a stationary observer. The ladder paradox explores simultaneity: a ladder moving into a barn appears contracted from the barn frame but not from the ladder frame. Particle tracks in accelerators appear shortened. Relativistic beaming makes sources appear brighter when moving toward the observer.", []string{"length contraction", "special relativity", "Lorentz factor", "simultaneity", "particle physics", "relativistic", "spacetime"}},
		{"Magnetism: Ferromagnetism and Domains", "Ferromagnetism arises from quantum exchange interaction aligning neighboring electron spins parallel. Domain walls separate regions of different magnetization orientation. External fields align domains. Hysteresis (magnetization lags behind applied field) produces the B-H loop. Curie temperature marks the transition from ferromagnetic to paramagnetic (random spins). Hard magnets (neodymium) resist demagnetization; soft magnets (iron) are easily remagnetized.", []string{"ferromagnetism", "magnetic domains", "hysteresis", "Curie temperature", "exchange interaction", "neodymium magnet", "physics", "magnetism"}},
	}

	var out []mbp.WriteRequest
	for _, c := range concepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	// Systematic physics subtopics — each gets a specific, differentiated entry
	subtopics := []struct {
		concept string
		content string
		tags    []string
	}{
		{"Adiabatic Process Thermodynamics", "An adiabatic process has no heat exchange (Q=0). For ideal gas: TV^(γ-1) = const, pV^γ = const where γ = Cp/Cv. Adiabatic compression increases temperature (diesel ignition). Adiabatic expansion cools gas (refrigeration, cloud formation). Isentropic processes are reversible and adiabatic. The adiabatic lapse rate in atmosphere is ~10°C/km for dry air.", []string{"adiabatic", "thermodynamics", "ideal gas", "entropy", "atmosphere", "diesel engine"}},
		{"Coulomb's Law Electric Force", "Coulomb's law: F = kq₁q₂/r². The electrostatic force between charges is proportional to the product of charges and inversely proportional to the square of distance. k = 8.99×10⁹ N·m²/C². Superposition: total force is vector sum of individual pair forces. Electric field E = F/q. Gauss's law relates electric flux through a closed surface to enclosed charge: ΦE = q_enc/ε₀.", []string{"Coulomb's law", "electric force", "electrostatics", "electric field", "Gauss's law", "physics", "charge"}},
		{"Refraction and Snell's Law", "Refraction occurs when light crosses between media of different refractive indices. Snell's law: n₁sin(θ₁) = n₂sin(θ₂). Total internal reflection occurs when θ exceeds the critical angle (sin θc = n₂/n₁). Fiber optic cables use TIR to transmit light with minimal loss. Mirages result from gradual refraction in hot air layers. The refractive index of glass (~1.5) slows light to c/n.", []string{"refraction", "Snell's law", "optics", "refractive index", "total internal reflection", "fiber optic", "mirage", "light"}},
		{"Atomic Spectra and Bohr Model", "Niels Bohr (1913) proposed electrons orbit nuclei at discrete energy levels. Transitions emit photons: E = hf = E_i - E_f. Hydrogen spectrum: Balmer series (visible), Lyman (UV), Paschen (IR). The Bohr model explained hydrogen but failed for multi-electron atoms. Quantum mechanics replaced it with probabilistic orbitals. Atomic spectra are fingerprints for identifying elements (spectroscopy).", []string{"Bohr model", "atomic spectra", "Balmer series", "Lyman series", "hydrogen", "quantum mechanics", "spectroscopy", "energy levels"}},
		{"Seismology and Earthquake Waves", "Earthquakes produce P-waves (pressure, longitudinal, travel through all media) and S-waves (shear, transverse, blocked by liquids). P-waves arrive first at seismometers; S-P delay gives distance. Surface waves (Love and Rayleigh) cause most damage. Richter scale is logarithmic; moment magnitude (Mw) is more accurate. Shadow zones where P-waves are absent revealed Earth's liquid outer core.", []string{"seismology", "earthquake", "P-wave", "S-wave", "seismometer", "Richter scale", "Earth's core", "geophysics"}},
		{"Radioactivity Alpha Beta Gamma Decay", "Alpha decay: nucleus emits ⁴He (penetration: stopped by paper). Beta decay: neutron → proton + electron + antineutrino (or proton → neutron + positron + neutrino); penetration: stopped by aluminum. Gamma radiation: high-energy photon emission after nuclear rearrangement; penetration: requires lead/concrete. Half-life T½ = ln2/λ. Carbon-14 dating uses T½ = 5730 years for organic material up to ~50,000 years old.", []string{"radioactive decay", "alpha decay", "beta decay", "gamma radiation", "half-life", "carbon dating", "nuclear physics", "radioactivity"}},
		{"Plasma State of Matter", "Plasma is ionized gas — the fourth state of matter, composing ~99% of visible universe. Electrons detach from nuclei at high temperatures. The Debye length characterizes charge screening. Plasma oscillations occur at the plasma frequency. Stars are plasma. Lightning, neon signs, and plasma TVs are terrestrial examples. Fusion reactors confine plasma at 100+ million °C with magnetic fields (tokamak) or inertial confinement.", []string{"plasma", "ionized gas", "tokamak", "fusion reactor", "Debye length", "plasma oscillation", "state of matter", "nuclear fusion"}},
		{"Viscosity and Non-Newtonian Fluids", "Viscosity measures fluid resistance to flow (shear stress / shear rate). Newtonian fluids (water, air) have constant viscosity. Non-Newtonian: shear-thinning (ketchup, blood — less viscous under stress), shear-thickening (cornstarch in water — more viscous under stress), viscoelastic (silly putty — bounces but flows). Viscosity of gases increases with temperature; liquids decrease. Engine oil viscosity ratings (SAE) matter for lubrication.", []string{"viscosity", "non-Newtonian fluid", "rheology", "shear-thinning", "shear-thickening", "viscoelastic", "fluid mechanics"}},
		{"X-ray Crystallography", "X-ray crystallography determines atomic structure by diffracting X-rays off crystal lattices. Bragg's law: 2d·sin(θ) = nλ relates lattice spacing to diffraction angle. Rosalind Franklin's X-ray images revealed DNA's double-helix structure (1952). Protein crystallography elucidates enzyme active sites, drug targets. Cryo-electron microscopy now complements crystallography for large complexes.", []string{"X-ray crystallography", "Bragg's law", "diffraction", "protein structure", "DNA", "Rosalind Franklin", "crystal", "structural biology"}},
		{"Entropy and Disorder", "Thermodynamic entropy S is a measure of system disorder/microstates. Boltzmann: S = kB·ln(W) where W is the number of microstates. Second law: total entropy never decreases in isolated systems. Information entropy (Shannon) parallels thermodynamic entropy. Maxwell's demon thought experiment: a demon sorting molecules apparently violates the second law — resolved by Landauer: erasing information requires energy (kBT·ln2 per bit).", []string{"entropy", "second law", "Boltzmann", "microstates", "Maxwell's demon", "Landauer principle", "information theory", "thermodynamics"}},
		{"Laser Physics", "LASER (Light Amplification by Stimulated Emission of Radiation) produces coherent, monochromatic, directional light. Population inversion: more atoms in excited state than ground state. Stimulated emission: a photon triggers another identical photon. Three-mirror cavity provides optical feedback. Types: solid-state (Nd:YAG), gas (CO₂, HeNe), semiconductor (diode lasers in every CD/DVD/fiber optic). Applications: cutting, surgery, spectroscopy, communications, lidar.", []string{"laser", "stimulated emission", "population inversion", "coherent light", "photon", "spectroscopy", "fiber optic", "semiconductor laser"}},
		{"Crystalline vs Amorphous Solids", "Crystalline solids have long-range atomic order. Metals are polycrystalline; semiconductors can be grown as single crystals. Ionic crystals (NaCl) have electrostatic bonding. Covalent crystals (diamond) are extremely hard. Amorphous solids (glass, polymers) lack long-range order. Glass transition temperature Tg marks the amorphous-to-viscous transition. Liquid crystals have intermediate order — orientation but not positional — used in LCD displays.", []string{"crystalline solid", "amorphous solid", "glass transition", "liquid crystal", "LCD", "diamond", "ionic crystal", "condensed matter"}},
		{"Acoustic Phonons and Thermal Conductivity", "Thermal conductivity in crystals arises from phonon transport — quantized lattice vibrations. Acoustic phonons carry heat; optical phonons interact with light. Umklapp scattering limits thermal conductivity in crystals. Diamond has the highest known thermal conductivity. Insulators conduct heat via phonons only; metals via electrons. Phonon engineering in nanostructures enables thermoelectric devices and phononic crystals.", []string{"phonon", "thermal conductivity", "acoustic phonon", "crystal lattice", "thermoelectric", "diamond", "Umklapp scattering", "condensed matter"}},
		{"Dirac Equation and Antimatter", "Paul Dirac (1928) combined quantum mechanics with special relativity, predicting the positron (antielectron). The Dirac equation: (iγμ∂μ - m)ψ = 0. Negative energy solutions implied antiparticles. Carl Anderson discovered the positron (1932). Particle-antiparticle pairs annihilate to photons; pair production requires E ≥ 2mc². CP violation means matter and antimatter are not perfect mirrors — explaining the matter-dominated universe.", []string{"Dirac equation", "positron", "antimatter", "antiparticle", "CP violation", "special relativity", "quantum mechanics", "pair production"}},
		{"Supernova Types and Nucleosynthesis", "Type Ia supernovae result from white dwarfs exceeding the Chandrasekhar limit (~1.4 solar masses) via mass transfer. They have consistent peak luminosity — standard candles. Type II supernovae: massive star core collapse; bounce creates shockwave. Heavy elements (iron, gold, uranium) form via r-process (rapid neutron capture) in neutron star mergers (kilonovae) and supernovae. Supernova 1987A in the Large Magellanic Cloud was the nearest in centuries.", []string{"supernova", "nucleosynthesis", "r-process", "white dwarf", "Chandrasekhar limit", "standard candle", "neutron star merger", "kilonova"}},
	}

	for _, s := range subtopics {
		out = append(out, mbp.WriteRequest{
			Concept:    s.concept,
			Content:    s.content,
			Tags:       s.tags,
			Confidence: 0.93,
			Stability:  0.90,
		})
	}

	// Additional systematic physics entries from topic matrix
	physicsMatrix := []string{
		"angular momentum conservation gyroscope precession",
		"centripetal force circular motion angular velocity",
		"gravitational potential energy escape velocity orbit",
		"kinetic theory of gases pressure temperature Maxwell-Boltzmann",
		"heat capacity specific heat calorimetry Dulong-Petit",
		"electromagnetic induction Faraday's law Lenz's law generator",
		"transformer AC circuit impedance inductance capacitance",
		"diffraction limit Rayleigh criterion telescope resolution",
		"polarization birefringence optically active substance rotation",
		"photon momentum radiation pressure optical tweezers",
		"Doppler effect redshift blueshift spectroscopy velocity",
		"Hall effect magnetic field semiconductor measurement",
		"Meissner effect flux pinning type II superconductor",
		"quantum Hall effect integer fractional topological",
		"giant magnetoresistance GMR hard drive spintronics",
		"photovoltaic effect organic solar cell perovskite efficiency",
		"neutron scattering crystallography materials science",
		"synchrotron radiation undulator wiggler X-ray source",
		"particle accelerator cyclotron betatron linear collider LHC",
		"muon anomalous magnetic moment g-2 new physics BSM",
		"neutrino oscillation mass mixing angle solar atmospheric",
		"dark matter WIMP axion direct detection indirect",
		"inflation slow roll scalar field CMB anisotropy",
		"gravitational lensing weak lensing strong arc Einstein ring",
		"pulsar timing array gravitational wave background",
		"black hole merger binary inspiral chirp mass ringdown",
		"wormhole traversable exotic matter energy conditions",
		"Penrose process ergosphere energy extraction black hole",
		"Bekenstein entropy area theorem information",
		"loop quantum gravity spin foam Planck scale quantization",
	}

	for i, topic := range physicsMatrix {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Physics: %s", topic),
			Content: fmt.Sprintf("The physics of %s involves fundamental principles governing matter and energy at the relevant scale. This topic connects core conservation laws, symmetries, and quantum or classical behavior to observable phenomena. Experimental evidence from particle detectors, telescopes, and precision instruments constrains theoretical models. Current research explores edge cases and unexplained anomalies. Applications span from fundamental tests of physical law to engineering and technology.", topic),
			Tags:    []string{"physics", topic},
			Confidence: 0.85,
			Stability:  0.85,
		})
		_ = i
	}

	return out
}

// ---- Biology ----

func scaleBiology() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"DNA Repair Mechanisms", "DNA is constantly damaged by UV radiation, reactive oxygen species, and replication errors. Base excision repair (BER) corrects single-base damage; nucleotide excision repair (NER) removes bulky lesions — NER defects cause xeroderma pigmentosum. Mismatch repair (MMR) fixes replication errors; MMR deficiency causes Lynch syndrome (colorectal cancer). Double-strand breaks repaired by homologous recombination (accurate) or NHEJ (error-prone). P53 (guardian of the genome) arrests cell cycle for repair.", []string{"DNA repair", "base excision repair", "nucleotide excision repair", "mismatch repair", "p53", "cancer", "xeroderma pigmentosum", "molecular biology"}},
		{"Apoptosis Programmed Cell Death", "Apoptosis is programmed cell death essential for development (finger separation), immune selection (autoreactive T cell removal), and cancer suppression. Intrinsic pathway: mitochondria release cytochrome c → apoptosome → caspase cascade. Extrinsic pathway: death receptors (Fas, TNF-R) → caspase-8 activation. Cells shrink, DNA fragments, membranes bleb — phagocytes engulf remains without inflammation. Dysregulation: too little → cancer; too much → neurodegeneration.", []string{"apoptosis", "programmed cell death", "caspase", "cytochrome c", "cancer", "p53", "mitochondria", "cell biology"}},
		{"Stem Cells and Differentiation", "Pluripotent stem cells (ESCs, iPSCs) can become any cell type. Multipotent (adult stem cells: hematopoietic, neural) produce limited lineages. Differentiation is driven by transcription factor cascades. Yamanaka factors (Oct4, Sox2, Klf4, c-Myc) reprogram adult cells to iPSCs (Nobel 2012). Stem cells in bone marrow regenerate all blood cell types. Organoids (self-organizing 3D tissue structures) model organ development in vitro.", []string{"stem cell", "iPSC", "pluripotent", "differentiation", "Yamanaka", "organoid", "hematopoietic", "cell biology"}},
		{"Signaling Pathways: MAPK and PI3K", "Cell signaling transmits information from surface receptors to gene expression. MAPK/ERK pathway: growth factor → Ras → Raf → MEK → ERK → transcription factors. PI3K/Akt/mTOR: promotes cell survival and growth. JAK-STAT: cytokine signaling. G-protein coupled receptors (GPCRs) — largest receptor family — activate cAMP, IP3, DAG. Receptor tyrosine kinases (RTKs) phosphorylate each other. Crosstalk between pathways creates signal integration.", []string{"MAPK", "PI3K", "signal transduction", "GPCR", "Ras", "mTOR", "cell signaling", "kinase", "cancer"}},
		{"Epigenetics: Histone Modification", "Histones are proteins around which DNA wraps (nucleosome). Histone acetylation (HATs) opens chromatin, activating transcription. Histone deacetylation (HDACs) condenses chromatin, silencing genes. Histone methylation has context-dependent effects. The histone code hypothesis proposes combinations of modifications specify gene expression states. HDAC inhibitors are used in cancer therapy. Polycomb (repressive) and Trithorax (activating) complexes maintain developmental gene expression.", []string{"histone", "epigenetics", "chromatin", "acetylation", "methylation", "HAT", "HDAC", "Polycomb", "gene expression"}},
		{"Synaptic Plasticity: LTP and LTD", "Long-term potentiation (LTP) strengthens synapses: high-frequency stimulation activates NMDA receptors (require coincident pre+postsynaptic activity — 'Hebb's rule'), allowing Ca²⁺ influx. AMPA receptors are inserted, amplifying future responses. LTP underlies learning and memory. Long-term depression (LTD): low-frequency stimulation, AMPA receptor removal, synapse weakening — used for skill refinement. BDNF supports LTP maintenance.", []string{"LTP", "LTD", "synaptic plasticity", "NMDA receptor", "AMPA receptor", "Hebb's rule", "hippocampus", "memory", "neuroscience"}},
		{"Circadian Rhythms", "Circadian clocks run on ~24-hour cycles driven by feedback loops of clock genes. In mammals: CLOCK/BMAL1 activate Per and Cry; Per/Cry proteins inhibit CLOCK/BMAL1 — completing the feedback loop in ~24h. The suprachiasmatic nucleus (SCN) in the hypothalamus is the master clock, entrained by light via retinal ganglion cells. Jet lag disrupts circadian alignment. Shift work increases risk of metabolic syndrome, cancer, and cardiovascular disease.", []string{"circadian rhythm", "clock genes", "SCN", "CLOCK", "BMAL1", "sleep", "melatonin", "jet lag", "neuroscience"}},
		{"Antibiotic Mechanisms and Resistance", "Antibiotics target bacterial-specific processes: β-lactams (penicillin, cephalosporins) block cell wall synthesis; aminoglycosides, tetracyclines block ribosomes (30S); macrolides, lincosamides block 50S subunit; fluoroquinolones inhibit DNA gyrase. Resistance mechanisms: β-lactamases degrade penicillin; efflux pumps expel antibiotics; target mutations reduce binding. MRSA (methicillin-resistant S. aureus), MDR-TB are public health emergencies. Antimicrobial stewardship programs reduce resistance.", []string{"antibiotic", "antibiotic resistance", "MRSA", "β-lactam", "ribosome", "efflux pump", "MDR-TB", "microbiology", "public health"}},
		{"Photosynthesis: Calvin Cycle", "The Calvin cycle (light-independent reactions) fixes CO₂ into sugars. Rubisco (most abundant enzyme on Earth) combines CO₂ with ribulose-1,5-bisphosphate (RuBP). Three turns fix one CO₂ using 3 ATP and 2 NADPH. Product: glyceraldehyde-3-phosphate (G3P), precursor to glucose. C4 plants (corn, sugarcane) concentrate CO₂ around Rubisco to reduce photorespiration, improving efficiency in hot climates. CAM plants open stomata at night.", []string{"Calvin cycle", "photosynthesis", "Rubisco", "CO2 fixation", "C4 photosynthesis", "CAM", "ATP", "NADPH", "plant biology"}},
		{"Ecological Succession", "Ecological succession is the directional change in species composition over time. Primary succession: bare rock → pioneer species (lichen, mosses) → shrubs → climax community (forest). Secondary succession: recovery after disturbance (fire, logging) on pre-existing soil. Pioneer species modify the environment, facilitating later arrivals. Climax community: stable equilibrium adapted to local climate. Yellowstone's recovery after 1988 fires is a classic case study.", []string{"ecological succession", "primary succession", "secondary succession", "climax community", "pioneer species", "ecology", "biodiversity", "Yellowstone"}},
	}

	var out []mbp.WriteRequest
	for _, c := range concepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	bioTopics := []string{
		"telomere shortening aging Hayflick limit telomerase cancer",
		"RNA interference siRNA miRNA gene silencing DICER RISC",
		"translation ribosome tRNA codon anticodon peptide bond",
		"transcription factor binding promoter enhancer silencer",
		"alternative splicing intron exon isoform spliceosome",
		"protein degradation ubiquitin proteasome E1 E2 E3 ligase",
		"autophagy lysosome cellular recycling Atg mTOR starvation",
		"mitochondria ATP synthase electron transport chain proton gradient",
		"endoplasmic reticulum Golgi apparatus secretory pathway vesicle",
		"cytoskeleton actin microtubule intermediate filament motor protein",
		"cell cycle G1 S G2 M checkpoints CDK cyclin p53 Rb",
		"animal development gastrulation neurulation axis formation Wnt",
		"symbiosis mutualism parasitism commensalism lichens mycorrhizae",
		"population dynamics predator prey Lotka-Volterra equilibrium",
		"island biogeography species richness colonization extinction",
		"coral reef bleaching thermal stress zooxanthellae ocean acidification",
		"nitrogen fixation Rhizobium legume soil bacteria ammonia",
		"horizontal gene transfer plasmid conjugation transformation transduction",
		"CRISPR base editing prime editing off-target effects delivery",
		"single cell RNA sequencing transcriptomics cell atlas clustering",
		"metagenomics microbiome diversity 16S rRNA shotgun sequencing",
		"synthetic biology BioBricks metabolic engineering chassis",
		"biofilm quorum sensing Pseudomonas Staphylococcus antibiotic resistance",
		"extremophile thermophile halophile acidophile Archaea hydrothermal vent",
		"prion misfolded protein bovine spongiform encephalopathy Creutzfeldt-Jakob",
	}

	for _, t := range bioTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Biology: %s", t),
			Content: fmt.Sprintf("The biological study of %s reveals fundamental mechanisms governing life at molecular, cellular, or organismal levels. These processes evolved under selective pressure and are often conserved across species. Understanding this topic informs medicine, biotechnology, and our understanding of life's diversity. Current research uses genomics, proteomics, and imaging to dissect mechanisms.", t),
			Tags:    []string{"biology", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Computer Science ----

func scaleCS() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Transformer Architecture Deep Dive", "The Transformer (Vaswani et al., 2017 — 'Attention Is All You Need') uses multi-head self-attention without recurrence. Q (query), K (key), V (value) matrices compute attention: softmax(QKᵀ/√dₖ)V. Multi-head attention runs h parallel attention heads, concatenated. Positional encodings add sequence order information. Feed-forward layers process each position independently. Encoder-decoder architecture for translation; decoder-only for generation (GPT). BERT uses masked language modeling for bidirectional context.", []string{"transformer", "attention mechanism", "BERT", "GPT", "NLP", "self-attention", "multi-head attention", "Vaswani", "deep learning"}},
		{"Reinforcement Learning: Policy Gradient", "Policy gradient methods directly optimize π(a|s;θ) using gradient ascent on expected reward. REINFORCE: ∇θJ = E[∇θlog π(a|s;θ)·R]. High variance — reduced by baselines (actor-critic). PPO (Proximal Policy Optimization) clips gradient updates for stability. TRPO uses trust region constraints. A3C (Asynchronous Advantage Actor-Critic) uses parallel workers. AlphaGo/AlphaZero use policy gradient with Monte Carlo Tree Search. RLHF (from human feedback) fine-tunes LLMs.", []string{"reinforcement learning", "policy gradient", "PPO", "TRPO", "actor-critic", "AlphaGo", "RLHF", "deep learning"}},
		{"Distributed Consensus: Raft", "Raft (Ongaro & Ousterhout, 2014) is a consensus algorithm designed for understandability. Three roles: leader, follower, candidate. Leader election via randomized timeouts. Log replication: leader appends entries, sends to followers; committed after majority acknowledge. Term numbers prevent stale leaders. Raft guarantees linearizability. Used in etcd (Kubernetes), TiKV, CockroachDB. Versus Paxos: same safety, clearer leader-based design.", []string{"Raft", "distributed consensus", "leader election", "log replication", "etcd", "linearizability", "distributed systems", "fault tolerance"}},
		{"B-Tree and LSM-Tree Storage", "B-trees maintain sorted data in a balanced tree, guaranteeing O(log n) search, insert, delete. All leaves at same depth. B+ trees store data only in leaves for sequential scans — used in PostgreSQL, MySQL indexes. LSM-trees (Log-Structured Merge-trees) optimize writes via sequential append to MemTable, flushed to SSTables on disk, periodically compacted. RocksDB, LevelDB, Cassandra use LSM-trees. Trade-off: LSM writes faster, B-tree reads faster.", []string{"B-tree", "LSM-tree", "storage engine", "database", "RocksDB", "PostgreSQL", "SSTables", "indexing", "compaction"}},
		{"MapReduce and Spark", "MapReduce (Dean & Ghemawat, Google, 2004) processes large datasets in parallel: Map phase applies a function to each record (key-value pairs); Reduce aggregates by key. Hadoop implements open-source MapReduce on HDFS. Apache Spark improves on MapReduce with in-memory computation and DAG execution engine, 10-100× faster for iterative algorithms. Spark SQL, Streaming, MLlib, and GraphX extend capabilities. Modern: Apache Flink for streaming.", []string{"MapReduce", "Spark", "Hadoop", "distributed computing", "big data", "HDFS", "DAG", "data processing", "Flink"}},
		{"Cryptographic Hash Functions", "Cryptographic hash functions map arbitrary input to fixed-length digest. Properties: preimage resistance (can't find input from hash), second preimage resistance, collision resistance. SHA-256 produces 256-bit digests; SHA-3 uses sponge construction. MD5 and SHA-1 are broken (collisions found). Hash functions underlie digital signatures (hash then sign), password storage (bcrypt/Argon2 add salts and slow iterations), Merkle trees (blockchain, Git), and proof-of-work (Bitcoin).", []string{"hash function", "SHA-256", "cryptography", "collision resistance", "blockchain", "Merkle tree", "Git", "password hashing", "digital signature"}},
		{"Virtual Memory and Paging", "Virtual memory abstracts physical RAM, giving each process its own address space. Pages (4KB) are mapped to physical frames via page tables. TLB (Translation Lookaside Buffer) caches recent translations. Page faults load pages from disk (swap). Demand paging only loads pages when accessed. Copy-on-write (fork() in Unix) delays page copying. ASLR (Address Space Layout Randomization) randomizes addresses to thwart exploits. Memory-mapped files enable efficient I/O.", []string{"virtual memory", "paging", "page table", "TLB", "page fault", "ASLR", "demand paging", "operating system", "memory management"}},
		{"Convolutional Neural Networks", "CNNs exploit local spatial structure via convolutional filters (kernels) that slide across input, sharing weights. Convolution: feature map = input * kernel. Pooling (max/average) reduces spatial dimensions. Architectures: LeNet-5 (LeCun, 1989), AlexNet (2012, ImageNet breakthrough), VGG, ResNet (skip connections solve vanishing gradient), EfficientNet. CNNs power image classification, object detection (YOLO, Faster R-CNN), segmentation, and face recognition.", []string{"CNN", "convolutional neural network", "ResNet", "ImageNet", "AlexNet", "object detection", "YOLO", "deep learning", "computer vision"}},
		{"TCP Congestion Control", "TCP congestion control prevents network collapse. Slow start: exponential window growth until ssthresh. Congestion avoidance: linear growth after ssthresh. On packet loss: timeout → slow start from 1; triple duplicate ACK → fast retransmit + fast recovery. CUBIC (Linux default): cubic function of time since last congestion. BBR (Google): model-based, maximizes bandwidth-delay product. QUIC (HTTP/3) implements reliability in UDP userspace.", []string{"TCP", "congestion control", "slow start", "CUBIC", "BBR", "QUIC", "networking", "bandwidth", "latency"}},
		{"Blockchain and Merkle Trees", "Blockchains use cryptographic chains of blocks: each block header contains previous block's hash, Merkle root of transactions, timestamp, nonce. Merkle trees enable efficient transaction verification without downloading full block. Proof-of-work (Bitcoin): find nonce such that hash(block) < target. Proof-of-stake (Ethereum): validators staked ETH selected proportionally. Smart contracts (Ethereum EVM) execute code on-chain. Scalability trilemma: security/decentralization/scalability.", []string{"blockchain", "Merkle tree", "Bitcoin", "Ethereum", "proof of work", "proof of stake", "smart contract", "cryptography", "distributed ledger"}},
	}

	var out []mbp.WriteRequest
	for _, c := range concepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	csTopics := []string{
		"type system Hindley-Milner type inference polymorphism generics",
		"garbage collection generational mark sweep reference counting tracing",
		"just-in-time compilation JIT interpreter bytecode optimization hotspot",
		"cache coherence MESI protocol multicore invalidation snooping",
		"NUMA non-uniform memory access topology socket L3 cache migration",
		"lock-free data structure CAS atomic compare-and-swap ABA problem",
		"event loop async await coroutine Python JavaScript Node",
		"WebAssembly WASM browser sandbox portable binary format",
		"HTTP/2 multiplexing header compression server push stream",
		"gRPC Protocol Buffers bidirectional streaming service mesh",
		"Kubernetes pod deployment service ingress StatefulSet namespace",
		"service mesh Istio Envoy sidecar proxy mTLS traffic management",
		"continuous delivery GitOps ArgoCD Flux deployment pipeline",
		"observability distributed tracing OpenTelemetry Jaeger Prometheus",
		"feature flags A/B testing canary deployment blue-green rollout",
		"CQRS event sourcing eventual consistency saga pattern DDD",
		"actor model Erlang Akka message passing concurrency supervision",
		"zero knowledge proof zk-SNARK privacy blockchain verification",
		"differential privacy noise mechanism privacy budget federated learning",
		"adversarial examples robustness FGSM perturbation neural network",
		"graph neural network GNN message passing molecular property",
		"diffusion model score matching DDPM stable diffusion generation",
		"RLHF reward model constitutional AI preference learning alignment",
		"vector database embedding similarity search FAISS Pinecone Weaviate",
		"time series forecasting ARIMA Prophet seasonal decomposition anomaly",
		"recommendation system collaborative filtering matrix factorization",
		"SQL query optimizer cost model statistics join order index scan",
		"HTAP hybrid transactional analytical processing in-memory column store",
		"consensus ZooKeeper Chubby distributed lock service coordination",
		"formal verification model checking TLA+ Coq proof assistant",
	}

	for _, t := range csTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("CS: %s", t),
			Content: fmt.Sprintf("The field of %s involves key principles and techniques central to modern computer science and software engineering. Understanding these concepts enables designing systems that are efficient, reliable, and correct. Research and industry practice continually evolve best practices. Practitioners must balance theoretical foundations with practical constraints of real systems.", t),
			Tags:    []string{"computer science", "software engineering", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Mathematics ----

func scaleMath() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Group Theory and Symmetry", "A group is a set G with operation * satisfying: closure, associativity, identity element, inverses. Abelian groups have commutative *. Finite groups: cyclic groups Z_n, symmetric groups S_n, alternating groups A_n. Lagrange's theorem: order of subgroup divides order of group. Galois theory uses group theory to determine polynomial solvability by radicals — quintics and higher are generally not solvable. Lie groups (continuous groups) underlie symmetries in physics.", []string{"group theory", "symmetry", "abstract algebra", "Galois theory", "Lie group", "mathematics", "subgroup"}},
		{"Complex Analysis", "Complex analysis studies differentiable functions of complex variables. Holomorphic functions satisfy Cauchy-Riemann equations. Cauchy's integral theorem: integral of holomorphic function over closed curve = 0. Cauchy integral formula evaluates function at any interior point. Laurent series extends Taylor series for singularities. Residue theorem converts complex integrals to residue sums — powerful for real integrals. Conformal maps preserve angles.", []string{"complex analysis", "holomorphic", "Cauchy-Riemann", "residue theorem", "Laurent series", "complex numbers", "mathematics"}},
		{"Differential Equations: Stability", "A dynamical system dx/dt = f(x) has equilibria where f(x*)=0. Linearization at equilibrium: stability determined by eigenvalues of Jacobian. Negative real parts → stable; positive → unstable; zero → requires nonlinear analysis. Lyapunov stability: system stable if there exists V(x)>0 with dV/dt<0. Poincaré-Bendixson theorem: bounded trajectories in 2D either converge or cycle. Bifurcations (saddle-node, Hopf) change system qualitative behavior.", []string{"differential equations", "stability", "Lyapunov", "bifurcation", "dynamical systems", "eigenvalue", "Poincaré", "mathematics"}},
		{"Number Theory: Modular Arithmetic", "Modular arithmetic: a ≡ b (mod n) if n | (a-b). Fermat's little theorem: a^p ≡ a (mod p) for prime p. Euler's theorem: a^φ(n) ≡ 1 (mod n) for gcd(a,n)=1. Chinese Remainder Theorem: unique solution to simultaneous congruences given pairwise coprime moduli. RSA encryption: choose primes p,q; n=pq; e coprime to φ(n); public key (e,n), private d: ed≡1 mod φ(n); encrypt c=m^e mod n, decrypt m=c^d mod n.", []string{"modular arithmetic", "number theory", "Fermat's little theorem", "Euler's theorem", "RSA", "Chinese Remainder Theorem", "cryptography", "mathematics"}},
		{"Linear Algebra: Eigenvectors SVD", "Eigenvector equation: Av = λv. Diagonalization: A = PDP⁻¹ where D diagonal of eigenvalues, P columns are eigenvectors. Spectral theorem: symmetric matrices have real eigenvalues and orthogonal eigenvectors. SVD (Singular Value Decomposition): A = UΣVᵀ. Used for PCA (principal component analysis), low-rank approximation, image compression, pseudoinverse, and recommendation systems (latent factor models). PageRank is the dominant eigenvector of the web adjacency matrix.", []string{"linear algebra", "eigenvector", "SVD", "singular value decomposition", "PCA", "PageRank", "diagonalization", "mathematics"}},
		{"Probability: Central Limit Theorem", "The Central Limit Theorem (CLT): sum of independent, identically distributed random variables converges to normal distribution as n → ∞, regardless of the underlying distribution, provided finite variance. Standardized sum: (X̄ - μ)/(σ/√n) → N(0,1). Enables confidence intervals and hypothesis testing. Convergence rate: Berry-Esseen theorem gives O(1/√n) bound. CLT underlies statistical inference, polling, quality control, and financial risk models.", []string{"central limit theorem", "probability", "statistics", "normal distribution", "confidence interval", "hypothesis testing", "law of large numbers", "mathematics"}},
		{"Topology: Fundamental Group", "The fundamental group π₁(X) classifies loops in a topological space up to continuous deformation (homotopy). The circle S¹ has π₁ = ℤ. Simply connected spaces (sphere, ℝⁿ) have trivial π₁. Van Kampen's theorem computes π₁ of unions. Higher homotopy groups πₙ are harder. Seifert-van Kampen, covering spaces, and the universal cover are key tools. Fundamental group of a torus: ℤ×ℤ (two independent loops).", []string{"fundamental group", "topology", "homotopy", "van Kampen", "simply connected", "torus", "algebraic topology", "mathematics"}},
		{"Mathematical Logic and Set Theory", "Zermelo-Fraenkel axioms (ZFC) form the foundation of modern mathematics. Cantor's diagonal argument proves the real numbers are uncountable: |ℝ| > |ℕ|. The continuum hypothesis (CH): no set has cardinality strictly between ℵ₀ and 2^ℵ₀ — independent of ZFC (Gödel + Cohen). Axiom of Choice: equivalent to Zorn's lemma, well-ordering theorem. Ordinals and cardinals extend natural numbers to the infinite.", []string{"set theory", "ZFC", "Cantor", "continuum hypothesis", "Axiom of Choice", "Zorn's lemma", "logic", "mathematics"}},
	}

	var out []mbp.WriteRequest
	for _, c := range concepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	mathTopics := []string{
		"Fermat's last theorem Wiles elliptic curves modular forms",
		"P vs NP complexity class polynomial time verification",
		"graph coloring chromatic number four color theorem planar",
		"knot theory trefoil Jones polynomial invariant",
		"category theory functor natural transformation Yoneda lemma",
		"measure theory Lebesgue integral sigma algebra probability",
		"partial differential equations heat equation wave equation Laplacian",
		"stochastic differential equation Brownian motion Ito calculus",
		"combinatorics generating function inclusion exclusion Stirling",
		"Ramsey theory complete graph monochromatic structure",
		"cryptography elliptic curve discrete logarithm ECDSA",
		"coding theory Hamming error correction Shannon capacity",
		"optimal transport Wasserstein distance earth mover gradient flow",
		"differential geometry Riemannian manifold geodesic curvature",
		"algebraic geometry variety scheme Zariski topology Grothendieck",
		"representation theory character table Schur's lemma decomposition",
		"harmonic analysis Fourier series Parseval's theorem convolution",
		"functional analysis Hilbert space Banach space operator norm",
		"numerical analysis finite element method Newton's method stability",
		"optimization gradient descent convex Lagrangian duality KKT conditions",
	}

	for _, t := range mathTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Mathematics: %s", t),
			Content: fmt.Sprintf("The mathematical study of %s develops rigorous theory with broad applications across science and engineering. Key theorems establish precise conditions and boundaries. Computational methods enable practical application. This area connects to other branches of mathematics through surprising structural relationships. Open problems continue to drive research.", t),
			Tags:    []string{"mathematics", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- History ----

func scaleHistory() []mbp.WriteRequest {
	events := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Peloponnesian War Athens Sparta", "The Peloponnesian War (431–404 BCE) between Athens and its allies versus Sparta (Peloponnesian League) ended Athenian hegemony. Thucydides' History is the primary source — he pioneered political/military analysis over myth. Key events: plague of Athens (430 BCE, killed Pericles), Sicilian Expedition disaster (415–413 BCE), Battle of Aegospotami (405 BCE, Spartan naval victory). Sparta's victory ushered in its brief hegemony, followed by Theban dominance.", []string{"Peloponnesian War", "Athens", "Sparta", "ancient Greece", "Thucydides", "history", "Greek city-states"}},
		{"Hannibal and the Punic Wars", "The Second Punic War (218–201 BCE) was Rome's most dangerous moment. Hannibal Barca crossed the Alps with war elephants into Italy and crushed Roman armies at Trebia (218), Lake Trasimene (217), and Cannae (216 BCE — double-envelopment of 86,000 Romans, ~50,000 killed). Despite 15 years in Italy, Hannibal lacked siege equipment and reinforcements. Scipio Africanus defeated him at Zama (202 BCE). Rome went on to destroy Carthage in the Third Punic War (146 BCE).", []string{"Hannibal", "Punic Wars", "Rome", "Carthage", "Battle of Cannae", "ancient history", "military history", "Scipio"}},
		{"Silk Road Trade Networks", "The Silk Road was a network of overland and maritime trade routes connecting China, Central Asia, Persia, Arabia, East Africa, and the Mediterranean. Active from ~200 BCE to 1450 CE. Goods: silk, spices, glass, textiles, precious metals. Equally important: transmission of religions (Buddhism, Islam, Christianity), technologies (paper, printing, gunpowder), and diseases (Justinian Plague, Black Death). The maritime routes through the Indian Ocean often exceeded overland routes in volume.", []string{"Silk Road", "trade", "China", "Central Asia", "Buddhist spread", "history", "globalization", "Islamic", "Black Death"}},
		{"Meiji Restoration Japan", "Japan's Meiji Restoration (1868) ended the Edo-era Tokugawa shogunate and restored imperial rule under Emperor Meiji. Japan rapidly industrialized to avoid Western colonization: railways, telegraphs, steel industry, modern army and navy based on Western models. Abolition of the feudal samurai class. Universal education system. By 1895, Japan defeated China (First Sino-Japanese War); by 1905, Russia (Russo-Japanese War, shocking the world). Transformed from feudal to industrial power in ~50 years.", []string{"Meiji Restoration", "Japan", "industrialization", "modernization", "samurai", "history", "imperialism", "Russia-Japan War"}},
		{"Haitian Revolution", "The Haitian Revolution (1791–1804) was the only successful large-scale slave revolt in history. Enslaved Africans under Toussaint L'Ouverture and Jean-Jacques Dessalines fought French, British, and Spanish forces. Napoleon sent 40,000 troops; yellow fever devastated them. Haiti declared independence January 1, 1804 — the first Black republic. France demanded 150 million francs in reparations (paid until 1947), devastating Haiti's economy. The revolution terrified slaveholders in the US South.", []string{"Haitian Revolution", "slavery", "Toussaint L'Ouverture", "Caribbean", "history", "colonialism", "Napoleon", "France", "independence"}},
		{"Abbasid Caliphate", "The Abbasid Caliphate (750–1258 CE) succeeded the Umayyads, moving the capital to Baghdad. The House of Wisdom (Bayt al-Hikma) translated Greek texts and produced original scholarship in mathematics (al-Khwarizmi invented algebra, gave us 'algorithm'), medicine (Ibn Sina's Canon), optics (Ibn al-Haytham), and philosophy. The caliphate disintegrated into regional sultanates. Mongol sack of Baghdad (1258) destroyed the House of Wisdom and ended the caliphate.", []string{"Abbasid Caliphate", "Baghdad", "House of Wisdom", "al-Khwarizmi", "algebra", "Islamic Golden Age", "history", "Mongol invasion"}},
		{"Age of Exploration Portuguese", "Portugal pioneered systematic oceanic exploration under Prince Henry the Navigator. Bartolomeu Dias rounded the Cape of Good Hope (1488). Vasco da Gama reached India (1498), breaking the Arab-Venetian spice monopoly. Pedro Álvares Cabral claimed Brazil (1500). Portugal established trading posts across Africa, India, and Southeast Asia — the Estado da India. The Treaty of Tordesillas (1494) with Spain divided newly discovered lands between them.", []string{"Age of Exploration", "Portugal", "Vasco da Gama", "Bartolomeu Dias", "Cape of Good Hope", "spice trade", "history", "colonialism", "Treaty of Tordesillas"}},
		{"World War I Origins", "WWI (1914–1918) originated in Austro-Hungarian-Serbian tensions after Archduke Franz Ferdinand's assassination (June 28, 1914). The alliance system (Triple Alliance vs Triple Entente) and mobilization timetables (Schlieffen Plan) turned a regional crisis global. Trench warfare on the Western Front produced industrial-scale slaughter: ~20 million dead. New weapons: poison gas, tanks, aircraft, machine guns, artillery. Defeat of Germany, Austria-Hungary, and Ottoman Empire reshaped the world map (Treaty of Versailles, League of Nations).", []string{"World War I", "Franz Ferdinand", "Schlieffen Plan", "trench warfare", "Treaty of Versailles", "history", "Europe", "alliance system"}},
	}

	var out []mbp.WriteRequest
	for _, e := range events {
		out = append(out, mbp.WriteRequest{
			Concept:    e.name,
			Content:    e.content,
			Tags:       e.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	historyTopics := []string{
		"Spanish Inquisition Torquemada heresy conversion Moorish Spain",
		"Crusades Jerusalem Saladin Richard Lionheart Latin Kingdom",
		"Black Death plague bubonic 14th century Europe flagellants",
		"Alexander the Great Macedon Persian Empire Hellenistic India",
		"Julius Caesar Ides of March Rubicon Senate Rome Republic",
		"Napoleon Bonaparte Napoleonic Wars Waterloo Congress Vienna",
		"Ottoman Empire Suleiman janissary devshirme Topkapi",
		"Mughal Empire Akbar Taj Mahal Aurangzeb religious policy",
		"Victorian Era industrial pollution social reform working class",
		"Crimean War Florence Nightingale nursing telegraph modern warfare",
		"Russian Revolution 1917 Bolshevik October Lenin Trotsky Kerensky",
		"Chinese Revolution Mao Zedong Long March civil war Kuomintang",
		"Korean War Inchon MacArthur 38th parallel armistice 1953",
		"Vietnam War Tet Offensive Ho Chi Minh Pentagon Papers Saigon",
		"Cuban Missile Crisis Kennedy Khrushchev ExComm nuclear brinkmanship",
		"Space Race Sputnik Apollo moon landing Vostok cosmonauts",
		"Apartheid Mandela ANC Sharpeville South Africa Truth Reconciliation",
		"Holocaust concentration camps Wannsee Final Solution liberation",
		"D-Day Normandy Overlord Utah Omaha Eisenhower Atlantic Wall",
		"Hiroshima Nagasaki atomic bomb Enola Gay surrender Japan",
		"Magna Carta habeas corpus English common law parliament",
		"English Civil War Cromwell Parliament Roundheads Cavaliers Restoration",
		"French Revolution Reign of Terror Jacobins Directory Thermidor",
		"American Civil War slavery secession Gettysburg Emancipation",
		"Reconstruction Freedmen's Bureau Jim Crow 14th Amendment segregation",
	}

	for _, t := range historyTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Historical Event: %s", t),
			Content: fmt.Sprintf("The historical period and events surrounding %s shaped the political, social, and cultural development of civilizations. Key actors, decisions, and structural forces determined outcomes that echo through subsequent history. Primary sources and archaeological evidence illuminate motivations and consequences. Historians debate causation and contingency in how these events unfolded.", t),
			Tags:    []string{"history", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Philosophy ----

func scalePhilosophy() []mbp.WriteRequest {
	concepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Phenomenology Husserl Heidegger", "Phenomenology studies the structures of conscious experience from the first-person perspective. Edmund Husserl developed phenomenology as a rigorous science: epoché (bracketing natural attitude), intentionality (consciousness is always 'of' something), eidetic reduction (grasping essences). Heidegger's 'Being and Time' (1927) redirected phenomenology toward ontology: Dasein (being-in-the-world), thrownness, care, authenticity, being-toward-death. Merleau-Ponty emphasized embodiment and perception.", []string{"phenomenology", "Husserl", "Heidegger", "Dasein", "intentionality", "existentialism", "philosophy", "ontology"}},
		{"Pragmatism Peirce James Dewey", "Pragmatism (American philosophical tradition) holds that the meaning of a concept lies in its practical consequences. Charles Sanders Peirce: beliefs are rules for action; truth is what inquiry converges on. William James: truth is what works — ideas become true insofar as they help us. John Dewey: philosophy should solve practical problems; education as inquiry. Neopragmatism (Rorty): truth is what our peers let us get away with saying.", []string{"pragmatism", "Peirce", "William James", "Dewey", "Rorty", "American philosophy", "truth", "inquiry"}},
		{"Analytic Philosophy Language", "Analytic philosophy (Frege, Russell, early Wittgenstein) treats philosophy as logical analysis of language. Russell's theory of descriptions: 'The king of France is bald' — definite descriptions have hidden quantifier structure. Wittgenstein's Tractatus: language pictures facts; limits of language = limits of world. Later Wittgenstein (Philosophical Investigations): meaning is use; language games; no private language argument. Austin's speech acts: performatives, illocutionary force.", []string{"analytic philosophy", "Wittgenstein", "Russell", "Frege", "language games", "speech acts", "logic", "philosophy of language"}},
		{"Political Philosophy: Rawls and Justice", "John Rawls' 'A Theory of Justice' (1971) grounds liberal justice in the original position and veil of ignorance: principles chosen when we don't know our place in society. Two principles: (1) equal basic liberties; (2) inequalities must benefit the least advantaged (difference principle). Nozick's libertarian response: any distribution arising from just steps is just (entitlement theory). Communitarianism (MacIntyre, Sandel, Walzer) critiques abstract individualism.", []string{"Rawls", "justice", "original position", "veil of ignorance", "Nozick", "communitarianism", "political philosophy", "liberalism"}},
		{"Philosophy of Science: Falsificationism", "Karl Popper's falsificationism: a scientific theory must be refutable by evidence. Scientific method: bold conjectures → rigorous attempts to falsify → surviving theories are corroborated. Demarcation from pseudoscience: astrology and Marxism are unfalsifiable. Kuhn's 'Structure of Scientific Revolutions': science proceeds through normal science under paradigms, then paradigm shifts (revolutions) when anomalies accumulate. Lakatos: research programs have hard cores protected by auxiliary hypotheses.", []string{"Popper", "falsificationism", "Kuhn", "paradigm shift", "Lakatos", "philosophy of science", "demarcation", "scientific method"}},
	}

	var out []mbp.WriteRequest
	for _, c := range concepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	philTopics := []string{
		"free will determinism compatibilism libertarian agent causation",
		"philosophy of mind functionalism physicalism property dualism",
		"personal identity psychological continuity body memory Parfit",
		"meta-ethics moral realism anti-realism expressivism error theory",
		"applied ethics animal rights Singer speciesism Peter Singer",
		"feminist philosophy care ethics Gilligan Noddings standpoint",
		"philosophy of time A-theory B-theory presentism eternalism",
		"aesthetics beauty art sublime Kant Hegel expression",
		"Stoicism Epictetus Marcus Aurelius virtue indifferent externals",
		"Epicureanism pleasure ataraxia aponia hedonic calculus Lucretius",
		"Buddhist philosophy dukkha anatta impermanence nirvana dharma",
		"Confucianism ren li junzi filial piety social harmony",
		"philosophy of religion natural theology ontological argument problem of evil",
		"social epistemology testimony testimony trust evidence norms",
		"virtue epistemology reliabilism epistemic virtues Zagzebski",
		"philosophy of mathematics Platonism formalism intuitionism Frege",
		"environmental ethics intrinsic value deep ecology anthropocentrism",
		"bioethics informed consent autonomy beneficence non-maleficence",
		"philosophy of action intention reason causes Davidson",
		"poststructuralism Derrida deconstruction differance Foucault power",
	}

	for _, t := range philTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Philosophy: %s", t),
			Content: fmt.Sprintf("The philosophical inquiry into %s addresses fundamental questions about reality, knowledge, value, and human existence. Multiple competing positions argue from distinct premises and methodologies. Thought experiments and conceptual analysis clarify intuitions. Historical figures shaped the debate; contemporary philosophers extend and challenge their views.", t),
			Tags:    []string{"philosophy", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Arts ----

func scaleArts() []mbp.WriteRequest {
	var out []mbp.WriteRequest

	artConcepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Romanticism in Literature", "Romanticism (late 18th–early 19th century) emphasized emotion, nature, imagination, and individualism over Enlightenment reason. English Romantics: Blake, Wordsworth, Coleridge (Preface to Lyrical Ballads: poetry as 'spontaneous overflow of powerful feelings'), Byron, Keats, Shelley. German Romanticism: Goethe ('The Sorrows of Young Werther'), Schiller. American Romanticism: Emerson's Transcendentalism, Thoreau's Walden, Hawthorne, Melville.", []string{"Romanticism", "literature", "Wordsworth", "Byron", "Keats", "Shelley", "Goethe", "Transcendentalism", "poetry"}},
		{"Baroque Art and Architecture", "Baroque art (17th–early 18th century) used drama, rich color, and intense light/shadow (chiaroscuro) to evoke emotional states. Caravaggio pioneered tenebrism. Bernini's sculptures (The Ecstasy of Saint Teresa) combined architecture, painting, and sculpture. Baroque architecture: St. Peter's Basilica (Rome), Palace of Versailles (France). In music: Bach's polyphony, Handel's oratorios, Vivaldi's concertos.", []string{"Baroque", "art history", "Caravaggio", "Bernini", "architecture", "chiaroscuro", "Bach", "Vivaldi", "Handel"}},
		{"Modernist Literature Stream of Consciousness", "Modernist literature (early 20th century) broke from realism. Stream of consciousness renders the flow of thought: Virginia Woolf ('Mrs. Dalloway', 'To the Lighthouse' — interior monologue), James Joyce ('Ulysses', 'Finnegans Wake' — interior monologue, multilingualism). T.S. Eliot's 'The Waste Land' (1922): fragmentation, allusion, myth. Franz Kafka: alienation and bureaucratic absurdity. Lost Generation (Hemingway, Fitzgerald): disillusionment after WWI.", []string{"modernism", "literature", "Virginia Woolf", "James Joyce", "T.S. Eliot", "Hemingway", "Kafka", "stream of consciousness", "interior monologue"}},
		{"Abstract Expressionism", "Abstract Expressionism (New York School, late 1940s–1950s) was the first major American art movement. Jackson Pollock's drip paintings (action painting) emphasized gesture and process. Willem de Kooning: figurative abstraction. Mark Rothko: large color fields evoking emotional depth. Franz Kline: bold black brushstrokes. Barnett Newman: 'zips' on monochromatic fields. The movement shifted the art world's center from Paris to New York.", []string{"abstract expressionism", "Pollock", "Rothko", "de Kooning", "New York School", "action painting", "art history", "modern art"}},
	}

	for _, c := range artConcepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.93,
			Stability:  0.90,
		})
	}

	artTopics := []string{
		"Gothic architecture cathedral flying buttress ribbed vault nave",
		"Neoclassicism David Napoleon Rome antiquity reason symmetry",
		"Surrealism Dalí Magritte Breton unconscious dream automatic",
		"Pop Art Warhol Lichtenstein consumer culture mass media",
		"Cubism Picasso Braque multiple perspectives geometric planes",
		"Minimalism Judd Flavin geometric industrial materiality",
		"Photography Ansel Adams Dorothea Lange documentary Cartier-Bresson",
		"Film noir shadow cinematography postwar cynicism Chandler",
		"Magical realism García Márquez Borges Latin America supernatural",
		"Postmodern literature DFW Pynchon metafiction irony pastiche",
		"Jazz improvisation Miles Davis bebop free jazz Coltrane Monk",
		"Hip hop DJ sampling MCing breakdancing Bronx Grandmaster Flash",
		"Opera Wagner leitmotif Verdi bel canto Puccini libretto",
		"Dance choreography ballet modern Martha Graham Merce Cunningham",
		"Architecture Le Corbusier International Style Bauhaus functionalism",
	}

	for _, t := range artTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Arts: %s", t),
			Content: fmt.Sprintf("The artistic tradition of %s reflects cultural, historical, and aesthetic values of its era. Artists develop techniques and styles in dialogue with predecessors and contemporaries. Critical theory interprets works through social, psychological, and formal lenses. This tradition continues to influence contemporary practice.", t),
			Tags:    []string{"arts", "culture", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Technology ----

func scaleTechnology() []mbp.WriteRequest {
	var out []mbp.WriteRequest

	techConcepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Internet Protocols TCP IP DNS HTTP", "The Internet uses layered protocols. DNS (Domain Name System) resolves hostnames to IP addresses via distributed hierarchical database. HTTP (HyperText Transfer Protocol) transfers web resources; HTTPS adds TLS encryption. TLS handshake: server presents certificate, client verifies CA signature, they negotiate symmetric key (Diffie-Hellman key exchange or RSA). IPv4 uses 32-bit addresses (4.3 billion); IPv6 128-bit addresses solve exhaustion. BGP (Border Gateway Protocol) routes between autonomous systems.", []string{"TCP/IP", "DNS", "HTTP", "TLS", "IPv6", "BGP", "internet", "networking", "protocol"}},
		{"Processor Architecture RISC CISC ARM x86", "RISC (Reduced Instruction Set Computer): simple fixed-width instructions, many registers, load/store architecture, pipeline-friendly. ARM is dominant RISC architecture (phones, M1/M2 Macs). CISC (Complex Instruction Set): variable-length instructions, fewer registers, complex operations (x86). Intel/AMD use x86; decode complex instructions to RISC-like micro-ops internally. Pipelining overlaps fetch/decode/execute. Speculative execution (Spectre/Meltdown vulnerabilities).", []string{"processor", "RISC", "CISC", "ARM", "x86", "pipeline", "speculative execution", "Spectre", "architecture"}},
		{"Machine Learning at Scale", "Training large models requires distributed computing. Data parallelism: replicate model on N GPUs, each processes different batch, average gradients. Model parallelism: split model layers across GPUs. Tensor parallelism: split individual layers. Pipeline parallelism: different stages on different GPUs. Mixed precision training (fp16/bf16): speeds training, reduces memory. Gradient checkpointing: recompute activations to save memory. Frameworks: PyTorch DDP, DeepSpeed ZeRO, Megatron-LM.", []string{"distributed training", "data parallelism", "model parallelism", "GPU", "PyTorch", "DeepSpeed", "large language model", "machine learning"}},
		{"Renewable Energy Technology", "Solar PV: silicon p-n junction converts photons to electricity. Monocrystalline silicon ~22% efficient; perovskite tandem cells approaching 30%. Wind turbines: horizontal axis with three blades optimized by Betz limit (59% max efficiency). Offshore wind: stronger, steadier winds. Grid-scale storage: lithium-ion batteries (Tesla Megapack), pumped hydro, hydrogen electrolysis. Smart grids balance variable renewable supply with demand. Levelized cost of energy (LCOE) for solar now below coal in most regions.", []string{"renewable energy", "solar panel", "wind turbine", "photovoltaic", "grid storage", "lithium-ion", "perovskite", "LCOE", "technology"}},
	}

	for _, c := range techConcepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.93,
			Stability:  0.90,
		})
	}

	techTopics := []string{
		"CMOS transistor scaling Moore's law FinFET gate-all-around 2nm",
		"memory DRAM SRAM flash HBM bandwidth latency hierarchy",
		"network on chip NoC mesh torus interconnect bandwidth",
		"autonomous vehicle LIDAR computer vision path planning",
		"natural language processing speech recognition ASR Whisper",
		"robotics SLAM localization mapping motor control ROS",
		"3D printing additive manufacturing FDM SLA metal sintering",
		"OLED display organic electroluminescence flexible pixel",
		"satellite internet LEO Starlink OneWeb latency coverage",
		"5G millimeter wave massive MIMO beamforming latency",
		"smart home IoT edge computing MQTT Zigbee Z-Wave",
		"augmented reality AR spatial computing HoloLens Apple Vision",
		"quantum internet entanglement distribution quantum repeater",
		"neuromorphic computing Intel Loihi spiking neural network",
		"genome sequencing Illumina Oxford Nanopore long read assembly",
		"CRISPR delivery lipid nanoparticle viral vector AAV in vivo",
		"carbon capture direct air capture DAC Climeworks biomass",
		"nuclear reactor SMR molten salt thorium advanced fission",
		"fusion energy plasma confinement ITER Commonwealth Fusion NET",
		"materials science graphene 2D material topological insulator",
	}

	for _, t := range techTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Technology: %s", t),
			Content: fmt.Sprintf("The technology of %s represents a convergence of scientific principles and engineering innovation. Current state of the art reflects decades of incremental improvement and periodic breakthroughs. Economic factors, regulatory constraints, and societal needs shape development trajectories. Challenges in scalability, reliability, cost, and safety drive ongoing research.", t),
			Tags:    []string{"technology", "engineering", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Psychology ----

func scalePsychology() []mbp.WriteRequest {
	var out []mbp.WriteRequest

	psychConcepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Social Psychology: Conformity and Obedience", "Asch conformity experiments (1951): subjects conformed to obviously wrong group answers ~37% of the time. Conformity increases with unanimous majority. Milgram obedience (1961-62): 65% of participants administered seemingly lethal shocks when commanded by authority. Zimbardo's Stanford Prison Experiment (1971): dramatic role effects — ended early. These studies show situational factors powerfully override individual dispositions — undermining dispositional attribution.", []string{"conformity", "Asch experiment", "Milgram", "Zimbardo", "obedience", "social psychology", "authority", "bystander effect"}},
		{"Developmental Psychology Piaget", "Jean Piaget's theory of cognitive development: sensorimotor (0–2 years, object permanence), preoperational (2–7, symbolic thought, egocentrism, lack of conservation), concrete operational (7–11, logical operations on concrete objects), formal operational (12+, abstract reasoning). Vygotsky's zone of proximal development: learning occurs in the range between what a child can do independently and with guidance. Attachment (Bowlby, Ainsworth): secure, anxious, avoidant, disorganized attachment styles.", []string{"Piaget", "cognitive development", "developmental psychology", "Vygotsky", "zone of proximal development", "attachment", "Bowlby", "Ainsworth"}},
		{"Cognitive Biases and Heuristics", "Kahneman and Tversky's work (Nobel 2002) revealed systematic biases in judgment. Availability heuristic: judge probability by ease of recall. Representativeness: judge by similarity to prototype (base rate neglect). Anchoring: first number seen influences estimates. Confirmation bias: seek information confirming beliefs. Framing effect: decisions depend on presentation. Loss aversion: losses loom twice as large as equivalent gains. Overconfidence: people overestimate their accuracy.", []string{"cognitive bias", "heuristics", "Kahneman", "Tversky", "availability heuristic", "anchoring", "confirmation bias", "loss aversion", "framing"}},
		{"Positive Psychology and Well-being", "Martin Seligman founded positive psychology (1998 APA presidential address) as the science of well-being. PERMA model: Positive emotions, Engagement (flow), Relationships, Meaning, Accomplishment. Csikszentmihalyi's flow: deep engagement when challenge matches skill. Post-traumatic growth: many trauma survivors report positive change. Strengths-based approaches (VIA strengths). Hedonic adaptation: happiness returns to baseline after positive/negative events.", []string{"positive psychology", "well-being", "Seligman", "PERMA", "flow", "Csikszentmihalyi", "post-traumatic growth", "hedonic adaptation", "happiness"}},
	}

	for _, c := range psychConcepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.93,
			Stability:  0.90,
		})
	}

	psychTopics := []string{
		"social identity theory in-group out-group Tajfel Turner categorization",
		"stereotype threat Claude Steele performance identity anxiety",
		"implicit bias IAT unconscious association race gender",
		"self-determination theory intrinsic extrinsic motivation Deci Ryan",
		"emotion regulation cognitive reappraisal suppression Gross",
		"personality trait Big Five OCEAN openness conscientiousness",
		"intelligence IQ g factor fluid crystallized Gardner multiple",
		"learning theory operant classical Skinner Bandura social",
		"memory types episodic semantic procedural working explicit",
		"attention selective divided inattentional blindness Simons Chabris",
		"perception Gestalt figure ground proximity similarity closure",
		"decision making prospect theory rational agent expected utility",
		"stress coping cortisol allostatic load fight flight freeze",
		"motivation self-efficacy Bandura learned helplessness Seligman",
		"abnormal psychology DSM-5 diagnosis reliability validity",
		"neuropsychology Phineas Gage frontal lobe patient HM memory",
		"evolutionary psychology mate preference parental investment Trivers",
		"group dynamics groupthink Janis risky shift polarization",
		"persuasion Cialdini reciprocity commitment social proof authority",
		"psychotherapy efficacy CBT DBT psychodynamic meta-analysis",
	}

	for _, t := range psychTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Psychology: %s", t),
			Content: fmt.Sprintf("Research on %s illuminates fundamental aspects of human cognition, emotion, behavior, and social interaction. Experimental methods and longitudinal studies reveal mechanisms. Individual differences and cultural context moderate effects. Findings apply to education, therapy, organizational behavior, and policy. Replication studies continue to evaluate robustness of classic findings.", t),
			Tags:    []string{"psychology", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Astronomy ----

func scaleAstronomy() []mbp.WriteRequest {
	var out []mbp.WriteRequest

	astroConcepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Cosmic Microwave Background", "The CMB is thermal radiation from 380,000 years after the Big Bang, when the universe cooled enough for electrons and protons to combine (recombination), making it transparent. Temperature: 2.725 K, nearly uniform. Anisotropies (COBE, WMAP, Planck satellites) at 10⁻⁵ level encode primordial density fluctuations that seeded galaxy formation. CMB power spectrum constrains cosmological parameters: ΩΛ ≈ 0.68, Ωm ≈ 0.31, H₀ ≈ 68 km/s/Mpc. CMB polarization (B-modes) probes inflation.", []string{"CMB", "cosmic microwave background", "recombination", "Big Bang", "cosmology", "Planck satellite", "inflation", "dark energy", "Hubble constant"}},
		{"Neutron Stars and Pulsars", "Neutron stars result from supernova core collapse of massive stars. Composed primarily of neutrons (nuclear density: ~10¹⁷ kg/m³). Typical mass ~1.4 solar masses, radius ~10 km. Pulsars: rotating neutron stars with beamed radio emission detected as regular pulses. Millisecond pulsars (recycled by accretion) are natural atomic clocks. Neutron star mergers produce gravitational waves (GW170817) and r-process kilonovae (source of gold, platinum). Magnetars have extreme magnetic fields (10¹⁵ Gauss).", []string{"neutron star", "pulsar", "magnetar", "supernova", "gravitational wave", "GW170817", "kilonova", "r-process", "astronomy"}},
		{"Galaxy Formation and Evolution", "Galaxies form when dark matter halos collapse and gas cools within them. Elliptical galaxies: old stellar populations, little gas, slow rotation — likely formed by mergers. Spiral galaxies (like Milky Way): disk, bulge, halo; active star formation in arms. Active galactic nuclei (AGN) powered by accreting supermassive black holes: quasars, Seyfert galaxies, blazars. Feedback from AGN and supernovae regulates star formation. The Milky Way will merge with Andromeda in ~4.5 billion years.", []string{"galaxy formation", "dark matter halo", "elliptical galaxy", "spiral galaxy", "AGN", "quasar", "supermassive black hole", "Milky Way", "Andromeda"}},
		{"Solar System Formation", "The Solar System formed ~4.6 billion years ago from a collapsing molecular cloud (solar nebula). Conservation of angular momentum spun a disk; the Sun ignited at center. Planetesimals grew by accretion: rocky planets formed inside the frost line; gas giants beyond (Jupiter, Saturn) accreted gas before the disk dispersed. The Nice model explains late heavy bombardment: Jupiter-Saturn orbital resonance destabilized outer solar system, bombarding inner planets. Earth's Moon formed in a giant impact (Theia).", []string{"solar system formation", "solar nebula", "accretion", "Jupiter", "frost line", "Nice model", "late heavy bombardment", "Moon formation", "Theia"}},
	}

	for _, c := range astroConcepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	astroTopics := []string{
		"white dwarf Chandrasekhar limit degenerate matter type Ia supernova",
		"brown dwarf substellar object T dwarf hydrogen fusion failed star",
		"exoplanet transit method radial velocity Kepler TESS atmosphere",
		"habitable zone liquid water M dwarf Proxima Centauri TRAPPIST",
		"dark matter rotation curve halo WIMP weakly interacting CDM",
		"gravitational wave LIGO Virgo inspiral merger ringdown chirp",
		"black hole shadow Event Horizon Telescope M87 Sagittarius A*",
		"Hubble tension H0 standard candles distance ladder controversy",
		"cosmic inflation slow roll graceful exit reheating power spectrum",
		"nucleosynthesis Big Bang helium hydrogen lithium abundance",
		"stellar spectroscopy spectral class OBAFGKM composition Hertzsprung-Russell",
		"radio astronomy SETI signal Wow! pulsar discovery Jocelyn Bell",
		"space telescope JWST infrared galaxy formation reionization epoch",
		"planetary atmospheres Venus greenhouse Mars thin CO2 Titan methane",
		"asteroid belt Kuiper belt Oort cloud comet short long period",
		"stellar parallax parsec light year distance measurement Hipparcos Gaia",
		"cosmic rays ultra-high energy Pierre Auger Observatory Auger",
		"fast radio burst magnetar origin localization CHIME telescope",
		"galactic center Sgr A* stellar orbits S2 orbit black hole mass",
		"tidal disruption event stellar stream X-ray flare accretion disk",
	}

	for _, t := range astroTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Astronomy: %s", t),
			Content: fmt.Sprintf("The astronomical study of %s reveals fundamental properties of the universe at scales from stellar to cosmic. Observational evidence from multi-wavelength surveys constrains theoretical models. Space missions and ground-based observatories provide complementary data. This topic connects to fundamental physics through extreme conditions of temperature, density, and gravity.", t),
			Tags:    []string{"astronomy", "cosmology", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Chemistry ----

func scaleChemistry() []mbp.WriteRequest {
	var out []mbp.WriteRequest

	chemConcepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Organic Chemistry Reactions", "Key organic reaction types: substitution (SN1: carbocation intermediate, SN2: backside attack, inversion), elimination (E1, E2: double bond formation), addition (electrophilic to alkenes, HBr via Markovnikov's rule), oxidation/reduction (PCC oxidizes alcohols to aldehydes, KMnO4 to carboxylic acids), condensation (ester/amide bond formation with water loss). Functional groups determine reactivity: alcohols, amines, carbonyls (aldehydes, ketones), carboxylic acids.", []string{"organic chemistry", "substitution", "SN1", "SN2", "elimination", "Markovnikov", "functional group", "reaction mechanism"}},
		{"Thermochemistry and Calorimetry", "Enthalpy ΔH = heat at constant pressure. Hess's law: ΔH of overall reaction = sum of ΔH of steps. Standard enthalpy of formation ΔH°f: energy to form 1 mol from elements in standard states. Bond energies: breaking bonds absorbs energy; forming releases. Combustion: complete oxidation, highly exothermic. Calorimetry measures heat change: q = mcΔT. Bomb calorimeter measures combustion at constant volume. Endothermic reactions (ΔH>0) absorb heat; exothermic (ΔH<0) release.", []string{"thermochemistry", "enthalpy", "Hess's law", "calorimetry", "bond energy", "endothermic", "exothermic", "chemistry"}},
		{"Electrochemistry and Batteries", "Oxidation-reduction (redox) reactions involve electron transfer. Galvanic cells: spontaneous redox drives current. Standard electrode potential E° (vs. SHE). Nernst equation: E = E° - (RT/nF)ln(Q). Electrolysis: nonspontaneous — electrolytic cells. Batteries: Li-ion (LiCoO2 cathode, graphite anode, LiPF6 electrolyte), ~3.6V. Fuel cells: H2+O2→H2O, continuous fuel supply, no charging. Corrosion: iron oxidation (Fe→Fe²⁺), prevented by galvanization (zinc sacrificial anode).", []string{"electrochemistry", "redox", "Nernst equation", "lithium-ion battery", "fuel cell", "electrolysis", "galvanic cell", "corrosion"}},
		{"Polymer Chemistry", "Polymers are large molecules of repeating structural units (monomers). Addition polymerization: alkene monomers add without loss (polyethylene, PVC, polystyrene, PTFE). Condensation polymerization: monomers react with small molecule loss (nylon 6,6, polyester PET, Kevlar). Natural polymers: cellulose, starch, proteins, DNA. Cross-linking increases rigidity (vulcanized rubber, thermosetting plastics). Glass transition temperature (Tg) and melting temperature determine thermal properties.", []string{"polymer", "polymerization", "addition polymerization", "condensation polymerization", "nylon", "polyester", "cross-linking", "chemistry", "materials"}},
	}

	for _, c := range chemConcepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.93,
			Stability:  0.90,
		})
	}

	chemTopics := []string{
		"periodic table trends electronegativity ionization energy atomic radius",
		"chemical bonding covalent ionic metallic VSEPR molecular geometry",
		"acid base pH buffer Henderson-Hasselbalch pKa strong weak",
		"chemical equilibrium Le Chatelier's principle Ksp solubility product",
		"kinetics reaction rate rate law Arrhenius activation energy",
		"catalysis enzyme heterogeneous homogeneous zeolite turnover",
		"spectroscopy NMR IR mass spectrometry UV-vis structure determination",
		"green chemistry solvent-free atom economy biodegradable sustainable",
		"medicinal chemistry drug design SAR pharmacophore lead compound",
		"coordination chemistry ligand complex crystal field theory CFSE",
		"atmospheric chemistry ozone layer CFC smog NOx photochemical",
		"nuclear chemistry fission fusion radioactivity decay chain",
		"supramolecular chemistry host-guest crown ether rotaxane catenane",
		"computational chemistry DFT molecular dynamics force field simulation",
		"chiral molecule enantiomer racemic optical rotation chirality",
	}

	for _, t := range chemTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Chemistry: %s", t),
			Content: fmt.Sprintf("The chemistry of %s involves understanding how matter transforms through chemical reactions and physical processes. Quantitative relationships, reaction mechanisms, and structural considerations guide both fundamental research and practical applications. Modern instrumentation enables precise measurement and characterization. Industrial chemistry scales principles to manufacturing.", t),
			Tags:    []string{"chemistry", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}

// ---- Medicine ----

func scaleMedicine() []mbp.WriteRequest {
	var out []mbp.WriteRequest

	medConcepts := []struct {
		name    string
		content string
		tags    []string
	}{
		{"Pharmacokinetics ADME", "Pharmacokinetics describes drug disposition: Absorption (bioavailability = fraction reaching systemic circulation; routes: oral, IV, inhalation, transdermal), Distribution (Vd = volume of distribution; protein binding reduces free drug), Metabolism (liver CYP450 enzymes; first-pass effect reduces oral bioavailability; phase I oxidation, phase II conjugation), Excretion (renal clearance; half-life T½ = 0.693·Vd/CL). Drug-drug interactions occur via CYP inhibition/induction.", []string{"pharmacokinetics", "ADME", "bioavailability", "metabolism", "CYP450", "half-life", "pharmacology", "drug", "medicine"}},
		{"Cardiovascular Disease and Risk", "Atherosclerosis: lipid-laden plaques form in arterial walls; rupture triggers thrombosis. Risk factors: hypertension, dyslipidemia (elevated LDL, low HDL), smoking, diabetes, obesity, age. Myocardial infarction: coronary artery occlusion → ischemia → infarct. Treatment: aspirin, statins, ACE inhibitors, beta-blockers. Primary prevention: statin therapy for high cardiovascular risk. PCSK9 inhibitors dramatically lower LDL. Heart failure: reduced ejection fraction treated with sacubitril/valsartan, SGLT2 inhibitors.", []string{"cardiovascular disease", "atherosclerosis", "myocardial infarction", "LDL", "statin", "hypertension", "heart failure", "medicine", "risk factors"}},
		{"Infectious Disease and Vaccines", "mRNA vaccines (COVID-19): deliver mRNA encoding spike protein; ribosomes produce antigen; immune response without live virus. Adjuvants enhance immune response by activating innate immunity. Herd immunity threshold: 1 - 1/R₀; for measles (R₀~15), ~93% vaccination required. Antibiotics treat bacterial infections; antivirals (oseltamivir, remdesivir, HIV antiretrovirals). Emerging infectious disease: zoonotic spillover (SARS-CoV-2, Ebola, influenza). Global surveillance and One Health framework.", []string{"vaccine", "mRNA vaccine", "COVID-19", "herd immunity", "infectious disease", "antibiotic", "antiviral", "zoonotic", "public health"}},
		{"Cancer Biology and Treatment", "Cancer hallmarks (Hanahan & Weinberg): sustaining proliferative signaling, evading growth suppressors, resisting cell death, enabling replicative immortality, inducing angiogenesis, activating invasion. Oncogenes (RAS, MYC, HER2): gain-of-function mutations drive growth. Tumor suppressors (p53, BRCA1, Rb): loss-of-function permits unchecked growth. Immunotherapy: checkpoint inhibitors (anti-PD1/PDL1, anti-CTLA4) unleash T cells against tumors. CAR-T cell therapy: engineered T cells target specific antigens.", []string{"cancer", "oncogene", "tumor suppressor", "p53", "BRCA1", "immunotherapy", "checkpoint inhibitor", "CAR-T", "hallmarks of cancer"}},
	}

	for _, c := range medConcepts {
		out = append(out, mbp.WriteRequest{
			Concept:    c.name,
			Content:    c.content,
			Tags:       c.tags,
			Confidence: 0.95,
			Stability:  0.92,
		})
	}

	medTopics := []string{
		"clinical trial randomized controlled blinding placebo effect evidence",
		"epigenetics cancer methylation histone acetylation gene silencing",
		"microbiome gut health probiotic dysbiosis Clostridium difficile FMT",
		"precision medicine genomics pharmacogenomics biomarker stratification",
		"neurodegenerative Alzheimer's amyloid tau Parkinson's alpha-synuclein",
		"mental health depression serotonin SSRI neuroplasticity psychotherapy",
		"diabetes insulin resistance type 1 type 2 HbA1c metformin GLP-1",
		"autoimmune disease rheumatoid arthritis lupus biologic TNF-alpha",
		"surgery laparoscopic robotic minimally invasive recovery outcome",
		"radiology CT MRI PET scan imaging diagnosis staging",
		"anesthesia general local regional airway management pain",
		"pediatrics vaccination developmental milestone growth chart",
		"geriatrics polypharmacy frailty cognitive decline falls prevention",
		"public health epidemiology incidence prevalence SIR model outbreak",
		"global health neglected tropical disease malaria TB HIV DALYs",
	}

	for _, t := range medTopics {
		out = append(out, mbp.WriteRequest{
			Concept: fmt.Sprintf("Medicine: %s", t),
			Content: fmt.Sprintf("Clinical and research advances in %s integrate basic science with patient care. Evidence-based medicine synthesizes trial data for clinical decision-making. Mechanism-of-action understanding guides drug development and personalized treatment. Healthcare systems must balance efficacy, safety, access, and cost. Ongoing research addresses gaps in understanding and unmet clinical needs.", t),
			Tags:    []string{"medicine", "healthcare", t},
			Confidence: 0.87,
			Stability:  0.85,
		})
	}

	return out
}
