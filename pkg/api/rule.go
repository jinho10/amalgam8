// Copyright 2017 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package api

// Rule represents an individual rule.
type Rule struct {
	ID          string   `json:"id" yaml:"id"`
	Priority    int      `json:"priority" yaml:"priority"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Destination string   `json:"destination" yaml:"destination"`
	Match       *Match   `json:"match,omitempty" yaml:"match,omitempty"`
	Route       *Route   `json:"route,omitempty" yaml:"route,omitempty"`
	Actions     []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// Source definition.
type Source struct {
	Name string   `json:"name" yaml:"name"`
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Match definition
type Match struct {
	Source  *Source           `json:"source,omitempty" yaml:"source,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// Route definition
type Route struct {
	Backends []Backend `json:"backends" yaml:"backends"`
}

// URI for backends.
type URI struct {
	Path          string `json:"path"`
	Prefix        string `json:"prefix"`
	PrefixRewrite string `json:"prefix_rewrite"`
}

// Backend represents a backend to route to.
type Backend struct {
	Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
	Tags    []string `json:"tags" yaml:"tags"`
	URI     *URI     `json:"uri,omitempty" yaml:"uri,omitempty"`
	Weight  float64  `json:"weight,omitempty" yaml:"weight,omitempty"`
	Timeout float64  `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retries int      `json:"retries,omitempty" yaml:"retries,omitempty"`
}

// Action definition
type Action struct {
	Action      string   `json:"action" yaml:"action"`
	Duration    float64  `json:"duration,omitempty" yaml:"duration,omitempty"`
	Probability float64  `json:"probability,omitempty" yaml:"probability,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	ReturnCode  int      `json:"return_code,omitempty" yaml:"return_code,omitempty"`
	LogKey      string   `json:"log_key,omitempty" yaml:"log_key,omitempty"`
	LogValue    string   `json:"log_value,omitempty" yaml:"log_value,omitempty"`
}

// RulesByService definition
type RulesByService struct {
	Services map[string][]Rule `json:"services" yaml:"services"`
}
