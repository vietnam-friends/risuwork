package scenario

import (
	"fmt"
	"net/http"
	"time"

	"github.com/isucon/isucandar/agent"
)

func MustNewAgent(targetHost string, timeout time.Duration, benchID string) *agent.Agent {
	ag, err := NewAgent(targetHost, timeout, benchID)
	if err != nil {
		panic(err)
	}
	return ag
}

func NewAgent(targetHost string, timeout time.Duration, benchID string) (*agent.Agent, error) {
	agentOptions := []agent.AgentOption{
		agent.WithBaseURL(fmt.Sprintf("http://%s/", targetHost)),
		agent.WithTimeout(timeout),
		agent.WithUserAgent("risu-benchmarker"),
		WithRoundTripper(&AddHeaderTransport{RoundTripper: agent.DefaultTransport.Clone(), Key: "X-risu-bench-id", Value: benchID}),
	}
	return agent.NewAgent(agentOptions...)
}

func WithRoundTripper(rt http.RoundTripper) agent.AgentOption {
	return func(a *agent.Agent) error {
		a.HttpClient.Transport = rt
		return nil
	}
}

type AddHeaderTransport struct {
	RoundTripper http.RoundTripper
	Key          string
	Value        string
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(adt.Key, adt.Value)
	return adt.RoundTripper.RoundTrip(req)
}
