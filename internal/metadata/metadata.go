package metadata

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadMetadata() Metadata {
	data, err := os.ReadFile(".stick/metadata.json")
	if err != nil {
		return Metadata{VirtualBranches: make(map[string]VirtualBranch)}
	}
	var metadata Metadata
	json.Unmarshal(data, &metadata)
	return metadata
}

func SaveMetadata(metadata Metadata) {
	data, err := json.MarshalIndent(metadata, "", "	")
	if err != nil {
		fmt.Println("failed to marshal metadata: ", err)
		return
	}
	err = os.WriteFile(".stick/metadata.json", data, 0644)
	if err != nil {
		fmt.Println("failed to write metadata: ", err)
	}
}
