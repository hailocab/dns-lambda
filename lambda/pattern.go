package lambda

import (
	"bytes"
	"text/template"
)

var (
	// IPRecordPattern holds the pattern format for an IP record
	IPRecordPattern Pattern = "ip-{{.IP}}.{{.Region}}.i.{{.Domain}}"
)

// Pattern is a holder for a DNS hostname record format
type Pattern string

// Parse returns a completed template
func (p Pattern) Parse(data interface{}) (string, error) {
	tmpl, err := template.New("pattern").Parse(string(p))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
