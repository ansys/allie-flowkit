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
