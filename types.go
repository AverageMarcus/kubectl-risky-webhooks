package main

type Webhook struct {
	Name              string `header:"name"`
	Webhook           string `header:"webhook"`
	FailureIgnore     bool
	Replicas          int32 `header:"replicas"`
	PDB               bool  `header:"has PDB"`
	KubeSystemIgnored bool  `header:"kube-system ignored"`
}
