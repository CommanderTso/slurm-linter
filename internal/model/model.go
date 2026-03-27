package model

// SlurmConfig is the parsed in-memory representation of slurm.conf.
type SlurmConfig struct {
	// Globals holds top-level key=value parameters (e.g. ClusterName, AuthType).
	// Key is the parameter name; value is the raw string value.
	Globals map[string]string

	// GlobalLines maps each global parameter name to its 1-based line number.
	GlobalLines map[string]int

	// Nodes holds all NodeName= stanza definitions.
	Nodes []NodeDef

	// Partitions holds all PartitionName= stanza definitions.
	Partitions []Partition
}

// NodeDef represents one NodeName= line in slurm.conf.
// Name may be a bracket range expression like "node[01-04]".
type NodeDef struct {
	Name   string            // raw name expression
	Params map[string]string // additional params (CPUs, RealMemory, State, etc.)
	Line   int               // 1-based line number
}

// Partition represents one PartitionName= line in slurm.conf.
type Partition struct {
	Name   string
	Params map[string]string // Nodes, MaxTime, Default, State, etc.
	Line   int
}

// TopologyConfig is the parsed in-memory representation of topology.conf.
type TopologyConfig struct {
	Switches []Switch
}

// Switch represents one SwitchName= line in topology.conf.
// Exactly one of Nodes or Switches will be non-empty.
type Switch struct {
	Name     string // switch name
	Nodes    string // raw node range expression (e.g. "node[01-04]"), or ""
	Switches string // raw switch range expression (e.g. "s[0-2]"), or ""
	Line     int    // 1-based line number
}
