package utils

import (
	"strings"
)

const defaultCodeBlockBuffer = 500

// SplitMessage splits long messages into chunks, preserving code block integrity.
// The function prefers to split at maxLen - defaultCodeBlockBuffer to leave room for code blocks,
// but may extend up to maxLen when needed to avoid breaking incomplete code blocks.
// Please refer to pkg/channels/discord.go for usage.
func SplitMessage(content string, maxLen int) []string {
	var messages []string
	codeBlockBuffer := defaultCodeBlockBuffer

	for len(content) > 0 {
		if len(content) <= maxLen {
			messages = append(messages, content)
			break
		}

		// Effective split point: maxLen minus buffer, to leave room for code blocks
		effectiveLimit := maxLen - codeBlockBuffer
		if effectiveLimit < maxLen/2 {
			effectiveLimit = maxLen / 2
		}

		// Find natural split point within the effective limit
		msgEnd := FindLastNewline(content[:effectiveLimit], 200)
		if msgEnd <= 0 {
			msgEnd = FindLastSpace(content[:effectiveLimit], 100)
		}
		if msgEnd <= 0 {
			msgEnd = effectiveLimit
		}

		// Check if this would end with an incomplete code block
		candidate := content[:msgEnd]
		unclosedIdx := FindLastUnclosedCodeBlock(candidate)

		if unclosedIdx >= 0 {
			// Message would end with incomplete code block
			// Try to extend up to maxLen to include the closing ```
			if len(content) > msgEnd {
				closingIdx := FindNextClosingCodeBlock(content, msgEnd)
				if closingIdx > 0 && closingIdx <= maxLen {
					// Extend to include the closing ```
					msgEnd = closingIdx
				} else {
					// Can't find closing within maxLen, split before the code block
					msgEnd = FindLastNewline(content[:unclosedIdx], 200)
					if msgEnd <= 0 {
						msgEnd = FindLastSpace(content[:unclosedIdx], 100)
					}
					if msgEnd <= 0 {
						msgEnd = unclosedIdx
					}
				}
			}
		}

		if msgEnd <= 0 {
			msgEnd = effectiveLimit
		}

		messages = append(messages, content[:msgEnd])
		content = strings.TrimSpace(content[msgEnd:])
	}

	return messages
}

// FindLastUnclosedCodeBlock finds the last opening ``` that doesn't have a closing ```
// Returns the position of the opening ``` or -1 if all code blocks are complete
func FindLastUnclosedCodeBlock(text string) int {
	inCodeBlock := false
	lastOpenIdx := -1

	for i := 0; i < len(text); i++ {
		if i+2 < len(text) && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			// Toggle code block state on each fence
			if !inCodeBlock {
				// Entering a code block: record this opening fence
				lastOpenIdx = i
			}
			inCodeBlock = !inCodeBlock
			i += 2
		}
	}

	if inCodeBlock {
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
