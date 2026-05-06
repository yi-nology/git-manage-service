package ai

import "fmt"

func BuildProviderBindingContext(remoteRepos []string, localRepos []string, existingBindings map[string]string, maxChars int) string {
	b := NewContextBuilder()

	b.AddListSection("Local Repositories", localRepos)
	b.AddListSection("Remote Repositories", remoteRepos)

	if len(existingBindings) > 0 {
		bindingsList := make([]string, 0, len(existingBindings))
		for local, remote := range existingBindings {
			bindingsList = append(bindingsList, fmt.Sprintf("%s → %s", local, remote))
		}
		b.AddListSection("Existing Bindings", bindingsList)
	}

	return b.Build()
}

func BuildPatchAnalysisContext(patchContent string, targetBranch string, fileList []string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Target Branch", targetBranch)
	b.AddListSection("Files in Patch", fileList)

	if len(patchContent) > maxChars {
		patchContent = ClampText(patchContent, maxChars)
	}

	b.AddCodeSection("Patch Content", "diff", patchContent)
	return b.Build()
}
