package utils

import (
	"strings"
)

// SplitMessage splits long messages into chunks, preserving code block integrity
// Uses natural boundaries (newlines, spaces) and extends messages slightly to avoid breaking code blocks
func SplitMessage(content string, limit int) []string {
	var messages []string

	for len(content) > 0 {
		if len(content) <= limit {
			messages = append(messages, content)
			break
		}

		msgEnd := limit

		// Find natural split point within the limit
		msgEnd = FindLastNewline(content[:limit], 200)
		if msgEnd <= 0 {
			msgEnd = FindLastSpace(content[:limit], 100)
		}
		if msgEnd <= 0 {
			msgEnd = limit
		}

		// Check if this would end with an incomplete code block
		candidate := content[:msgEnd]
		unclosedIdx := FindLastUnclosedCodeBlock(candidate)

		if unclosedIdx >= 0 {
			// Message would end with incomplete code block
			// Try to extend to include the closing ``` (with some buffer)
			extendedLimit := limit + 500 // Allow 500 char buffer for code blocks
			if len(content) > extendedLimit {
				closingIdx := FindNextClosingCodeBlock(content, msgEnd)
				if closingIdx > 0 && closingIdx <= extendedLimit {
					// Extend to include the closing ```
					msgEnd = closingIdx
				} else {
					// Can't find closing, split before the code block
					msgEnd = FindLastNewline(content[:unclosedIdx], 200)
					if msgEnd <= 0 {
						msgEnd = FindLastSpace(content[:unclosedIdx], 100)
					}
					if msgEnd <= 0 {
						msgEnd = unclosedIdx
					}
				}
			} else {
				// Remaining content fits within extended limit
				msgEnd = len(content)
			}
		}

		if msgEnd <= 0 {
			msgEnd = limit
		}

		messages = append(messages, content[:msgEnd])
		content = strings.TrimSpace(content[msgEnd:])
	}

	return messages
}

// FindLastUnclosedCodeBlock finds the last opening ``` that doesn't have a closing ```
// Returns the position of the opening ``` or -1 if all code blocks are complete
func FindLastUnclosedCodeBlock(text string) int {
	count := 0
	lastOpenIdx := -1

	for i := 0; i < len(text); i++ {
		if i+2 < len(text) && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			if count == 0 {
				lastOpenIdx = i
			}
			count++
			i += 2
		}
	}

	// If odd number of ``` markers, last one is unclosed
	if count%2 == 1 {
		return lastOpenIdx
	}
	return -1
}

// FindNextClosingCodeBlock finds the next closing ``` starting from a position
// Returns the position after the closing ``` or -1 if not found
func FindNextClosingCodeBlock(text string, startIdx int) int {
	for i := startIdx; i < len(text); i++ {
		if i+2 < len(text) && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			return i + 3
		}
	}
	return -1
}

// FindLastNewline finds the last newline character within the last N characters
// Returns the position of the newline or -1 if not found
func FindLastNewline(s string, searchWindow int) int {
	searchStart := len(s) - searchWindow
	if searchStart < 0 {
		searchStart = 0
	}
	for i := len(s) - 1; i >= searchStart; i-- {
		if s[i] == '\n' {
			return i
		}
	}
	return -1
}

// FindLastSpace finds the last space character within the last N characters
// Returns the position of the space or -1 if not found
func FindLastSpace(s string, searchWindow int) int {
	searchStart := len(s) - searchWindow
	if searchStart < 0 {
		searchStart = 0
	}
	for i := len(s) - 1; i >= searchStart; i-- {
		if s[i] == ' ' || s[i] == '\t' {
			return i
		}
	}
	return -1
}
