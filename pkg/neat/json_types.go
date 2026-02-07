package neat

import (
	"encoding/json"
	"fmt"
	"strings"
)

var activationNames = map[ActivationType]string{
	ActivationLinear:   "linear",
	ActivationSigmoid:  "sigmoid",
	ActivationTanh:     "tanh",
	ActivationRelu:     "relu",
	ActivationSin:      "sin",
	ActivationCos:      "cos",
	ActivationGaussian: "gaussian",
	ActivationAbs:      "abs",
	ActivationSquare:   "square",
}

var activationValues = func() map[string]ActivationType {
	out := make(map[string]ActivationType, len(activationNames))
	for k, v := range activationNames {
		out[v] = k
	}
	return out
}()

// String returns the activation name.
func (a ActivationType) String() string {
	if name, ok := activationNames[a]; ok {
		return name
	}
	return fmt.Sprintf("activation(%d)", a)
}

// MarshalJSON encodes activation type as a string.
func (a ActivationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON decodes activation type from string or integer.
func (a *ActivationType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		key := strings.ToLower(name)
		if val, ok := activationValues[key]; ok {
			*a = val
			return nil
		}
		return fmt.Errorf("unknown activation type %q", name)
	}

	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*a = ActivationType(num)
		return nil
	}
	return fmt.Errorf("invalid activation type")
}

var nodeKindNames = map[NodeKind]string{
	NodeInput:  "input",
	NodeHidden: "hidden",
	NodeOutput: "output",
}

var nodeKindValues = func() map[string]NodeKind {
	out := make(map[string]NodeKind, len(nodeKindNames))
	for k, v := range nodeKindNames {
		out[v] = k
	}
	return out
}()

// String returns the node kind name.
func (k NodeKind) String() string {
	if name, ok := nodeKindNames[k]; ok {
		return name
	}
	return fmt.Sprintf("kind(%d)", k)
}

// MarshalJSON encodes node kind as a string.
func (k NodeKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalJSON decodes node kind from string or integer.
func (k *NodeKind) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		key := strings.ToLower(name)
		if val, ok := nodeKindValues[key]; ok {
			*k = val
			return nil
		}
		return fmt.Errorf("unknown node kind %q", name)
	}

	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*k = NodeKind(num)
		return nil
	}
	return fmt.Errorf("invalid node kind")
}
