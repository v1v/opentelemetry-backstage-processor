package backstageprocessor

// createDemoLabels only for illustrating how it works in the demo environment.
func createDemoLabels(original map[string]RepoInfo) {
	labels := make(map[string]RepoInfo)

	labels["demo-devops-bcn-ingress-nginx"] = RepoInfo{
		Repo:     "ingress-nginx",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-bcn-elasticsearch"] = RepoInfo{
		Repo:     "elasticsearch",
		Org:      "platform",
		Division: "engineering",
	}
	labels["demo-devops-bcn-apm-server"] = RepoInfo{
		Repo:     "apm-server",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-bcn-opentelemetry-lambda"] = RepoInfo{
		Repo:     "opentelemetry-lambda",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-bcn-elastic-otel-node"] = RepoInfo{
		Repo:     "elastic-otel-node",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-bcn-elastic-otel-java"] = RepoInfo{
		Repo:     "elastic-otel-java",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-bcn-setup-go"] = RepoInfo{
		Repo:     "setup-go",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-bcn-demo"] = RepoInfo{
		Repo:     "demo",
		Org:      "platform",
		Division: "engineering",
	}
	labels["demo-devops-bcn-codeql-action"] = RepoInfo{
		Repo:     "codeql-action",
		Org:      "platform",
		Division: "engineering",
	}

	for key, value := range labels {
		original[key] = value
	}
}
