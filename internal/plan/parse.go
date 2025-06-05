package plan

import (
	"encoding/json"
	"fmt"

	tfJson "github.com/hashicorp/terraform-json"
)

type Plan tfJson.Plan

/*
Takes a simple JSON Terraform plan and parses it into a struct.
Errors if there is a problem parsing the json.
*/
func ParsePlan(data []byte) (*Plan, error) {

	p := &Plan{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("unable to unmarshal terraform plan caused by: %v", err)
	}
	return p, nil
}

/*
Parses Json byte data into an InspectFilter
*/
func ParseInspectFilter(data []byte) (*InspectFilter, error) {

	i := &InspectFilter{}
	if err := json.Unmarshal(data, i); err != nil {
		return nil, fmt.Errorf("unable to unmarshal inspect filter caused by: %v", err)
	}
	return i, nil
}
