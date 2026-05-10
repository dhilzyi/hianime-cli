package reanime

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SvelteKit devalue wire format:
//
// The response is a JSON object:
//
//	{
//	  "type": "data",
//	  "nodes": [ <node>, <node>, ... ]
//	}
//
// Each node looks like:
//
//	{
//	  "type": "data",
//	  "data": [ <root>, <val1>, <val2>, ... ],
//	  "uses": {}
//	}
//
// data[0] is the root descriptor. It can be:
//   - A map[string]interface{} where every value is an int index into data[]
//   - A []interface{} where every element is an int index into data[]
//   - A scalar (string, number, bool, nil) — returned as-is
//
// Resolution is recursive: after resolving an index you may get another
// map/array of indices, so keep resolving until you hit a scalar.

// devalueResponse is the top-level shape.
type devalueResponse struct {
	Type  string        `json:"type"`
	Nodes []devalueNode `json:"nodes"`
}

// devalueNode is one chunk inside nodes[].
type devalueNode struct {
	Type string        `json:"type"`
	Data []interface{} `json:"data"`
	Uses interface{}   `json:"uses"`
}

// resolve recursively resolves a value from the node's data array.
func resolve(data []interface{}, idx int) interface{} {
	if idx < 0 || idx >= len(data) {
		return nil
	}

	node := data[idx]

	switch v := node.(type) {

	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, raw := range v {
			valIdx, ok := toInt(raw)
			if !ok {
				result[key] = raw
				continue
			}
			result[key] = resolve(data, valIdx)
		}
		return result

	case []interface{}:
		result := make([]interface{}, 0, len(v))
		for _, raw := range v {
			itemIdx, ok := toInt(raw)
			if !ok {
				result = append(result, raw)
				continue
			}
			result = append(result, resolve(data, itemIdx))
		}
		return result

	default:
		// scalar: string, float64, bool, nil
		return v
	}
}

// decodeNode decodes a single devalue node into a Go value.
func decodeNode(node devalueNode) interface{} {
	if len(node.Data) == 0 {
		return nil
	}
	return resolve(node.Data, 0)
}

// DecodeDevalue parses a raw SvelteKit devalue JSON payload and returns
// a slice of decoded values, one per node.
func DecodeDevalue(raw []byte) ([]interface{}, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()

	var resp devalueResponse
	if err := dec.Decode(&resp); err != nil {
		// Maybe it's just a bare nodes array — try that
		var nodes []devalueNode
		dec2 := json.NewDecoder(strings.NewReader(string(raw)))
		dec2.UseNumber()
		if err2 := dec2.Decode(&nodes); err2 != nil {
			return nil, fmt.Errorf("failed to parse devalue JSON: %w", err)
		}
		resp.Nodes = nodes
	}

	// Normalize json.Number → float64 throughout all nodes
	for i := range resp.Nodes {
		resp.Nodes[i].Data = normalizeNumbers(resp.Nodes[i].Data).([]interface{})
	}

	results := make([]interface{}, 0, len(resp.Nodes))
	for _, node := range resp.Nodes {
		if node.Type != "data" {
			continue
		}
		results = append(results, decodeNode(node))
	}

	// If there's only one meaningful node, unwrap it
	if len(results) == 1 {
		return results, nil
	}
	return results, nil
}

// toInt safely converts a JSON number to int.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		i := int(n)
		if float64(i) == n {
			return i, true
		}
	case int:
		return n, true
	case json.Number:
		i, err := n.Int64()
		if err == nil {
			return int(i), true
		}
	}
	return 0, false
}

// normalizeNumbers recursively converts json.Number → float64.
func normalizeNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case json.Number:
		f, err := val.Float64()
		if err != nil {
			return val.String()
		}
		return f
	case []interface{}:
		for i, item := range val {
			val[i] = normalizeNumbers(item)
		}
		return val
	case map[string]interface{}:
		for k, item := range val {
			val[k] = normalizeNumbers(item)
		}
		return val
	}
	return v
}
