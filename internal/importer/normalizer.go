package importer

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	EncodingUTF8    = "utf-8"
	EncodingGB18030 = "gb18030"
)

func NormalizeMarkdownBytes(input []byte) (string, string, error) {
	if len(input) == 0 {
		return "", "", errors.New("empty input")
	}

	if utf8.Valid(input) {
		return normalizeLineEndings(string(input)), EncodingUTF8, nil
	}

	decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(input)
	if err != nil {
		return "", "", err
	}

	return normalizeLineEndings(string(decoded)), EncodingGB18030, nil
}

func normalizeLineEndings(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	return strings.ReplaceAll(input, "\r", "\n")
}

func DetectLikelyGarbledText(input string) (bool, string) {
	if strings.ContainsRune(input, '\uFFFD') {
		return true, "文本中包含替换字符，疑似发生了解码失败"
	}

	garbledRunes := []rune{'鐨', '鍙', '浠', '鏄', '妯', '绗', '閫', '鎬', '璇', '缁', '锛', '銆'}
	hits := 0
	for _, candidate := range garbledRunes {
		hits += strings.Count(input, string(candidate))
		if hits >= 2 {
			return true, "文本中包含多处常见乱码字形，请先核对归一化结果"
		}
	}

	return false, ""
}
