package controller

import (
	"fmt"
	"strings"
)

const (
	maxTagNameLen     int    = 512
	maxTagValLen      int    = 256
	maxNumTags        int    = 50
	invalidTagChars   string = "<>%&\\?/"
	maxLabelNameLen   int    = 63
	maxLabelPrefixLen int    = 253
	maxLabelValLen    int    = 63
)

func ValidTagName(labelName string, configOptions ConfigOptions) bool {
	return validTagName(labelWithoutPrefix(labelName, configOptions.LabelPrefix))
}

func ConvertTagNameToValidLabelName(tagName string, configOptions ConfigOptions) string {
	// lstrip configOptions.TagPrefix if there
	// don't forget to get rid of '.' after 'node.labels'... are there prefixes here?
	result := tagName
	if strings.HasPrefix(tagName, fmt.Sprintf("%s", configOptions.TagPrefix)) {
		result = strings.TrimPrefix(tagName, fmt.Sprintf("%s", configOptions.TagPrefix))
	}

	// truncate name segment to 63 characters or less
	if len(result) > maxLabelNameLen {
		result = result[:maxLabelNameLen+1]
	}

	// must begin and end with alphanumeric character with -,_,. and alphanumerics in between

	// must have prefix
	// prefix must not be longer than 253 characters
	result = labelWithPrefix(result, configOptions.LabelPrefix)
	return result
}

func ConvertLabelNameToValidTagName(labelName string, configOptions ConfigOptions) string {
	// get rid of '/' and other characters.
	// also detect if 'azure.tags' is in the name to get rid of it? also get rid of '/' after 'azure.tags'
	// don't add if label name is a truncated version of a tag
	result := labelName
	if strings.HasPrefix(labelName, fmt.Sprintf("%s/", configOptions.LabelPrefix)) {
		result = strings.TrimPrefix(labelName, fmt.Sprintf("%s/", configOptions.LabelPrefix))
	}

	if validTagName(result) {
		// what now?
	}

	// result = tagWithPrefix(result, configOptions.TagPrefix)
	return result
}

func ConvertTagValToValidLabelVal(tagVal string) string {
	result := tagVal
	if len(result) > maxLabelValLen {
		result = result[:maxLabelValLen+1]
	}
	return result
}

func ConvertLabelValToValidTagVal() {
}

func labelWithPrefix(labelName, prefix string) string {
	return fmt.Sprintf("%s/%s", prefix, labelName)
}

func labelWithoutPrefix(labelName, prefix string) string {
	if strings.HasPrefix(labelName, fmt.Sprintf("%s/", prefix)) {
		return strings.TrimPrefix(labelName, fmt.Sprintf("%s/", prefix))
	}
	return labelName
}

// what character should I use?
func tagWithPrefix(tagName, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, tagName)
}

func tagWithoutPrefix(tagName, prefix string) string {
	return ""
}

func validTagName(labelName string) bool {
	if strings.ContainsAny(labelName, invalidTagChars) {
		return false
	}
	return true
}
