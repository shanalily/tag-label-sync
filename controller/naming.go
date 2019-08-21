package controller

import (
	"fmt"
	"strings"
)

func labelWithPrefix(labelName, prefix string) string {
	return fmt.Sprintf("%s/%s", prefix, labelName)
}

// what character should I use?
func tagWithPrefix(tagName, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, tagName)
}

func convertTagNameToValidLabelName(tagName string, configOptions ConfigOptions) string {
	// lstrip configOptions.TagPrefix if there
	result := tagName
	if strings.HasPrefix(tagName, configOptions.TagPrefix) {
		result = strings.TrimPrefix(tagName, configOptions.TagPrefix)
	}

	// truncate name segment to 63 characters or less
	if len(result) > 63 {
		result = result[:64]
	}

	// must begin and end with alphanumeric character with -,_,. and alphanumerics in between

	// must have prefix
	// prefix must not be longer than 253 characters
	result = labelWithPrefix(result, configOptions.LabelPrefix)
	return result
}

func convertLabelNameToValidTagName(labelName string, configOptions ConfigOptions) string {
	// get rid of '/' and other characters.
	// also detect if 'azure.tags' is in the name to get rid of it?
	// don't add if label name is a truncated version of a tag
	result := labelName
	if strings.HasPrefix(labelName, configOptions.LabelPrefix) {
		result = strings.TrimPrefix(labelName, configOptions.LabelPrefix)
	}
	result = tagWithPrefix(result, configOptions.TagPrefix)
	return result
}

func convertTagValToValidLabelVal() {
}

func convertLabelValToValidTagVal() {
}
