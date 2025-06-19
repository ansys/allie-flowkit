// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package externalfunctions

import (
	"encoding/json"
	"fmt"
)

// SetGenerateRequestJsonBody creates a JSON body for the generate request to RHSC Copilot.
// It takes various parameters to configure the request and returns the JSON string.
//
// Tags:
//   - @displayName: Copilot Generate Request JSON Body
//
// Parameters:
//   - query: the query string for the request.
//   - sessionID: the session ID for the request.
//   - mode: the mode of operation for the request.
//   - timeout: the timeout for the request in seconds.
//   - priority: the priority of the request.
//   - agentPreference: the preferred agent for the request.
//   - saveIntermediate: whether to save intermediate results.
//   - similarityTopK: the number of top similar results to consider.
//   - noCritique: whether to disable critique.
//   - maxIterations: the maximum number of iterations for the request.
//   - forceAzure: whether to force the use of Azure for the request.
func SetCopilotGenerateRequestJsonBody(
	query string,
	sessionID string,
	mode string,
	timeout int,
	priority int,
	agentPreference string,
	saveIntermediate bool,
	similarityTopK int,
	noCritique bool,
	maxIterations int,
	forceAzure bool,
) (jsonBody string) {

	req := copilotGenerateRequest{
		Query:     query,
		SessionID: sessionID,
		Mode:      mode,
		Timeout:   timeout,
		Priority:  priority,
		Options: copilotGenerateOptions{
			AgentPreference:  agentPreference,
			SaveIntermediate: saveIntermediate,
			SimilarityTopK:   similarityTopK,
			NoCritique:       noCritique,
			MaxIterations:    maxIterations,
			ForceAzure:       forceAzure,
		},
	}

	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		panic("Error marshalling request to JSON: " + err.Error())
	}

	jsonBody = string(bytes)
	fmt.Printf("Generated JSON body for Copilot Generate Request: %s\n", jsonBody)
	return jsonBody
}
