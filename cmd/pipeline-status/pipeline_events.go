package main

import "time"

type CodePipelineEvent struct {
	Account             string               `json:"account"`
	DetailType           string               `json:"detailType"`
	Region               string               `json:"region"`
	Source               string               `json:"source"`
	Time                 time.Time            `json:"time"`
	NotificationRuleArn  string               `json:"notificationRuleArn"`
	Detail               Detail               `json:"detail"`
	Resources            []string             `json:"resources"`
	AdditionalAttributes AdditionalAttributes `json:"additionalAttributes"`
}
type ExecutionResult struct {
	ExternalExecutionURL string `json:"external-execution-url"`
	ExternalExecutionID  string `json:"external-execution-id"`
}
type S3Location struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}
type OutputArtifacts struct {
	Name       string     `json:"name"`
	S3Location S3Location `json:"s3location"`
}
type Type struct {
	Owner    string `json:"owner"`
	Provider string `json:"provider"`
	Category string `json:"category"`
	Version  string `json:"version"`
}
type Detail struct {
	Pipeline        string            `json:"pipeline"`
	ExecutionID     string            `json:"execution-id"`
	Stage           string            `json:"stage"`
	ExecutionResult ExecutionResult   `json:"execution-result"`
	OutputArtifacts []OutputArtifacts `json:"output-artifacts"`
	Action          string            `json:"action"`
	State           string            `json:"state"`
	Region          string            `json:"region"`
	Type            Type              `json:"type"`
	Version         float64           `json:"version"`
}
type SourceActionVariables struct {
	BranchName     string `json:"BranchName"`
	CommitID       string `json:"CommitId"`
	RepositoryName string `json:"RepositoryName"`
}
type SourceActions struct {
	SourceActionName      string                `json:"sourceActionName"`
	SourceActionProvider  string                `json:"sourceActionProvider"`
	SourceActionVariables SourceActionVariables `json:"sourceActionVariables"`
}
type AdditionalAttributes struct {
	SourceActions []SourceActions `json:"sourceActions"`
}